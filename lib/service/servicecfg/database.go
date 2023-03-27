// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicecfg

import (
	"strings"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/azure"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/limiter"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/srv/db/common/enterprise"
)

// DatabasesConfig configures the database proxy service.
type DatabasesConfig struct {
	// Enabled enables the database proxy service.
	Enabled bool
	// Databases is a list of databases proxied by this service.
	Databases []Database
	// ResourceMatchers match cluster database resources.
	ResourceMatchers []services.ResourceMatcher
	// AWSMatchers match AWS hosted databases.
	AWSMatchers []services.AWSMatcher
	// AzureMatchers match Azure hosted databases.
	AzureMatchers []services.AzureMatcher
	// Limiter limits the connection and request rates.
	Limiter limiter.Config
}

// Database represents a single database that's being proxied.
type Database struct {
	// Name is the database name, used to refer to in CLI.
	Name string
	// Description is a free-form database description.
	Description string
	// Protocol is the database type, e.g. postgres or mysql.
	Protocol string
	// URI is the database endpoint to connect to.
	URI string
	// StaticLabels is a map of database static labels.
	StaticLabels map[string]string
	// MySQL are additional MySQL database options.
	MySQL MySQLOptions
	// DynamicLabels is a list of database dynamic labels.
	DynamicLabels services.CommandLabels
	// TLS keeps database connection TLS configuration.
	TLS DatabaseTLS
	// AWS contains AWS specific settings for RDS/Aurora/Redshift databases.
	AWS DatabaseAWS
	// GCP contains GCP specific settings for Cloud SQL databases.
	GCP DatabaseGCP
	// AD contains Active Directory configuration for database.
	AD DatabaseAD
	// Azure contains Azure database configuration.
	Azure DatabaseAzure
}

// CheckAndSetDefaults validates the database proxy configuration.
func (d *Database) CheckAndSetDefaults() error {
	if err := enterprise.ProtocolValidation(d.Protocol); err != nil {
		return trace.Wrap(err)
	}
	if d.Name == "" {
		return trace.BadParameter("empty database name")
	}

	// Mark the database as coming from the static configuration.
	if d.StaticLabels == nil {
		d.StaticLabels = make(map[string]string)
	}
	d.StaticLabels[types.OriginLabel] = types.OriginConfigFile

	if err := d.TLS.Mode.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	// We support Azure AD authentication and Kerberos auth with AD for SQL
	// Server. The first method doesn't require additional configuration since
	// it assumes the environment’s Azure credentials
	// (https://learn.microsoft.com/en-us/azure/developer/go/azure-sdk-authentication).
	// The second method requires additional information, validated by
	// DatabaseAD.
	if d.Protocol == defaults.ProtocolSQLServer &&
		(d.AD.Domain != "" || !strings.Contains(d.URI, azure.MSSQLEndpointSuffix)) {
		if err := d.AD.CheckAndSetDefaults(d.Name); err != nil {
			return trace.Wrap(err)
		}
	}

	// Do a test run with extra validations.
	db, err := d.ToDatabase()
	if err != nil {
		return trace.Wrap(err)
	}
	return trace.Wrap(services.ValidateDatabase(db))
}

// ToDatabase converts Database to types.Database.
func (d *Database) ToDatabase() (types.Database, error) {
	return types.NewDatabaseV3(types.Metadata{
		Name:        d.Name,
		Description: d.Description,
		Labels:      d.StaticLabels,
	}, types.DatabaseSpecV3{
		Protocol: d.Protocol,
		URI:      d.URI,
		CACert:   string(d.TLS.CACert),
		TLS: types.DatabaseTLS{
			CACert:     string(d.TLS.CACert),
			ServerName: d.TLS.ServerName,
			Mode:       d.TLS.Mode.ToProto(),
		},
		MySQL: types.MySQLOptions{
			ServerVersion: d.MySQL.ServerVersion,
		},
		AWS: types.AWS{
			AccountID:  d.AWS.AccountID,
			ExternalID: d.AWS.ExternalID,
			Region:     d.AWS.Region,
			Redshift: types.Redshift{
				ClusterID: d.AWS.Redshift.ClusterID,
			},
			RedshiftServerless: types.RedshiftServerless{
				WorkgroupName: d.AWS.RedshiftServerless.WorkgroupName,
				EndpointName:  d.AWS.RedshiftServerless.EndpointName,
			},
			RDS: types.RDS{
				InstanceID: d.AWS.RDS.InstanceID,
				ClusterID:  d.AWS.RDS.ClusterID,
			},
			ElastiCache: types.ElastiCache{
				ReplicationGroupID: d.AWS.ElastiCache.ReplicationGroupID,
			},
			MemoryDB: types.MemoryDB{
				ClusterName: d.AWS.MemoryDB.ClusterName,
			},
			SecretStore: types.SecretStore{
				KeyPrefix: d.AWS.SecretStore.KeyPrefix,
				KMSKeyID:  d.AWS.SecretStore.KMSKeyID,
			},
		},
		GCP: types.GCPCloudSQL{
			ProjectID:  d.GCP.ProjectID,
			InstanceID: d.GCP.InstanceID,
		},
		DynamicLabels: types.LabelsToV2(d.DynamicLabels),
		AD: types.AD{
			KeytabFile:  d.AD.KeytabFile,
			Krb5File:    d.AD.Krb5File,
			Domain:      d.AD.Domain,
			SPN:         d.AD.SPN,
			LDAPCert:    d.AD.LDAPCert,
			KDCHostName: d.AD.KDCHostName,
		},
		Azure: types.Azure{
			ResourceID:    d.Azure.ResourceID,
			IsFlexiServer: d.Azure.IsFlexiServer,
		},
	})
}

// MySQLOptions are additional MySQL options.
type MySQLOptions struct {
	// ServerVersion is the version reported by Teleport DB Proxy on initial handshake.
	ServerVersion string
}

// DatabaseTLS keeps TLS settings used when connecting to database.
type DatabaseTLS struct {
	// Mode is the TLS connection mode. See TLSMode for more details.
	Mode TLSMode
	// ServerName allows providing custom server name.
	// This name will override DNS name when validating certificate presented by the database.
	ServerName string
	// CACert is an optional database CA certificate.
	CACert []byte
}

// DatabaseAWS contains AWS specific settings for RDS/Aurora databases.
type DatabaseAWS struct {
	// Region is the cloud region database is running in when using AWS RDS.
	Region string
	// Redshift contains Redshift specific settings.
	Redshift DatabaseAWSRedshift
	// RDS contains RDS specific settings.
	RDS DatabaseAWSRDS
	// ElastiCache contains ElastiCache specific settings.
	ElastiCache DatabaseAWSElastiCache
	// MemoryDB contains MemoryDB specific settings.
	MemoryDB DatabaseAWSMemoryDB
	// SecretStore contains settings for managing secrets.
	SecretStore DatabaseAWSSecretStore
	// AccountID is the AWS account ID.
	AccountID string
	// ExternalID is an optional AWS external ID used to enable assuming an AWS role across accounts.
	ExternalID string
	// RedshiftServerless contains AWS Redshift Serverless specific settings.
	RedshiftServerless DatabaseAWSRedshiftServerless
}

// DatabaseAWSRedshift contains AWS Redshift specific settings.
type DatabaseAWSRedshift struct {
	// ClusterID is the Redshift cluster identifier.
	ClusterID string
}

// DatabaseAWSRedshiftServerless contains AWS Redshift Serverless specific settings.
type DatabaseAWSRedshiftServerless struct {
	// WorkgroupName is the Redshift Serverless workgroup name.
	WorkgroupName string
	// EndpointName is the Redshift Serverless VPC endpoint name.
	EndpointName string
}

// DatabaseAWSRDS contains AWS RDS specific settings.
type DatabaseAWSRDS struct {
	// InstanceID is the RDS instance identifier.
	InstanceID string
	// ClusterID is the RDS cluster (Aurora) identifier.
	ClusterID string
}

// DatabaseAWSElastiCache contains settings for ElastiCache databases.
type DatabaseAWSElastiCache struct {
	// ReplicationGroupID is the ElastiCache replication group ID.
	ReplicationGroupID string
}

// DatabaseAWSMemoryDB contains settings for MemoryDB databases.
type DatabaseAWSMemoryDB struct {
	// ClusterName is the MemoryDB cluster name.
	ClusterName string
}

// DatabaseAWSSecretStore contains secret store configurations.
type DatabaseAWSSecretStore struct {
	// KeyPrefix specifies the secret key prefix.
	KeyPrefix string
	// KMSKeyID specifies the AWS KMS key for encryption.
	KMSKeyID string
}

// DatabaseGCP contains GCP specific settings for Cloud SQL databases.
type DatabaseGCP struct {
	// ProjectID is the GCP project ID where the database is deployed.
	ProjectID string
	// InstanceID is the Cloud SQL instance ID.
	InstanceID string
}

// DatabaseAD contains database Active Directory configuration.
type DatabaseAD struct {
	// KeytabFile is the path to the Kerberos keytab file.
	KeytabFile string
	// Krb5File is the path to the Kerberos configuration file. Defaults to /etc/krb5.conf.
	Krb5File string
	// Domain is the Active Directory domain the database resides in.
	Domain string
	// SPN is the service principal name for the database.
	SPN string
	// LDAPCert is the Active Directory LDAP Certificate.
	LDAPCert string
	// KDCHostName is the Key Distribution Center Hostname for x509 authentication
	KDCHostName string
}

// IsEmpty returns true if the database AD configuration is empty.
func (d *DatabaseAD) IsEmpty() bool {
	return d.KeytabFile == "" && d.Krb5File == "" && d.Domain == "" && d.SPN == ""
}

// CheckAndSetDefaults validates database Active Directory configuration.
func (d *DatabaseAD) CheckAndSetDefaults(name string) error {
	if d.KeytabFile == "" && d.KDCHostName == "" {
		return trace.BadParameter("missing keytab file path or kdc_host_name for database %q", name)
	}
	if d.Krb5File == "" {
		d.Krb5File = defaults.Krb5FilePath
	}
	if d.Domain == "" {
		return trace.BadParameter("missing Active Directory domain for database %q", name)
	}
	if d.SPN == "" {
		return trace.BadParameter("missing service principal name for database %q", name)
	}

	if d.KDCHostName != "" {
		if d.LDAPCert == "" {
			return trace.BadParameter("missing LDAP certificate for x509 authentication for database %q", name)
		}
	}

	return nil
}

// DatabaseAzure contains Azure database configuration.
type DatabaseAzure struct {
	// ResourceID is the Azure fully qualified ID for the resource.
	ResourceID string
	// IsFlexiServer is true if the database is an Azure Flexible server.
	IsFlexiServer bool
}
