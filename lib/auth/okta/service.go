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

package okta

import (
	"context"

	"github.com/gravitational/trace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"

	oktapb "github.com/gravitational/teleport/api/gen/proto/go/teleport/okta/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/backend"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local"
)

// ServiceConfig is the service config for the Okta gRPC service.
type ServiceConfig struct {
	// Backend is the backend to use.
	Backend backend.Backend

	// Logger is the logger to use.
	Logger logrus.FieldLogger

	// Authorizer is the authorizer to use.
	Authorizer authz.Authorizer

	// OktaImportRules is the Okta import rules service to use.
	OktaImportRules services.OktaImportRules

	// OktaAssignments is the Okta assignments service to use.
	OktaAssignments services.OktaAssignments
}

func (c *ServiceConfig) CheckAndSetDefaults() error {
	if c.Backend == nil {
		return trace.BadParameter("backend is missing")
	}

	if c.Logger == nil {
		c.Logger = logrus.New().WithField(trace.Component, "okta_crud_service")
	}

	if c.Authorizer == nil {
		return trace.BadParameter("authorizer is missing")
	}

	var err error
	var oktaSvc *local.OktaService
	if c.OktaImportRules == nil || c.OktaAssignments == nil {
		oktaSvc, err = local.NewOktaService(c.Backend)
		if err != nil {
			return trace.Wrap(err)
		}
	}

	if c.OktaImportRules == nil {
		c.OktaImportRules = oktaSvc
	}

	if c.OktaAssignments == nil {
		c.OktaAssignments = oktaSvc
	}

	return nil
}

var _ oktapb.OktaServiceServer = (*Service)(nil)

type Service struct {
	oktapb.UnimplementedOktaServiceServer

	log             logrus.FieldLogger
	authorizer      authz.Authorizer
	oktaImportRules services.OktaImportRules
	oktaAssignments services.OktaAssignments
}

// NewService creates a new Okta gRPC service.
func NewService(cfg ServiceConfig) (*Service, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	return &Service{
		log:             cfg.Logger,
		authorizer:      cfg.Authorizer,
		oktaImportRules: cfg.OktaImportRules,
		oktaAssignments: cfg.OktaAssignments,
	}, nil
}

// ListOktaImportRules returns a paginated list of all Okta import rule resources.
func (s *Service) ListOktaImportRules(ctx context.Context, req *oktapb.ListOktaImportRulesRequest) (*oktapb.ListOktaImportRulesResponse, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbRead, types.VerbList)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	results, nextPageToken, err := s.oktaImportRules.ListOktaImportRules(ctx, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	importRulesV1 := make([]*types.OktaImportRuleV1, len(results))
	for i, r := range results {
		v1, ok := r.(*types.OktaImportRuleV1)
		if !ok {
			return nil, trace.BadParameter("unexpected Okta import rule type %T", r)
		}
		importRulesV1[i] = v1
	}

	return &oktapb.ListOktaImportRulesResponse{
		ImportRules:   importRulesV1,
		NextPageToken: nextPageToken,
	}, nil
}

// GetOktaImportRule returns the specified Okta import rule resources.
func (s *Service) GetOktaImportRule(ctx context.Context, req *oktapb.GetOktaImportRuleRequest) (*types.OktaImportRuleV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbRead)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	importRule, err := s.oktaImportRules.GetOktaImportRule(ctx, req.GetName())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	importRuleV1, ok := importRule.(*types.OktaImportRuleV1)
	if !ok {
		return nil, trace.BadParameter("unexpected Okta import rule type %T", importRule)
	}

	return importRuleV1, nil
}

// CreateOktaImportRule creates a new Okta import rule resource.
func (s *Service) CreateOktaImportRule(ctx context.Context, req *oktapb.CreateOktaImportRuleRequest) (*types.OktaImportRuleV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbCreate)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedRule, err := s.oktaImportRules.CreateOktaImportRule(ctx, req.GetImportRule())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedRuleV1, ok := returnedRule.(*types.OktaImportRuleV1)
	if !ok {
		return nil, trace.BadParameter("expected returned import rule of OktaImportRuleV1, got %T", returnedRuleV1)
	}
	return returnedRuleV1, trace.Wrap(err)
}

// UpdateOktaImportRule updates an existing Okta import rule resource.
func (s *Service) UpdateOktaImportRule(ctx context.Context, req *oktapb.UpdateOktaImportRuleRequest) (*types.OktaImportRuleV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbUpdate)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedRule, err := s.oktaImportRules.UpdateOktaImportRule(ctx, req.GetImportRule())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedRuleV1, ok := returnedRule.(*types.OktaImportRuleV1)
	if !ok {
		return nil, trace.BadParameter("expected returned import rule of OktaImportRuleV1, got %T", returnedRuleV1)
	}
	return returnedRuleV1, trace.Wrap(err)
}

// DeleteOktaImportRule removes the specified Okta import rule resource.
func (s *Service) DeleteOktaImportRule(ctx context.Context, req *oktapb.DeleteOktaImportRuleRequest) (*emptypb.Empty, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbDelete)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &emptypb.Empty{}, trace.Wrap(s.oktaImportRules.DeleteOktaImportRule(ctx, req.GetName()))
}

// DeleteAllOktaImportRules removes all Okta import rules.
func (s *Service) DeleteAllOktaImportRules(ctx context.Context, _ *oktapb.DeleteAllOktaImportRulesRequest) (*emptypb.Empty, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaImportRule, types.VerbDelete)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &emptypb.Empty{}, trace.Wrap(s.oktaImportRules.DeleteAllOktaImportRules(ctx))
}

// ListOktaAssignments returns a paginated list of all Okta assignment resources.
func (s *Service) ListOktaAssignments(ctx context.Context, req *oktapb.ListOktaAssignmentsRequest) (*oktapb.ListOktaAssignmentsResponse, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbList, types.VerbRead)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	results, nextPageToken, err := s.oktaAssignments.ListOktaAssignments(ctx, int(req.GetPageSize()), req.GetPageToken())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	assignmentsV1 := make([]*types.OktaAssignmentV1, len(results))
	for i, a := range results {
		v1, ok := a.(*types.OktaAssignmentV1)
		if !ok {
			return nil, trace.BadParameter("unexpected Okta assignment type %T", a)
		}
		assignmentsV1[i] = v1
	}

	return &oktapb.ListOktaAssignmentsResponse{
		Assignments:   assignmentsV1,
		NextPageToken: nextPageToken,
	}, nil
}

// GetOktaAssignment returns the specified Okta assignment resources.
func (s *Service) GetOktaAssignment(ctx context.Context, req *oktapb.GetOktaAssignmentRequest) (*types.OktaAssignmentV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbRead)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	assignment, err := s.oktaAssignments.GetOktaAssignment(ctx, req.GetName())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	assignmentV1, ok := assignment.(*types.OktaAssignmentV1)
	if !ok {
		return nil, trace.BadParameter("unexpected Okta assignment type %T", assignment)
	}

	return assignmentV1, nil
}

// CreateOktaAssignment creates a new Okta assignment resource.
func (s *Service) CreateOktaAssignment(ctx context.Context, req *oktapb.CreateOktaAssignmentRequest) (*types.OktaAssignmentV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbCreate)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedAssignment, err := s.oktaAssignments.CreateOktaAssignment(ctx, req.GetAssignment())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedAssignmentV1, ok := returnedAssignment.(*types.OktaAssignmentV1)
	if !ok {
		return nil, trace.BadParameter("expected returned import rule of OktaAssignmentV1, got %T", returnedAssignmentV1)
	}
	return returnedAssignmentV1, trace.Wrap(err)
}

// UpdateOktaAssignment updates an existing Okta assignment resource.
func (s *Service) UpdateOktaAssignment(ctx context.Context, req *oktapb.UpdateOktaAssignmentRequest) (*types.OktaAssignmentV1, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbUpdate)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedAssignment, err := s.oktaAssignments.UpdateOktaAssignment(ctx, req.GetAssignment())
	if err != nil {
		return nil, trace.Wrap(err)
	}
	returnedAssignmentV1, ok := returnedAssignment.(*types.OktaAssignmentV1)
	if !ok {
		return nil, trace.BadParameter("expected returned import rule of OktaAssignmentV1, got %T", returnedAssignmentV1)
	}
	return returnedAssignmentV1, trace.Wrap(err)
}

// DeleteOktaAssignment removes the specified Okta assignment resource.
func (s *Service) DeleteOktaAssignment(ctx context.Context, req *oktapb.DeleteOktaAssignmentRequest) (*emptypb.Empty, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbDelete)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &emptypb.Empty{}, trace.Wrap(s.oktaAssignments.DeleteOktaAssignment(ctx, req.GetName()))
}

// DeleteAllOktaAssignments removes all Okta assignments.
func (s *Service) DeleteAllOktaAssignments(ctx context.Context, _ *oktapb.DeleteAllOktaAssignmentsRequest) (*emptypb.Empty, error) {
	_, err := authz.AuthorizeWithVerbs(ctx, s.log, s.authorizer, true, types.KindOktaAssignment, types.VerbDelete)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &emptypb.Empty{}, trace.Wrap(s.oktaAssignments.DeleteAllOktaAssignments(ctx))
}
