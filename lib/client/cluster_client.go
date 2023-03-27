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

package client

import (
	"context"

	"github.com/gravitational/trace"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/ssh"

	"github.com/gravitational/teleport/api/client/proto"
	proxyclient "github.com/gravitational/teleport/api/client/proxy"
	"github.com/gravitational/teleport/lib/auth"
)

// ClusterClient facilitates communicating with both the
// Auth and Proxy services of a cluster.
type ClusterClient struct {
	tc          *TeleportClient
	ProxyClient *proxyclient.Client
	AuthClient  auth.ClientI
	Tracer      oteltrace.Tracer
}

// ClusterName returns the name of the cluster that the client
// is connected to.
func (c *ClusterClient) ClusterName() string {
	cluster := c.ProxyClient.ClusterName()
	if len(c.tc.JumpHosts) > 0 && cluster != "" {
		return cluster
	}

	return c.tc.SiteName
}

// Close terminates the connections to Auth and Proxy.
func (c *ClusterClient) Close() error {
	// close auth client first since it is tunneled through the proxy client
	return trace.NewAggregate(c.AuthClient.Close(), c.ProxyClient.Close())
}

// SessionSSHConfig returns the [ssh.ClientConfig] that should be used to connected to the
// provided target for the provided user. If per session MFA is required to establish the
// connection, then the MFA ceremony will be performed.
func (c *ClusterClient) SessionSSHConfig(ctx context.Context, user string, target NodeDetails) (*ssh.ClientConfig, error) {
	ctx, span := c.Tracer.Start(
		ctx,
		"clusterClient/SessionSSHConfig",
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			attribute.String("cluster", c.tc.SiteName),
		),
	)
	defer span.End()

	sshConfig := c.ProxyClient.SSHConfig(user)

	if target.MFACheck != nil && !target.MFACheck.Required {
		return sshConfig, nil
	}

	key, err := c.tc.localAgent.GetKey(target.Cluster, WithAllCerts...)
	if err != nil {
		if trace.IsNotFound(err) {
			// Either running inside the web UI in a proxy or using an identity
			// file. Fall back to whatever AuthMethod we currently have.
			return sshConfig, nil
		}
		return nil, trace.Wrap(err)
	}

	params := ReissueParams{
		NodeName:       nodeName(target.Addr),
		RouteToCluster: target.Cluster,
		MFACheck:       target.MFACheck,
	}

	// requiredCheck passed from param can be nil.
	if target.MFACheck == nil {
		check, err := c.AuthClient.IsMFARequired(ctx, params.isMFARequiredRequest(c.tc.HostLogin))
		if err != nil {
			return nil, trace.Wrap(err)
		}
		target.MFACheck = check
	}

	if !target.MFACheck.Required {
		log.Debug("MFA not required for access.")
		// MFA is not required.
		// SSH certs can be used without embedding the node name.
		if params.usage() == proto.UserCertsRequest_SSH {
			return sshConfig, nil
		}

		// All other targets need their name embedded in the cert for routing,
		// fall back to non-MFA reissue.
		key, err := c.reissueUserCerts(ctx, CertCacheKeep, params)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		am, err := key.AsAuthMethod()
		if err != nil {
			return nil, trace.Wrap(err)
		}

		sshConfig.Auth = []ssh.AuthMethod{am}
		return sshConfig, nil
	}

	// Always connect to root for getting new credentials, but attempt to reuse
	// the existing client if possible.
	rootClusterName, err := key.RootClusterName()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	mfaClt := c
	if params.RouteToCluster != rootClusterName {
		jumpHosts := c.tc.JumpHosts
		// In case of MFA connect to root teleport proxy instead of JumpHost to request
		// MFA certificates.
		if jumpHosts != nil {
			c.tc.JumpHosts = nil
		}
		clt, err := c.tc.ConnectToCluster(ctx)
		c.tc.JumpHosts = jumpHosts
		if err != nil {
			return nil, trace.Wrap(err)
		}

		mfaClt = clt
		defer clt.Close()
	}

	log.Debug("Attempting to issue a single-use user certificate with an MFA check.")
	key, err = performMFACeremony(ctx, mfaClt, params, key)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	log.Debug("Issued single-use user certificate after an MFA check.")
	am, err := key.AsAuthMethod()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	sshConfig.Auth = []ssh.AuthMethod{am}
	return sshConfig, nil
}

// reissueUserCerts gets new user certificates from the root Auth server.
func (c *ClusterClient) reissueUserCerts(ctx context.Context, cachePolicy CertCachePolicy, params ReissueParams) (*Key, error) {
	if params.RouteToCluster == "" {
		params.RouteToCluster = c.tc.SiteName
	}
	key := params.ExistingCreds
	if key == nil {
		var err error

		// Don't load the certs if we're going to drop all of them all as part
		// of the re-issue. If we load all of the old certs now we won't be able
		// to differentiate between legacy certificates (that need to be
		// deleted) and newly re-issued certs (that we definitely do *not* want
		// to delete) when it comes time to drop them from the local agent.
		var certOptions []CertOption
		if cachePolicy == CertCacheKeep {
			certOptions = WithAllCerts
		}

		key, err = c.tc.localAgent.GetKey(params.RouteToCluster, certOptions...)
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	req, err := c.prepareUserCertsRequest(params, key)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	root, err := c.tc.rootClusterName()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	clt, err := auth.NewClient(c.ProxyClient.ClientConfig(ctx, root))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	defer clt.Close()

	certs, err := clt.GenerateUserCerts(ctx, *req)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	key.ClusterName = params.RouteToCluster

	// Only update the parts of key that match the usage. See the docs on
	// proto.UserCertsRequest_CertUsage for which certificates match which
	// usage.
	//
	// This prevents us from overwriting the top-level key.TLSCert with
	// usage-restricted certificates.
	switch params.usage() {
	case proto.UserCertsRequest_All:
		key.Cert = certs.SSH
		key.TLSCert = certs.TLS
	case proto.UserCertsRequest_SSH:
		key.Cert = certs.SSH
	case proto.UserCertsRequest_App:
		key.AppTLSCerts[params.RouteToApp.Name] = certs.TLS
	case proto.UserCertsRequest_Database:
		dbCert, err := makeDatabaseClientPEM(params.RouteToDatabase.Protocol, certs.TLS, key)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		key.DBTLSCerts[params.RouteToDatabase.ServiceName] = dbCert
	case proto.UserCertsRequest_Kubernetes:
		key.KubeTLSCerts[params.KubernetesCluster] = certs.TLS
	case proto.UserCertsRequest_WindowsDesktop:
		key.WindowsDesktopCerts[params.RouteToWindowsDesktop.WindowsDesktop] = certs.TLS
	}
	return key, nil
}

// prepareUserCertsRequest creates a [proto.UserCertsRequest] with the fields
// set accordingly from the provided ReissueParams.
func (c *ClusterClient) prepareUserCertsRequest(params ReissueParams, key *Key) (*proto.UserCertsRequest, error) {
	tlsCert, err := key.TeleportTLSCertificate()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if len(params.AccessRequests) == 0 {
		// Get the active access requests to include in the cert.
		activeRequests, err := key.ActiveRequests()
		// key.ActiveRequests can return a NotFound error if it doesn't have an
		// SSH cert. That's OK, we just assume that there are no AccessRequests
		// in that case.
		if err != nil && !trace.IsNotFound(err) {
			return nil, trace.Wrap(err)
		}
		params.AccessRequests = activeRequests.AccessRequests
	}

	return &proto.UserCertsRequest{
		PublicKey:             key.MarshalSSHPublicKey(),
		Username:              tlsCert.Subject.CommonName,
		Expires:               tlsCert.NotAfter,
		RouteToCluster:        params.RouteToCluster,
		KubernetesCluster:     params.KubernetesCluster,
		AccessRequests:        params.AccessRequests,
		DropAccessRequests:    params.DropAccessRequests,
		RouteToDatabase:       params.RouteToDatabase,
		RouteToWindowsDesktop: params.RouteToWindowsDesktop,
		RouteToApp:            params.RouteToApp,
		NodeName:              params.NodeName,
		Usage:                 params.usage(),
		Format:                c.tc.CertificateFormat,
		RequesterName:         params.RequesterName,
	}, nil
}

// performMFACeremony runs the mfa ceremony to completion. If successful the returned
// [Key] will be authorized to connect to the target.
func performMFACeremony(ctx context.Context, clt *ClusterClient, params ReissueParams, key *Key) (*Key, error) {
	stream, err := clt.AuthClient.GenerateUserSingleUseCerts(ctx)
	if err != nil {
		if trace.IsNotImplemented(err) {
			// Probably talking to an older server, use the old non-MFA endpoint.
			log.WithError(err).Debug("Auth server does not implement GenerateUserSingleUseCerts.")
			// SSH certs can be used without reissuing.
			if params.usage() == proto.UserCertsRequest_SSH && key.Cert != nil {
				return key, nil
			}

			key, err := clt.reissueUserCerts(ctx, CertCacheKeep, params)
			return key, trace.Wrap(err)
		}
		return nil, trace.Wrap(err)
	}
	defer func() {
		// CloseSend closes the client side of the stream
		stream.CloseSend()
		// Recv to wait for the server side of the stream to end, this needs to
		// be called to ensure the spans are finished properly
		stream.Recv()
	}()

	initReq, err := clt.prepareUserCertsRequest(params, key)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	err = stream.Send(&proto.UserSingleUseCertsRequest{Request: &proto.UserSingleUseCertsRequest_Init{
		Init: initReq,
	}})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	mfaChal := resp.GetMFAChallenge()
	if mfaChal == nil {
		return nil, trace.BadParameter("server sent a %T on GenerateUserSingleUseCerts, expected MFAChallenge", resp.Response)
	}
	mfaResp, err := clt.tc.PromptMFAChallenge(ctx, clt.tc.WebProxyAddr, mfaChal, nil /* applyOpts */)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	err = stream.Send(&proto.UserSingleUseCertsRequest{Request: &proto.UserSingleUseCertsRequest_MFAResponse{MFAResponse: mfaResp}})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	resp, err = stream.Recv()
	if err != nil {
		return nil, trace.Wrap(err)
	}
	certResp := resp.GetCert()
	if certResp == nil {
		return nil, trace.BadParameter("server sent a %T on GenerateUserSingleUseCerts, expected SingleUseUserCert", resp.Response)
	}
	switch crt := certResp.Cert.(type) {
	case *proto.SingleUseUserCert_SSH:
		key.Cert = crt.SSH
	case *proto.SingleUseUserCert_TLS:
		switch initReq.Usage {
		case proto.UserCertsRequest_Kubernetes:
			key.KubeTLSCerts[initReq.KubernetesCluster] = crt.TLS
		case proto.UserCertsRequest_Database:
			dbCert, err := makeDatabaseClientPEM(params.RouteToDatabase.Protocol, crt.TLS, key)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			key.DBTLSCerts[params.RouteToDatabase.ServiceName] = dbCert
		case proto.UserCertsRequest_WindowsDesktop:
			key.WindowsDesktopCerts[params.RouteToWindowsDesktop.WindowsDesktop] = crt.TLS
		default:
			return nil, trace.BadParameter("server returned a TLS certificate but cert request usage was %s", initReq.Usage)
		}
	default:
		return nil, trace.BadParameter("server sent a %T SingleUseUserCert in response", certResp.Cert)
	}
	key.ClusterName = params.RouteToCluster

	return key, nil
}
