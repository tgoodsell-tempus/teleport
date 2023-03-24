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

package testlib

import (
	"context"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/integration/helpers"
	resourcesv1 "github.com/gravitational/teleport/integrations/operator/apis/resources/v1"
	resourcesv2 "github.com/gravitational/teleport/integrations/operator/apis/resources/v2"
	resourcesv3 "github.com/gravitational/teleport/integrations/operator/apis/resources/v3"
	resourcesv5 "github.com/gravitational/teleport/integrations/operator/apis/resources/v5"
	"github.com/gravitational/teleport/integrations/operator/controllers/resources"
	"github.com/gravitational/teleport/lib/modules"
	"github.com/gravitational/teleport/lib/service/servicecfg"
)

func createNamespaceForTest(t *testing.T, kc kclient.Client) *core.Namespace {
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ValidRandomResourceName("ns-")},
	}

	err := kc.Create(context.Background(), ns)
	require.NoError(t, err)

	return ns
}

func deleteNamespaceForTest(t *testing.T, kc kclient.Client, ns *core.Namespace) {
	err := kc.Delete(context.Background(), ns)
	require.NoError(t, err)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func ValidRandomResourceName(prefix string) string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return prefix + string(b)
}

func defaultTeleportServiceConfig(t *testing.T) (*helpers.TeleInstance, string) {
	modules.SetTestModules(t, &modules.TestModules{
		TestBuildType: modules.BuildEnterprise,
		TestFeatures: modules.Features{
			OIDC: true,
			SAML: true,
		},
	})

	teleportServer := helpers.NewInstance(t, helpers.InstanceConfig{
		ClusterName: "root.example.com",
		HostID:      uuid.New().String(),
		NodeName:    helpers.Loopback,
		Log:         logrus.StandardLogger(),
	})

	rcConf := servicecfg.MakeDefaultConfig()
	rcConf.DataDir = t.TempDir()
	rcConf.Auth.Enabled = true
	rcConf.Proxy.Enabled = true
	rcConf.Proxy.DisableWebInterface = true
	rcConf.SSH.Enabled = true
	rcConf.Version = "v2"

	roleName := ValidRandomResourceName("role-")
	unrestricted := []string{"list", "create", "read", "update", "delete"}
	role, err := types.NewRole(roleName, types.RoleSpecV6{
		Allow: types.RoleConditions{
			Rules: []types.Rule{
				types.NewRule("role", unrestricted),
				types.NewRule("user", unrestricted),
				types.NewRule("auth_connector", unrestricted),
				types.NewRule("login_rule", unrestricted),
			},
		},
	})
	require.NoError(t, err)

	operatorName := ValidRandomResourceName("operator-")
	_ = teleportServer.AddUserWithRole(operatorName, role)

	err = teleportServer.CreateEx(t, nil, rcConf)
	require.NoError(t, err)

	return teleportServer, operatorName
}

func FastEventually(t *testing.T, condition func() bool) {
	require.Eventually(t, condition, time.Second, 100*time.Millisecond)
}

func clientForTeleport(t *testing.T, teleportServer *helpers.TeleInstance, userName string) *client.Client {
	identityFilePath := helpers.MustCreateUserIdentityFile(t, teleportServer, userName, time.Hour)
	creds := client.LoadIdentityFile(identityFilePath)
	return clientWithCreds(t, teleportServer.Auth, creds)
}

func clientWithCreds(t *testing.T, authAddr string, creds client.Credentials) *client.Client {
	c, err := client.New(context.Background(), client.Config{
		Addrs:       []string{authAddr},
		Credentials: []client.Credentials{creds},
	})
	require.NoError(t, err)
	return c
}

type TestSetup struct {
	TeleportClient *client.Client
	K8sClient      kclient.Client
	K8sRestConfig  *rest.Config
	Namespace      *core.Namespace
	Operator       manager.Manager
	OperatorCancel context.CancelFunc
	OperatorName   string
}

// StartKubernetesOperator creates and start a new operator
func (s *TestSetup) StartKubernetesOperator(t *testing.T) {
	// If there was an operator running previously we make sure it is stopped
	if s.OperatorCancel != nil {
		s.StopKubernetesOperator()
	}

	// We have to create a new Manager on each start because the Manager does not support to be restarted
	clientAccessor := func(ctx context.Context) (*client.Client, error) {
		return s.TeleportClient, nil
	}

	k8sManager, err := ctrl.NewManager(s.K8sRestConfig, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	require.NoError(t, err)

	err = (&resources.RoleReconciler{
		Client:                 s.K8sClient,
		Scheme:                 k8sManager.GetScheme(),
		TeleportClientAccessor: clientAccessor,
	}).SetupWithManager(k8sManager)
	require.NoError(t, err)

	err = resources.NewUserReconciler(s.K8sClient, clientAccessor).SetupWithManager(k8sManager)
	require.NoError(t, err)

	err = resources.NewGithubConnectorReconciler(s.K8sClient, clientAccessor).SetupWithManager(k8sManager)
	require.NoError(t, err)

	err = resources.NewOIDCConnectorReconciler(s.K8sClient, clientAccessor).SetupWithManager(k8sManager)
	require.NoError(t, err)

	err = resources.NewSAMLConnectorReconciler(s.K8sClient, clientAccessor).SetupWithManager(k8sManager)
	require.NoError(t, err)

	err = resources.NewLoginRuleReconciler(s.K8sClient, clientAccessor).SetupWithManager(k8sManager)
	require.NoError(t, err)

	ctx, ctxCancel := context.WithCancel(context.Background())

	s.Operator = k8sManager
	s.OperatorCancel = ctxCancel

	go func() {
		err := s.Operator.Start(ctx)
		assert.NoError(t, err)
	}()
}

func (s *TestSetup) StopKubernetesOperator() {
	s.OperatorCancel()
}

func setupTeleportClient(t *testing.T, setup *TestSetup) {
	// Override teleport client with client to locally connected teleport
	// cluster (with default tsh credentials).
	if addr := os.Getenv("OPERATOR_TEST_TELEPORT_ADDR"); addr != "" {
		creds := client.LoadProfile("", "")
		setup.TeleportClient = clientWithCreds(t, addr, creds)
		return
	}

	// A TestOption already provided a TeleportClient, return.
	if setup.TeleportClient != nil {
		return
	}

	// Start a Teleport server for the test and set up a client connected to
	// that server.
	teleportServer, operatorName := defaultTeleportServiceConfig(t)
	require.NoError(t, teleportServer.Start())
	setup.TeleportClient = clientForTeleport(t, teleportServer, operatorName)
	setup.OperatorName = operatorName
	t.Cleanup(func() {
		err := teleportServer.StopAll()
		require.NoError(t, err)
	})

	t.Cleanup(func() {
		err := setup.TeleportClient.Close()
		require.NoError(t, err)
	})
}

type TestOption func(*TestSetup)

func WithTeleportClient(clt *client.Client) TestOption {
	return func(setup *TestSetup) {
		setup.TeleportClient = clt
	}
}

// SetupTestEnv creates a Kubernetes server, a teleport server and starts the operator
func SetupTestEnv(t *testing.T, opts ...TestOption) *TestSetup {
	// Hack to get the path of this file in order to find the crd path no matter
	// where this is called from.
	_, thisFileName, _, _ := runtime.Caller(0)
	crdPath := filepath.Join(filepath.Dir(thisFileName), "..", "..", "..", "config", "crd", "bases")
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{crdPath},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	err = resourcesv1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	err = resourcesv5.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	err = resourcesv2.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	err = resourcesv3.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	k8sClient, err := kclient.New(cfg, kclient.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)

	ns := createNamespaceForTest(t, k8sClient)

	t.Cleanup(func() {
		deleteNamespaceForTest(t, k8sClient, ns)
		err = testEnv.Stop()
		require.NoError(t, err)
	})

	setup := &TestSetup{
		K8sClient:     k8sClient,
		Namespace:     ns,
		K8sRestConfig: cfg,
	}

	for _, opt := range opts {
		opt(setup)
	}

	setupTeleportClient(t, setup)

	// Create and start the Kubernetes operator
	setup.StartKubernetesOperator(t)

	t.Cleanup(func() {
		setup.StopKubernetesOperator()
	})

	return setup
}
