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

package local

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/backend/memory"
)

// TestOktaImportRuleCRUD tests backend operations with Okta import rule resources.
func TestOktaImportRuleCRUD(t *testing.T) {
	ctx := context.Background()
	clock := clockwork.NewFakeClock()

	backend, err := memory.New(memory.Config{
		Context: ctx,
		Clock:   clock,
	})
	require.NoError(t, err)

	service, err := NewOktaService(backend)
	require.NoError(t, err)

	// Create a couple Okta import rule.
	importRule1, err := types.NewOktaImportRule(
		types.Metadata{
			Name: "importRule1",
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
	importRule2, err := types.NewOktaImportRule(
		types.Metadata{
			Name: "importRule2",
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

	// Initially we expect no import rule.
	out, nextToken, err := service.ListOktaImportRules(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, out)

	// Create both import rules.
	importRule, err := service.CreateOktaImportRule(ctx, importRule1)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(importRule1, importRule,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	importRule, err = service.CreateOktaImportRule(ctx, importRule2)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(importRule2, importRule,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch all import rules.
	out, nextToken, err = service.ListOktaImportRules(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, cmp.Diff([]types.OktaImportRule{importRule1, importRule2}, out,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch a paginated list of import rules
	paginatedOut := make([]types.OktaImportRule, 0, 2)
	for {
		out, nextToken, err = service.ListOktaImportRules(ctx, 1, nextToken)
		require.NoError(t, err)

		paginatedOut = append(paginatedOut, out...)
		if nextToken == "" {
			break
		}
	}

	require.Len(t, paginatedOut, 2)
	require.Empty(t, cmp.Diff([]types.OktaImportRule{importRule1, importRule2}, paginatedOut,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch a specific import rule.
	importRule, err = service.GetOktaImportRule(ctx, importRule2.GetName())
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(importRule2, importRule,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Try to fetch an import rule that doesn't exist.
	_, err = service.GetOktaImportRule(ctx, "doesnotexist")
	require.True(t, trace.IsNotFound(err), "expected not found error, got %v", err)

	// Try to create the same import rule.
	_, err = service.CreateOktaImportRule(ctx, importRule1)
	require.True(t, trace.IsAlreadyExists(err), "expected already exists error, got %v", err)

	// Update an import rule.
	importRule1.SetExpiry(clock.Now().Add(30 * time.Minute))
	_, err = service.UpdateOktaImportRule(ctx, importRule1)
	require.NoError(t, err)
	importRule, err = service.GetOktaImportRule(ctx, importRule1.GetName())
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(importRule1, importRule,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Delete an import rule
	err = service.DeleteOktaImportRule(ctx, importRule1.GetName())
	require.NoError(t, err)
	out, nextToken, err = service.ListOktaImportRules(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, cmp.Diff([]types.OktaImportRule{importRule2}, out,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Try to delete an import rule that doesn't exist.
	err = service.DeleteOktaImportRule(ctx, "doesnotexist")
	require.True(t, trace.IsNotFound(err), "expected not found error, got %v", err)

	// Delete all import rules.
	err = service.DeleteAllOktaImportRules(ctx)
	require.NoError(t, err)
	out, nextToken, err = service.ListOktaImportRules(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, out)
}

// TestOktaAssignmentCRUD tests backend operations with Okta assignment resources.
func TestOktaAssignmentCRUD(t *testing.T) {
	ctx := context.Background()
	clock := clockwork.NewFakeClock()

	backend, err := memory.New(memory.Config{
		Context: ctx,
		Clock:   clock,
	})
	require.NoError(t, err)

	service, err := NewOktaService(backend)
	require.NoError(t, err)

	// Create a couple Okta assignments.
	assignment1, err := types.NewOktaAssignment(
		types.Metadata{
			Name: "assignment1",
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
	assignment2, err := types.NewOktaAssignment(
		types.Metadata{
			Name: "assignment2",
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

	// Initially we expect no assignments.
	out, nextToken, err := service.ListOktaAssignments(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, out)

	// Create both assignments.
	assignment, err := service.CreateOktaAssignment(ctx, assignment1)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment1, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	assignment, err = service.CreateOktaAssignment(ctx, assignment2)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment2, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch all assignments.
	out, nextToken, err = service.ListOktaAssignments(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, cmp.Diff([]types.OktaAssignment{assignment1, assignment2}, out,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch a paginated list of assignments
	paginatedOut := make([]types.OktaAssignment, 0, 2)
	numPages := 0
	for {
		numPages++
		out, nextToken, err = service.ListOktaAssignments(ctx, 1, nextToken)
		require.NoError(t, err)

		paginatedOut = append(paginatedOut, out...)
		if nextToken == "" {
			break
		}
	}

	require.Equal(t, 2, numPages)
	require.Empty(t, cmp.Diff([]types.OktaAssignment{assignment1, assignment2}, paginatedOut,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Fetch a specific assignment.
	assignment, err = service.GetOktaAssignment(ctx, assignment2.GetName())
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment2, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Try to fetch an assignment that doesn't exist.
	_, err = service.GetOktaAssignment(ctx, "doesnotexist")
	require.True(t, trace.IsNotFound(err), "expected not found error, got %v", err)

	// Try to create the same assignment.
	_, err = service.CreateOktaAssignment(ctx, assignment1)
	require.True(t, trace.IsAlreadyExists(err), "expected already exists error, got %v", err)

	// Update an assignment.
	assignment1.SetExpiry(clock.Now().Add(30 * time.Minute))
	_, err = service.UpdateOktaAssignment(ctx, assignment1)
	require.NoError(t, err)
	assignment, err = service.GetOktaAssignment(ctx, assignment1.GetName())
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment1, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Update the statuses for an assignment.
	assignment1.GetActions()[0].SetStatus(constants.OktaAssignmentActionStatusProcessing)
	assignment, err = service.UpdateOktaAssignmentActionStatuses(ctx, assignment1.GetName(), constants.OktaAssignmentActionStatusProcessing)
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment1, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))
	assignment, err = service.GetOktaAssignment(ctx, assignment1.GetName())
	require.NoError(t, err)
	require.Empty(t, cmp.Diff(assignment1, assignment,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Delete an assignment
	err = service.DeleteOktaAssignment(ctx, assignment1.GetName())
	require.NoError(t, err)
	out, nextToken, err = service.ListOktaAssignments(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, cmp.Diff([]types.OktaAssignment{assignment2}, out,
		cmpopts.IgnoreFields(types.Metadata{}, "ID"),
	))

	// Try to delete an assignment that doesn't exist.
	err = service.DeleteOktaAssignment(ctx, "doesnotexist")
	require.True(t, trace.IsNotFound(err), "expected not found error, got %v", err)

	// Delete all assignments.
	err = service.DeleteAllOktaAssignments(ctx)
	require.NoError(t, err)
	out, nextToken, err = service.ListOktaAssignments(ctx, 200, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Empty(t, out)
}
