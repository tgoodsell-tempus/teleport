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

package metadata

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchInstallMethods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc        string
		getenv      func(string) string
		execCommand func(string, ...string) ([]byte, error)
		expected    []string
	}{
		{
			desc: "dockerfile if dockerfile",
			getenv: func(name string) string {
				if name == "TELEPORT_INSTALL_METHOD_DOCKERFILE" {
					return "true"
				}
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				return nil, trace.NotFound("command does not exist")
			},
			expected: []string{
				"dockerfile",
			},
		},
		{
			desc: "helm_kube_agent if helm",
			getenv: func(name string) string {
				if name == "TELEPORT_INSTALL_METHOD_HELM_KUBE_AGENT" {
					return "true"
				}
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				return nil, trace.NotFound("command does not exist")
			},
			expected: []string{
				"helm_kube_agent",
			},
		},
		{
			desc: "node_script if node script",
			getenv: func(name string) string {
				if name == "TELEPORT_INSTALL_METHOD_NODE_SCRIPT" {
					return "true"
				}
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				return nil, trace.NotFound("command does not exist")
			},
			expected: []string{
				"node_script",
			},
		},
		{
			desc: "systemctl if systemctl",
			getenv: func(name string) string {
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				if name != "systemctl" {
					return nil, trace.NotFound("command does not exist")
				}
				if len(args) != 2 {
					return nil, trace.NotFound("command does not exist")
				}
				if args[0] != "status" || args[1] != "teleport.service" {
					return nil, trace.NotFound("command does not exist")
				}
				output := `
● teleport.service - Teleport SSH Service
Loaded: loaded (/lib/systemd/system/teleport.service; enabled; vendor preset: enabled)
Active: active (running) since Wed 2022-11-09 10:52:49 UTC; 3 months 22 days ago
Main PID: 1815 (teleport)
	Tasks: 12 (limit: 1143)
Memory: 55.6M
	CPU: 2h 2min 27.181s
CGroup: /system.slice/teleport.service
		└─1815 /usr/local/bin/teleport start --pid-file=/run/teleport.pid
`
				return []byte(output), nil
			},
			expected: []string{
				"systemctl",
			},
		},
		{
			desc: "dockerfile and helm_kube_agent if dockerfile and helm",
			getenv: func(name string) string {
				if name == "TELEPORT_INSTALL_METHOD_DOCKERFILE" {
					return "true"
				}
				if name == "TELEPORT_INSTALL_METHOD_HELM_KUBE_AGENT" {
					return "true"
				}
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				return nil, trace.NotFound("command does not exist")
			},
			expected: []string{
				"dockerfile",
				"helm_kube_agent",
			},
		},
		{
			desc: "empty if none",
			getenv: func(name string) string {
				return ""
			},
			execCommand: func(name string, args ...string) ([]byte, error) {
				return nil, trace.NotFound("command does not exist")
			},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &fetchConfig{
				getenv:      tc.getenv,
				execCommand: tc.execCommand,
			}
			require.Equal(t, tc.expected, c.fetchInstallMethods())
		})
	}
}

func TestFetchContainerRuntime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc     string
		readFile func(string) ([]byte, error)
		expected string
	}{
		{
			desc: "docker if /.dockerenv exists",
			readFile: func(name string) ([]byte, error) {
				if name != "/.dockerenv" {
					return nil, trace.NotFound("file does not exist")
				}
				return []byte{}, nil
			},
			expected: "docker",
		},
		{
			desc: "empty if /.dockerenv does not exist",
			readFile: func(name string) ([]byte, error) {
				return nil, trace.NotFound("file does not exist")
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &fetchConfig{
				readFile: tc.readFile,
			}
			require.Equal(t, tc.expected, c.fetchContainerRuntime())
		})
	}
}

// newFakeClientSet builds a fake clientSet reporting a specific kubernetes
// version.  This is used to test version-specific behaviors.
func newFakeClientSet(gitVersion string) *fakeClientSet {
	cs := fakeClientSet{}
	cs.discovery = fakediscovery.FakeDiscovery{
		Fake: &cs.Fake,
		FakedServerVersion: &version.Info{
			GitVersion: gitVersion,
		},
	}
	return &cs
}

type fakeClientSet struct {
	fake.Clientset
	discovery fakediscovery.FakeDiscovery
}

// Discovery overrides the default fake.Clientset Discovery method and returns
// our custom discovery mock instead.
func (c *fakeClientSet) Discovery() discovery.DiscoveryInterface {
	return &c.discovery
}

func TestFetchContainerOrchestrator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc       string
		kubeClient kubernetes.Interface
		expected   string
	}{
		{
			desc:       "kubernetes with git version X",
			kubeClient: newFakeClientSet("X"),
			expected:   "kubernetes-X",
		},
		{
			desc:       "kubernetes with git version Y",
			kubeClient: newFakeClientSet("Y"),
			expected:   "kubernetes-Y",
		},
		{
			desc:       "empty if not on kubernetes",
			kubeClient: nil,
			expected:   "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &fetchConfig{
				kubeClient: tc.kubeClient,
			}
			require.Equal(t, tc.expected, c.fetchContainerOrchestrator())
		})
	}
}

func TestFetchCloudEnvironment(t *testing.T) {
	t.Parallel()

	success := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	testCases := []struct {
		desc     string
		httpDo   func(*http.Request) (*http.Response, error)
		expected string
	}{
		{
			desc: "aws if on aws",
			httpDo: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "http://169.254.169.254/latest/meta-data/" {
					return nil, trace.NotFound("not found")
				}
				if len(req.Header) != 0 {
					return nil, trace.NotFound("not found")
				}
				return success, nil
			},
			expected: "aws",
		},
		{
			desc: "gcp if on gcp ",
			httpDo: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "http://metadata.google.internal/computeMetadata/v1" {
					return nil, trace.NotFound("not found")
				}
				if len(req.Header) != 1 {
					return nil, trace.NotFound("not found")
				}
				if len(req.Header["Metadata-Flavor"]) != 1 {
					return nil, trace.NotFound("not found")
				}
				if req.Header["Metadata-Flavor"][0] != "Google" {
					return nil, trace.NotFound("not found")
				}
				return success, nil
			},
			expected: "gcp",
		},
		{
			desc: "azure if on azure",
			httpDo: func(req *http.Request) (*http.Response, error) {
				if req.URL.String() != "http://169.254.169.254/metadata/instance?api-version=2021-02-01" {
					return nil, trace.NotFound("not found")
				}
				if len(req.Header) != 1 {
					return nil, trace.NotFound("not found")
				}
				if len(req.Header["Metadata"]) != 1 {
					return nil, trace.NotFound("not found")
				}
				if req.Header["Metadata"][0] != "true" {
					return nil, trace.NotFound("not found")
				}
				return success, nil
			},
			expected: "azure",
		},
		{
			desc: "empty if not aws, gcp nor azure",
			httpDo: func(req *http.Request) (*http.Response, error) {
				return nil, trace.NotFound("not found")
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := &fetchConfig{
				context: context.Background(),
				httpDo:  tc.httpDo,
			}
			require.Equal(t, tc.expected, c.fetchCloudEnvironment())
		})
	}
}
