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

package aws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestExtractCredFromAuthHeader test the extractCredFromAuthHeader function logic.
func TestExtractCredFromAuthHeader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expCred *SigV4
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:  "valid header",
			input: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-date, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024",
			expCred: &SigV4{
				KeyID:     "AKIAIOSFODNN7EXAMPLE",
				Date:      "20130524",
				Region:    "us-east-1",
				Service:   "s3",
				Signature: "fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024",
				SignedHeaders: []string{
					"host",
					"range",
					"x-amz-date",
				},
			},
			wantErr: require.NoError,
		},
		{
			name:  "signed headers section missing",
			input: "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024",
			expCred: &SigV4{
				KeyID:     "AKIAIOSFODNN7EXAMPLE",
				Date:      "20130524",
				Region:    "us-east-1",
				Service:   "s3",
				Signature: "fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024",
			},
			wantErr: require.NoError,
		},
		{
			name:    "credential  section missing",
			input:   "AWS4-HMAC-SHA256 SignedHeaders=host;range;x-amz-date, Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024",
			wantErr: require.Error,
		},
		{
			name:    "invalid format",
			input:   "Credential=AKIAIOSFODNN7EXAMPLE/us-east-1/s3/aws4_request",
			wantErr: require.Error,
		},
		{
			name:    "missing credentials section",
			input:   "AWS4-HMAC-SHA256 SignedHeaders=host",
			wantErr: require.Error,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: require.Error,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSigV4(tc.input)
			tc.wantErr(t, err)
			require.Equal(t, tc.expCred, got)
		})
	}
}

// TestFilterAWSRoles verifies filtering AWS role ARNs by AWS account ID.
func TestFilterAWSRoles(t *testing.T) {
	acc1ARN1 := Role{
		ARN:     "arn:aws:iam::123456789012:role/EC2FullAccess",
		Display: "EC2FullAccess",
		Name:    "EC2FullAccess",
	}
	acc1ARN2 := Role{
		ARN:     "arn:aws:iam::123456789012:role/EC2ReadOnly",
		Display: "EC2ReadOnly",
		Name:    "EC2ReadOnly",
	}
	acc1ARN3 := Role{
		ARN:     "arn:aws:iam::123456789012:role/path/to/customrole",
		Display: "customrole",
		Name:    "path/to/customrole",
	}
	acc2ARN1 := Role{
		ARN:     "arn:aws:iam::210987654321:role/test-role",
		Display: "test-role",
		Name:    "test-role",
	}
	invalidARN := Role{
		ARN: "invalid-arn",
	}
	allARNS := []string{
		acc1ARN1.ARN, acc1ARN2.ARN, acc1ARN3.ARN, acc2ARN1.ARN, invalidARN.ARN,
	}
	tests := []struct {
		name      string
		accountID string
		outARNs   Roles
	}{
		{
			name:      "first account roles",
			accountID: "123456789012",
			outARNs:   Roles{acc1ARN1, acc1ARN2, acc1ARN3},
		},
		{
			name:      "second account roles",
			accountID: "210987654321",
			outARNs:   Roles{acc2ARN1},
		},
		{
			name:      "all roles",
			accountID: "",
			outARNs:   Roles{acc1ARN1, acc1ARN2, acc1ARN3, acc2ARN1},
		},
	}
	for _, test := range tests {
		require.Equal(t, test.outARNs, FilterAWSRoles(allARNS, test.accountID))
	}
}

func TestRoles(t *testing.T) {
	arns := []string{
		"arn:aws:iam::123456789012:role/test-role",
		"arn:aws:iam::123456789012:role/EC2FullAccess",
		"arn:aws:iam::123456789012:role/path/to/EC2FullAccess",
	}
	roles := FilterAWSRoles(arns, "123456789012")
	require.Len(t, roles, 3)

	t.Run("Sort", func(t *testing.T) {
		roles.Sort()
		require.Equal(t, "arn:aws:iam::123456789012:role/EC2FullAccess", roles[0].ARN)
		require.Equal(t, "arn:aws:iam::123456789012:role/path/to/EC2FullAccess", roles[1].ARN)
		require.Equal(t, "arn:aws:iam::123456789012:role/test-role", roles[2].ARN)
	})

	t.Run("FindRoleByARN", func(t *testing.T) {
		t.Run("found", func(t *testing.T) {
			for _, arn := range arns {
				role, found := roles.FindRoleByARN(arn)
				require.True(t, found)
				require.Equal(t, role.ARN, arn)
			}
		})

		t.Run("not found", func(t *testing.T) {
			_, found := roles.FindRoleByARN("arn:aws:iam::123456788912:role/unknown")
			require.False(t, found)
		})
	})

	t.Run("FindRolesByName", func(t *testing.T) {
		t.Run("found zero", func(t *testing.T) {
			rolesWithName := roles.FindRolesByName("unknown")
			require.Empty(t, rolesWithName)
		})

		t.Run("found one", func(t *testing.T) {
			rolesWithName := roles.FindRolesByName("path/to/EC2FullAccess")
			require.Len(t, rolesWithName, 1)
			require.Equal(t, "path/to/EC2FullAccess", rolesWithName[0].Name)
		})

		t.Run("found two", func(t *testing.T) {
			rolesWithName := roles.FindRolesByName("EC2FullAccess")
			require.Len(t, rolesWithName, 2)
			require.Equal(t, "EC2FullAccess", rolesWithName[0].Display)
			require.Equal(t, "EC2FullAccess", rolesWithName[1].Display)
			require.NotEqual(t, rolesWithName[0].ARN, rolesWithName[1].ARN)
		})
	})
}

func TestValidateRoleARNAndExtractRoleName(t *testing.T) {
	tests := []struct {
		name           string
		inputARN       string
		inputPartition string
		inputAccountID string
		wantRoleName   string
		wantError      bool
	}{
		{
			name:           "success",
			inputARN:       "arn:aws:iam::123456789012:role/role-name",
			inputPartition: "aws",
			inputAccountID: "123456789012",
			wantRoleName:   "role-name",
		},
		{
			name:           "invalid arn",
			inputARN:       "arn::::aws:iam::123456789012:role/role-name",
			inputPartition: "aws",
			inputAccountID: "123456789012",
			wantError:      true,
		},
		{
			name:           "invalid partition",
			inputARN:       "arn:aws:iam::123456789012:role/role-name",
			inputPartition: "aws-cn",
			inputAccountID: "123456789012",
			wantError:      true,
		},
		{
			name:           "invalid account ID",
			inputARN:       "arn:aws:iam::123456789012:role/role-name",
			inputPartition: "aws",
			inputAccountID: "123456789000",
			wantError:      true,
		},
		{
			name:           "not role arn",
			inputARN:       "arn:aws:iam::123456789012:user/username",
			inputPartition: "aws",
			inputAccountID: "123456789012",
			wantError:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualRoleName, err := ValidateRoleARNAndExtractRoleName(test.inputARN, test.inputPartition, test.inputAccountID)
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.wantRoleName, actualRoleName)
			}
		})
	}
}

func TestParseRoleARN(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		arn             string
		wantErrContains string
	}{
		"valid role arn": {
			arn: "arn:aws:iam::123456789012:role/test-role",
		},
		"arn fails to parse": {
			arn:             "foobar",
			wantErrContains: "invalid AWS ARN",
		},
		"sts arn is not iam": {
			arn:             "arn:aws:sts::123456789012:federated-user/Alice",
			wantErrContains: "not an AWS IAM role",
		},
		"iam arn is not a role": {
			arn:             "arn:aws:iam::123456789012:user/test-user",
			wantErrContains: "not an AWS IAM role",
		},
		"iam role arn is missing role name": {
			arn:             "arn:aws:iam::123456789012:role",
			wantErrContains: "missing AWS IAM role name",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseRoleARN(tt.arn)
			if tt.wantErrContains != "" {
				require.Error(t, err, err.Error())
				require.ErrorContains(t, err, tt.wantErrContains)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestBuildRoleARN(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		user      string
		region    string
		accountID string
		wantErr   string
		wantARN   string
	}{
		"valid role arn in correct partition and account": {
			user:      "arn:aws:iam::123456789012:role/test-role",
			region:    "us-west-1",
			accountID: "123456789012",
			wantARN:   "arn:aws:iam::123456789012:role/test-role",
		},
		"valid role arn in correct account and default partition": {
			user:      "arn:aws:iam::123456789012:role/test-role",
			region:    "",
			accountID: "123456789012",
			wantARN:   "arn:aws:iam::123456789012:role/test-role",
		},
		"valid role arn in default partition and account": {
			user:      "arn:aws:iam::123456789012:role/test-role",
			region:    "",
			accountID: "",
			wantARN:   "arn:aws:iam::123456789012:role/test-role",
		},
		"role name with prefix in default partition and account": {
			user:      "role/test-role",
			region:    "",
			accountID: "123456789012",
			wantARN:   "arn:aws:iam::123456789012:role/test-role",
		},
		"role name in default partition and account": {
			user:      "test-role",
			region:    "",
			accountID: "123456789012",
			wantARN:   "arn:aws:iam::123456789012:role/test-role",
		},
		"role name in china partition and account": {
			user:      "test-role",
			region:    "cn-north-1",
			accountID: "123456789012",
			wantARN:   "arn:aws-cn:iam::123456789012:role/test-role",
		},
		"valid ARN is not an IAM role ARN": {
			user:      "arn:aws:iam::123456789012:user/test-user",
			region:    "",
			accountID: "",
			wantErr:   "not an AWS IAM role",
		},
		"valid role arn in different partition": {
			user:      "arn:aws-cn:iam::123456789012:role/test-role",
			region:    "us-west-1",
			accountID: "",
			wantErr:   `expected AWS partition "aws" but got "aws-cn"`,
		},
		"valid role arn in different account": {
			user:      "arn:aws:iam::123456789012:role/test-role",
			region:    "us-west-1",
			accountID: "111222333444",
			wantErr:   `expected AWS account ID "111222333444" but got "123456789012"`,
		},
		"role name with invalid account characters": {
			user:      "test-role",
			region:    "",
			accountID: "12345678901f",
			wantErr:   "must be 12-digit",
		},
		"role name with invalid account id length": {
			user:      "test-role",
			region:    "",
			accountID: "1234567890123",
			wantErr:   "must be 12-digit",
		},
	}

	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := BuildRoleARN(tt.user, tt.region, tt.accountID)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, got)
			require.Equal(t, tt.wantARN, got)
		})
	}
}

// TODO(gavin): use this as a base for testing, and then delete it.
// func TestRedshiftServerlessUsernameToRoleARN(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		inputUsername string
// 		expectRoleARN string
// 		expectError   bool
// 	}{
// 		{
// 			inputUsername: "arn:aws:iam::123456789012:role/rolename",
// 			expectRoleARN: "arn:aws:iam::123456789012:role/rolename",
// 		},
// 		{
// 			inputUsername: "arn:aws:iam::123456789012:user/user",
// 			expectError:   true,
// 		},
// 		{
// 			inputUsername: "arn:aws:not-iam::123456789012:role/rolename",
// 			expectError:   true,
// 		},
// 		{
// 			inputUsername: "role/rolename",
// 			expectRoleARN: "arn:aws:iam::123456789012:role/rolename",
// 		},
// 		{
// 			inputUsername: "rolename",
// 			expectRoleARN: "arn:aws:iam::123456789012:role/rolename",
// 		},
// 		{
// 			inputUsername: "IAM:user",
// 			expectError:   true,
// 		},
// 		{
// 			inputUsername: "IAMR:rolename",
// 			expectError:   true,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.inputUsername, func(t *testing.T) {
// 			actualRoleARN, err := redshiftServerlessUsernameToRoleARN(newRedshiftServerlessDatabase(t).GetAWS(), test.inputUsername)
// 			if test.expectError {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, test.expectRoleARN, actualRoleARN)
// 			}
// 		})
// 	}
// }
