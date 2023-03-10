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

package usagereporter

import (
	"github.com/gravitational/teleport/api/types"
	apievents "github.com/gravitational/teleport/api/types/events"
)

const (
	// tcpSessionType is the session_type in tp.session.start for TCP
	// Application Access.
	tcpSessionType = "app_tcp"
	// portSessionType is the session_type in tp.session.start for SSH port
	// forwarding.
	portSessionType = "ssh_port"
)

func ConvertAuditEvent(event apievents.AuditEvent) Anonymizable {
	switch e := event.(type) {
	case *apievents.UserLogin:
		// Only count successful logins.
		if !e.Success {
			return nil
		}

		// Note: we can have different behavior based on event code (local vs
		// SSO) if desired, but we currently only care about connector type /
		// method
		return &UserLoginEvent{
			UserName:      e.User,
			ConnectorType: e.Method,
		}

	case *apievents.SessionStart:
		// Note: session.start is only SSH and Kubernetes.
		sessionType := types.SSHSessionKind
		if e.KubernetesCluster != "" {
			sessionType = types.KubernetesSessionKind
		}

		return &SessionStartEvent{
			UserName:    e.User,
			SessionType: string(sessionType),
		}
	case *apievents.PortForward:
		return &SessionStartEvent{
			UserName:    e.User,
			SessionType: portSessionType,
		}
	case *apievents.DatabaseSessionStart:
		return &SessionStartEvent{
			UserName:    e.User,
			SessionType: string(types.DatabaseSessionKind),
		}
	case *apievents.AppSessionStart:
		sessionType := string(types.AppSessionKind)
		if types.IsAppTCP(e.AppURI) {
			sessionType = tcpSessionType
		}
		return &SessionStartEvent{
			UserName:    e.User,
			SessionType: sessionType,
		}
	case *apievents.WindowsDesktopSessionStart:
		return &SessionStartEvent{
			UserName:    e.User,
			SessionType: string(types.WindowsDesktopSessionKind),
		}

	case *apievents.GithubConnectorCreate:
		return &SSOCreateEvent{
			ConnectorType: types.KindGithubConnector,
		}
	case *apievents.OIDCConnectorCreate:
		return &SSOCreateEvent{
			ConnectorType: types.KindOIDCConnector,
		}
	case *apievents.SAMLConnectorCreate:
		return &SSOCreateEvent{
			ConnectorType: types.KindSAMLConnector,
		}

	case *apievents.RoleCreate:
		return &RoleCreateEvent{
			UserName: e.User,
			RoleName: e.ResourceMetadata.Name,
		}

	case *apievents.KubeRequest:
		return &KubeRequestEvent{
			UserName: e.User,
		}

	case *apievents.SFTP:
		return &SFTPEvent{
			UserName: e.User,
			Action:   int32(e.Action),
		}
	}

	return nil
}
