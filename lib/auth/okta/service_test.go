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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	oktapb "github.com/gravitational/teleport/api/gen/proto/go/teleport/okta/v1"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/authz"
	"github.com/gravitational/teleport/lib/backend/memory"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/services/local"
	"github.com/gravitational/teleport/lib/tlsca"
)

func TestOktaImportRules(t *testing.T) {
	ctx, svc := initSvc(t, types.KindOktaImportRule)

	listResp, err := svc.ListOktaImportRules(ctx, &oktapb.ListOktaImportRulesRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, listResp.ImportRules)

	r1 := newOktaImportRule(t, "1")
	r2 := newOktaImportRule(t, "2")
	r3 := newOktaImportRule(t, "3")

	createResp, err := svc.CreateOktaImportRule(ctx, &oktapb.CreateOktaImportRuleRequest{ImportRule: r1})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(r1, createResp))

	createResp, err = svc.CreateOktaImportRule(ctx, &oktapb.CreateOktaImportRuleRequest{ImportRule: r2})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(r2, createResp))

	createResp, err = svc.CreateOktaImportRule(ctx, &oktapb.CreateOktaImportRuleRequest{ImportRule: r3})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(r3, createResp))

	listResp, err = svc.ListOktaImportRules(ctx, &oktapb.ListOktaImportRulesRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, cmp.Diff([]*types.OktaImportRuleV1{r1, r2, r3}, listResp.ImportRules,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	r1.SetExpiry(time.Now().Add(30 * time.Minute))
	updateResp, err := svc.UpdateOktaImportRule(ctx, &oktapb.UpdateOktaImportRuleRequest{ImportRule: r1})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(r1, updateResp))

	r, err := svc.GetOktaImportRule(ctx, &oktapb.GetOktaImportRuleRequest{Name: r1.GetName()})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(r1, r,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	_, err = svc.DeleteOktaImportRule(ctx, &oktapb.DeleteOktaImportRuleRequest{Name: r1.GetName()})
	require.NoError(t, err)

	listResp, err = svc.ListOktaImportRules(ctx, &oktapb.ListOktaImportRulesRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, cmp.Diff([]*types.OktaImportRuleV1{r2, r3}, listResp.ImportRules,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	_, err = svc.DeleteAllOktaImportRules(ctx, &oktapb.DeleteAllOktaImportRulesRequest{})
	require.NoError(t, err)

	listResp, err = svc.ListOktaImportRules(ctx, &oktapb.ListOktaImportRulesRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, listResp.ImportRules)
}

func TestOktaAssignments(t *testing.T) {
	ctx, svc := initSvc(t, types.KindOktaAssignment)

	listResp, err := svc.ListOktaAssignments(ctx, &oktapb.ListOktaAssignmentsRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, listResp.Assignments)

	a1 := newOktaAssignment(t, "1")
	a2 := newOktaAssignment(t, "2")
	a3 := newOktaAssignment(t, "3")

	createResp, err := svc.CreateOktaAssignment(ctx, &oktapb.CreateOktaAssignmentRequest{Assignment: a1})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(a1, createResp))

	createResp, err = svc.CreateOktaAssignment(ctx, &oktapb.CreateOktaAssignmentRequest{Assignment: a2})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(a2, createResp))

	createResp, err = svc.CreateOktaAssignment(ctx, &oktapb.CreateOktaAssignmentRequest{Assignment: a3})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(a3, createResp))

	listResp, err = svc.ListOktaAssignments(ctx, &oktapb.ListOktaAssignmentsRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, cmp.Diff([]*types.OktaAssignmentV1{a1, a2, a3}, listResp.Assignments,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	a1.SetExpiry(time.Now().Add(30 * time.Minute))
	updateResp, err := svc.UpdateOktaAssignment(ctx, &oktapb.UpdateOktaAssignmentRequest{Assignment: a1})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(a1, updateResp))

	a, err := svc.GetOktaAssignment(ctx, &oktapb.GetOktaAssignmentRequest{Name: a1.GetName()})
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(a1, a,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	_, err = svc.DeleteOktaAssignment(ctx, &oktapb.DeleteOktaAssignmentRequest{Name: a1.GetName()})
	require.NoError(t, err)

	listResp, err = svc.ListOktaAssignments(ctx, &oktapb.ListOktaAssignmentsRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, cmp.Diff([]*types.OktaAssignmentV1{a2, a3}, listResp.Assignments,
		cmpopts.IgnoreFields(types.Metadata{}, "ID")))

	_, err = svc.DeleteAllOktaAssignments(ctx, &oktapb.DeleteAllOktaAssignmentsRequest{})
	require.NoError(t, err)

	listResp, err = svc.ListOktaAssignments(ctx, &oktapb.ListOktaAssignmentsRequest{})
	require.NoError(t, err)
	require.Empty(t, listResp.NextPageToken)
	require.Empty(t, listResp.Assignments)
}

func initSvc(t *testing.T, kind string) (context.Context, *Service) {
	ctx := context.Background()
	backend, err := memory.New(memory.Config{})
	require.NoError(t, err)

	clusterConfigSvc, err := local.NewClusterConfigurationService(backend)
	require.NoError(t, err)
	trustSvc := local.NewCAService(backend)
	roleSvc := local.NewAccessService(backend)
	userSvc := local.NewIdentityService(backend)

	require.NoError(t, clusterConfigSvc.SetAuthPreference(ctx, types.DefaultAuthPreference()))
	require.NoError(t, clusterConfigSvc.SetClusterAuditConfig(ctx, types.DefaultClusterAuditConfig()))
	require.NoError(t, clusterConfigSvc.SetClusterNetworkingConfig(ctx, types.DefaultClusterNetworkingConfig()))
	require.NoError(t, clusterConfigSvc.SetSessionRecordingConfig(ctx, types.DefaultSessionRecordingConfig()))

	accessPoint := struct {
		services.ClusterConfiguration
		services.Trust
		services.RoleGetter
		services.UserGetter
	}{
		ClusterConfiguration: clusterConfigSvc,
		Trust:                trustSvc,
		RoleGetter:           roleSvc,
		UserGetter:           userSvc,
	}

	accessService := local.NewAccessService(backend)
	eventService := local.NewEventsService(backend)
	lockWatcher, err := services.NewLockWatcher(ctx, services.LockWatcherConfig{
		ResourceWatcherConfig: services.ResourceWatcherConfig{
			Client:    eventService,
			Component: "test",
		},
		LockGetter: accessService,
	})
	require.NoError(t, err)

	authorizer, err := authz.NewAuthorizer(authz.AuthorizerOpts{
		ClusterName: "test-cluster",
		AccessPoint: accessPoint,
		LockWatcher: lockWatcher,
	})
	require.NoError(t, err)

	role, err := types.NewRole("import-rules", types.RoleSpecV6{
		Allow: types.RoleConditions{
			Rules: []types.Rule{
				{
					Resources: []string{kind},
					Verbs:     []string{types.VerbList, types.VerbRead, types.VerbUpdate, types.VerbCreate, types.VerbDelete},
				},
			},
		},
	})
	require.NoError(t, err)
	roleSvc.CreateRole(ctx, role)
	require.NoError(t, err)

	user, err := types.NewUser("test-user")
	user.AddRole(role.GetName())
	require.NoError(t, err)
	userSvc.CreateUser(user)
	require.NoError(t, err)

	svc, err := NewService(ServiceConfig{
		Backend:    backend,
		Authorizer: authorizer,
	})
	require.NoError(t, err)

	ctx = authz.ContextWithUser(ctx, authz.LocalUser{
		Username: user.GetName(),
		Identity: tlsca.Identity{
			Username: user.GetName(),
			Groups:   []string{role.GetName()},
		},
	})

	return ctx, svc
}

func newOktaImportRule(t *testing.T, name string) *types.OktaImportRuleV1 {
	importRule, err := types.NewOktaImportRule(
		types.Metadata{
			Name: name,
		},
		types.OktaImportRuleSpecV1{
			Mappings: []*types.OktaImportRuleMappingV1{
				{
					Match: []*types.OktaImportRuleMatchV1{
						{
							AppIDs: []string{"yes"},
						},
					},
					AddLabels: map[string]string{
						"label1": "value1",
					},
				},
				{
					Match: []*types.OktaImportRuleMatchV1{
						{
							GroupIDs: []string{"yes"},
						},
					},
					AddLabels: map[string]string{
						"label1": "value1",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	return importRule.(*types.OktaImportRuleV1)
}

func newOktaAssignment(t *testing.T, name string) *types.OktaAssignmentV1 {
	assignment, err := types.NewOktaAssignment(
		types.Metadata{
			Name: name,
		},
		types.OktaAssignmentSpecV1{
			User: "test-user@test.user",
			Actions: []*types.OktaAssignmentActionV1{
				{
					Status: types.OktaAssignmentActionV1_PENDING,
					Target: &types.OktaAssignmentActionTargetV1{
						Type: types.OktaAssignmentActionTargetV1_APPLICATION,
						Id:   "123456",
					},
				},
				{
					Status: types.OktaAssignmentActionV1_SUCCESSFUL,
					Target: &types.OktaAssignmentActionTargetV1{
						Type: types.OktaAssignmentActionTargetV1_GROUP,
						Id:   "234567",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	return assignment.(*types.OktaAssignmentV1)
}
