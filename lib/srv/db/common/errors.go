/*
Copyright 2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/gravitational/trace"
	"github.com/gravitational/trace/trail"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/status"

	awslib "github.com/gravitational/teleport/lib/cloud/aws"
	azurelib "github.com/gravitational/teleport/lib/cloud/azure"
	"github.com/gravitational/teleport/lib/defaults"
)

// ConvertError converts errors to trace errors.
func ConvertError(err error) error {
	if err == nil {
		return nil
	}
	// Unwrap original error first.
	if _, ok := err.(*trace.TraceErr); ok {
		return ConvertError(trace.Unwrap(err))
	}
	if pgErr, ok := err.(pgError); ok {
		return ConvertError(pgErr.Unwrap())
	}
	if causer, ok := err.(causer); ok {
		return ConvertError(causer.Cause())
	}
	if _, ok := status.FromError(err); ok {
		return trail.FromGRPC(err)
	}
	switch e := trace.Unwrap(err).(type) {
	case *googleapi.Error:
		return convertGCPError(e)
	case awserr.RequestFailure:
		return awslib.ConvertRequestFailureError(e)
	case *azcore.ResponseError:
		return azurelib.ConvertResponseError(e)
	case *pgconn.PgError:
		return convertPostgresError(e)
	case *mysql.MyError:
		return convertMySQLError(e)
	}
	return err // Return unmodified.
}

// convertGCPError converts GCP errors to trace errors.
func convertGCPError(err *googleapi.Error) error {
	switch err.Code {
	case http.StatusForbidden:
		return trace.AccessDenied(err.Error())
	case http.StatusConflict:
		return trace.CompareFailed(err.Error())
	}
	return err // Return unmodified.
}

// convertPostgresError converts Postgres driver errors to trace errors.
func convertPostgresError(err *pgconn.PgError) error {
	switch err.Code {
	case pgerrcode.InvalidAuthorizationSpecification, pgerrcode.InvalidPassword:
		return trace.AccessDenied(err.Error())
	}
	return err // Return unmodified.
}

// convertMySQLError converts MySQL driver errors to trace errors.
func convertMySQLError(err *mysql.MyError) error {
	switch err.Code {
	case mysql.ER_ACCESS_DENIED_ERROR:
		return trace.AccessDenied(err.Error())
	}
	return err // Return unmodified.
}

// causer defines an interface for errors wrapped by the "errors" package.
type causer interface {
	Cause() error
}

// pgError defines an interface for errors wrapped by Postgres driver.
type pgError interface {
	Unwrap() error
}

// ConvertConnectError converts common connection errors to trace errors with
// extra information/recommendations if necessary.
func ConvertConnectError(err error, sessionCtx *Session) error {
	errString := err.Error()
	switch {
	case strings.Contains(errString, "x509: certificate signed by unknown authority"):
		return trace.AccessDenied("Database service cannot validate database's certificate: %v. Please verify if the correct CA bundle is used in the database config.", err)

	case strings.Contains(errString, "tls: unknown certificate authority"):
		return trace.AccessDenied("Database cannot validate client certificate generated by database service: %v.", err)
	}

	err = ConvertError(err)

	if trace.IsAccessDenied(err) {
		switch {
		case sessionCtx.Database.IsRDS():
			return createRDSAccessDeniedError(err, sessionCtx)
		case sessionCtx.Database.IsAzure():
			return createAzureAccessDeniedError(err, sessionCtx)
		}
	}

	return trace.Wrap(err)
}

// createRDSAccessDeniedError creates an error with help message to setup IAM
// auth for RDS.
func createRDSAccessDeniedError(err error, sessionCtx *Session) error {
	policy, getPolicyErr := sessionCtx.Database.GetIAMPolicy()
	if getPolicyErr != nil {
		policy = fmt.Sprintf("failed to generate IAM policy: %v", getPolicyErr)
	}

	switch sessionCtx.Database.GetProtocol() {
	case defaults.ProtocolMySQL:
		return trace.AccessDenied(`Could not connect to database:

  %v

Make sure that IAM auth is enabled for MySQL user %q and Teleport database
agent's IAM policy has "rds-connect" permissions (note that IAM changes may
take a few minutes to propagate):

%v
`, err, sessionCtx.DatabaseUser, policy)

	case defaults.ProtocolPostgres:
		return trace.AccessDenied(`Could not connect to database:

  %v

Make sure that Postgres user %q has "rds_iam" role and Teleport database
agent's IAM policy has "rds-connect" permissions (note that IAM changes may
take a few minutes to propagate):

%v
`, err, sessionCtx.DatabaseUser, policy)

	default:
		return trace.Wrap(err)
	}
}

// createAzureAccessDeniedError creates an error with help message to setup AAD
// auth for PostgreSQL/MySQL.
func createAzureAccessDeniedError(err error, sessionCtx *Session) error {
	switch sessionCtx.Database.GetProtocol() {
	case defaults.ProtocolMySQL:
		return trace.AccessDenied(`Could not connect to database:

  %v

Make sure that Azure Active Directory auth is configured for MySQL user %q and the Teleport database
agent's service principal. See: https://goteleport.com/docs/database-access/guides/azure-postgres-mysql/
`, err, sessionCtx.DatabaseUser)
	case defaults.ProtocolPostgres:
		return trace.AccessDenied(`Could not connect to database:

  %v

Make sure that Azure Active Directory auth is configured for Postgres user %q and the Teleport database
agent's service principal. See: https://goteleport.com/docs/database-access/guides/azure-postgres-mysql/
`, err, sessionCtx.DatabaseUser)
	default:
		return trace.Wrap(err)
	}
}
