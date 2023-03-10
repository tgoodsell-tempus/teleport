/**
 * Copyright 2022 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"fmt"
	"testing"

	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/defaults"
)

func getTestVal(isTestField bool, testVal string) string {
	if isTestField {
		return testVal
	}

	return "_"
}

func TestServerSorter(t *testing.T) {
	t.Parallel()

	testValsUnordered := []string{"d", "b", "a", "c"}

	makeServers := func(testVals []string, testField string) []Server {
		servers := make([]Server, len(testVals))
		for i := 0; i < len(testVals); i++ {
			testVal := testVals[i]
			var err error
			servers[i], err = NewServer(
				getTestVal(testField == ResourceMetadataName, testVal),
				KindNode,
				ServerSpecV2{
					Hostname: getTestVal(testField == ResourceSpecHostname, testVal),
					Addr:     getTestVal(testField == ResourceSpecAddr, testVal),
				})
			require.NoError(t, err)
		}
		return servers
	}

	cases := []struct {
		name      string
		wantErr   bool
		fieldName string
	}{
		{
			name:      "by name",
			fieldName: ResourceMetadataName,
		},
		{
			name:      "by hostname",
			fieldName: ResourceSpecHostname,
		},
		{
			name:      "by addr",
			fieldName: ResourceSpecAddr,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%s desc", c.name), func(t *testing.T) {
			sortBy := SortBy{Field: c.fieldName, IsDesc: true}
			servers := Servers(makeServers(testValsUnordered, c.fieldName))
			require.NoError(t, servers.SortByCustom(sortBy))
			targetVals, err := servers.GetFieldVals(c.fieldName)
			require.NoError(t, err)
			require.IsDecreasing(t, targetVals)
		})

		t.Run(fmt.Sprintf("%s asc", c.name), func(t *testing.T) {
			sortBy := SortBy{Field: c.fieldName}
			servers := Servers(makeServers(testValsUnordered, c.fieldName))
			require.NoError(t, servers.SortByCustom(sortBy))
			targetVals, err := servers.GetFieldVals(c.fieldName)
			require.NoError(t, err)
			require.IsIncreasing(t, targetVals)
		})
	}

	// Test error.
	sortBy := SortBy{Field: "unsupported"}
	servers := makeServers(testValsUnordered, "does-not-matter")
	require.True(t, trace.IsNotImplemented(Servers(servers).SortByCustom(sortBy)))
}

func TestServerCheckAndSetDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		server    *ServerV2
		assertion func(t *testing.T, s *ServerV2, err error)
	}{
		{
			name: "Teleport node",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindTeleportNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:       "1.2.3.4:3022",
					Hostname:   "teleport-node",
					PublicAddr: "1.2.3.4:3080",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.NoError(t, err)
				expectedServer := &ServerV2{
					Kind:    KindNode,
					SubKind: SubKindTeleportNode,
					Version: V2,
					Metadata: Metadata{
						Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
						Namespace: defaults.Namespace,
					},
					Spec: ServerSpecV2{
						Addr:       "1.2.3.4:3022",
						Hostname:   "teleport-node",
						PublicAddr: "1.2.3.4:3080",
					},
				}
				require.Equal(t, expectedServer, s)
			},
		},
		{
			name: "Teleport node subkind unset",
			server: &ServerV2{
				Kind:    KindNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:       "1.2.3.4:3022",
					Hostname:   "teleport-node",
					PublicAddr: "1.2.3.4:3080",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.NoError(t, err)
				expectedServer := &ServerV2{
					Kind:    KindNode,
					Version: V2,
					Metadata: Metadata{
						Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
						Namespace: defaults.Namespace,
					},
					Spec: ServerSpecV2{
						Addr:       "1.2.3.4:3022",
						Hostname:   "teleport-node",
						PublicAddr: "1.2.3.4:3080",
					},
				}
				require.Equal(t, expectedServer, s)
			},
		},
		{
			name: "OpenSSH node",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:     "1.2.3.4:3022",
					Hostname: "openssh-node",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.NoError(t, err)
				expectedServer := &ServerV2{
					Kind:    KindNode,
					SubKind: SubKindOpenSSHNode,
					Version: V2,
					Metadata: Metadata{
						Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
						Namespace: defaults.Namespace,
					},
					Spec: ServerSpecV2{
						Addr:     "1.2.3.4:3022",
						Hostname: "openssh-node",
					},
				}
				require.Equal(t, expectedServer, s)
			},
		},
		{
			name: "OpenSSH node with unset name",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Spec: ServerSpecV2{
					Addr:     "1.2.3.4:22",
					Hostname: "openssh-node",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, s.Metadata.Name)
			},
		},
		{
			name: "OpenSSH node with unset addr",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Hostname: "openssh-node",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.EqualError(t, err, `Addr must be set when server SubKind is "openssh"`)
			},
		},
		{
			name: "OpenSSH node with unset hostname",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr: "1.2.3.4:3022",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.EqualError(t, err, `Hostname must be set when server SubKind is "openssh"`)
			},
		},
		{
			name: "OpenSSH node with public addr",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:       "1.2.3.4:3022",
					Hostname:   "openssh-node",
					PublicAddr: "1.2.3.4:80",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.EqualError(t, err, `PublicAddr must not be set when server SubKind is "openssh"`)
			},
		},
		{
			name: "OpenSSH node with invalid addr",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: SubKindOpenSSHNode,
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:     "invalid-addr",
					Hostname: "openssh-node",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.ErrorContains(t, err, `invalid Addr "invalid-addr"`)
			},
		},
		{
			name: "node with invalid subkind",
			server: &ServerV2{
				Kind:    KindNode,
				SubKind: "invalid-subkind",
				Version: V2,
				Metadata: Metadata{
					Name:      "5da56852-2adb-4540-a37c-80790203f6a9",
					Namespace: defaults.Namespace,
				},
				Spec: ServerSpecV2{
					Addr:     "1.2.3.4:22",
					Hostname: "node",
				},
			},
			assertion: func(t *testing.T, s *ServerV2, err error) {
				require.EqualError(t, err, `invalid SubKind "invalid-subkind"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.server.CheckAndSetDefaults()
			tt.assertion(t, tt.server, err)
		})
	}
}
