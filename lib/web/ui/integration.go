// Copyright 2023 Gravitational, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ui

import (
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
)

// Integration describes an Integration
type Integration struct {
	// Name is the Integration name.
	Name string `json:"name"`
	// SubKind is the Integration SubKind.
	SubKind string `json:"subKind"`
	// AWSRoleARN is the role associated with the integration when SubKind is `aws-oidc`
	AWSRoleARN string `json:"awsRoleARN,omitempty"`
}

// CreateIntegrationRequest is a request to create an Integration
type CreateIntegrationRequest struct {
	Integration
}

// CheckAndSetDefaults for the create request.
// Name and SubKind is required.
func (r *CreateIntegrationRequest) CheckAndSetDefaults() error {
	if r.Name == "" {
		return trace.BadParameter("missing integration name")
	}

	if r.SubKind == "" {
		return trace.BadParameter("missing subKind")
	}

	return nil
}

// UpdateIntegrationRequest is a request to update an Integration
type UpdateIntegrationRequest struct {
	// AWSRoleARN is the role associated with the integration when SubKind is `aws-oidc`
	AWSRoleARN string `json:"awsRoleARN,omitempty"`
}

// CheckAndSetDefaults checks if the provided values are valid.
func (r *UpdateIntegrationRequest) CheckAndSetDefaults() error {
	if r.AWSRoleARN == "" {
		return trace.BadParameter("missing awsRoleARN field")
	}
	return nil
}

// IntegrationsListResponse contains a list of Integrations.
// In case of exceeding the pagination limit (either via query param `limit` or the default 1000)
// a `nextToken` is provided and should be used to obtain the next page (as a query param `startKey`)
type IntegrationsListResponse struct {
	// Items is a list of resources retrieved.
	Items interface{} `json:"items"`
	// NextKey is the position to resume listing events.
	NextKey string `json:"nextKey"`
}

// MakeIntegrations creates a UI list of Integrations.
func MakeIntegrations(igs []types.Integration) []Integration {
	uiList := make([]Integration, 0, len(igs))

	for _, ig := range igs {
		uiList = append(uiList, MakeIntegration(ig))
	}

	return uiList
}

// MakeIntegration creates a UI Integration representation.
func MakeIntegration(ig types.Integration) Integration {
	return Integration{
		Name:       ig.GetName(),
		SubKind:    ig.GetSubKind(),
		AWSRoleARN: ig.GetAWSRoleARN(),
	}
}
