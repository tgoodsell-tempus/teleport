/*
Copyright 2023 Gravitational, Inc.
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

package types

import (
	"fmt"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/utils"
)

const (
	// IntegrationSubKindAWSOIDC is an integration with AWS that uses OpenID Connect as an Identity Provider.
	// More information can be found here: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html
	// This Integration requires the AWSRoleARN Spec field to be present.
	// That is the AWS role to be used when creating a token, to then issue an API Call to AWS.
	IntegrationSubKindAWSOIDC = "aws-oidc"
)

// Integration specifies is a connection configuration between Teleport and a 3rd party system.
type Integration interface {
	ResourceWithLabels

	// SubKind `aws-oidc` fields:
	// GetAWSRoleARN returns the AWS Role ARN that must be used to obtain a token for AWS API calls.
	GetAWSRoleARN() string
	// SetAWSRoleARN sets the AWS Role ARN.
	SetAWSRoleARN(string)
}

var _ ResourceWithLabels = (*IntegrationV1)(nil)

// NewIntegration returns a new Integration.
func NewIntegration(md Metadata, subKind string, spec IntegrationSpecV1) (Integration, error) {
	ig := &IntegrationV1{
		ResourceHeader: ResourceHeader{
			Metadata: md,
			Kind:     KindIntegration,
			Version:  V1,
			SubKind:  subKind,
		},
		Spec: spec,
	}
	if err := ig.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	return ig, nil
}

// String returns the integration string representation.
func (ig *IntegrationV1) String() string {
	return fmt.Sprintf("IntegrationV1(Name=%v, SubKind=%s, Labels=%v)",
		ig.GetName(), ig.GetSubKind(), ig.GetAllLabels())
}

// MatchSearch goes through select field values and tries to
// match against the list of search values.
func (ig *IntegrationV1) MatchSearch(values []string) bool {
	fieldVals := append(utils.MapToStrings(ig.GetAllLabels()), ig.GetName())
	return MatchSearch(fieldVals, values, nil)
}

// setStaticFields sets static resource header and metadata fields.
func (ig *IntegrationV1) setStaticFields() {
	ig.Kind = KindIntegration
	ig.Version = V1
}

// CheckAndSetDefaults checks and sets default values
func (ig *IntegrationV1) CheckAndSetDefaults() error {
	ig.setStaticFields()
	if err := ig.ResourceHeader.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	switch ig.ResourceHeader.SubKind {
	case IntegrationSubKindAWSOIDC:
		if ig.Spec.AWSRoleARN == "" {
			return trace.BadParameter("awsRoleARN is required for %q SubKind", ig.ResourceHeader.SubKind)
		}
	default:
		return trace.BadParameter("invalid SubKind")
	}

	return nil
}

// SubKind `aws-oidc` fields
// GetAWSRoleARN returns the AWS Role ARN that must be used to obtain a token for AWS API calls.
func (ig *IntegrationV1) GetAWSRoleARN() string {
	return ig.Spec.AWSRoleARN
}

// SetAWSRoleARN sets the AWS Role ARN.
func (ig *IntegrationV1) SetAWSRoleARN(awsRoleARN string) {
	ig.Spec.AWSRoleARN = awsRoleARN
}

// Integrations is a list of Integration resources.
type Integrations []Integration

// AsResources returns these groups as resources with labels.
func (igs Integrations) AsResources() []ResourceWithLabels {
	resources := make([]ResourceWithLabels, len(igs))
	for i, ig := range igs {
		resources[i] = ig
	}
	return resources
}

// Len returns the slice length.
func (igs Integrations) Len() int { return len(igs) }

// Less compares integrations by name.
func (igs Integrations) Less(i, j int) bool { return igs[i].GetName() < igs[j].GetName() }

// Swap swaps two integrations.
func (igs Integrations) Swap(i, j int) { igs[i], igs[j] = igs[j], igs[i] }
