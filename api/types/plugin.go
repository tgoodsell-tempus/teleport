/*
Copyright 2022 Gravitational, Inc.

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

package types

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gravitational/trace"
)

// PluginType represents the type of the plugin
type PluginType string

const (
	// PluginTypeUnknown is returned when no plugin type matches.
	PluginTypeUnknown PluginType = ""
	// PluginTypeSlack is the Slack access request plugin
	PluginTypeSlack = "slack"
)

// Plugin represents a plugin instance
type Plugin interface {
	// ResourceWithSecrets provides common resource methods.
	ResourceWithSecrets
	Clone() Plugin
	GetCredentials() PluginCredentials
	GetStatus() PluginStatus
	GetType() PluginType
	SetCredentials(PluginCredentials) error
	SetStatus(PluginStatus) error
}

// PluginCredentials are the credentials embedded in Plugin
type PluginCredentials interface {
	GetOauth2AccessToken() *PluginOAuth2AccessTokenCredentials
}

// PluginStatus is the plugin status
type PluginStatus interface {
	GetCode() PluginStatusCode
}

// NewPluginV1 creates a new PluginV1 resource.
func NewPluginV1(metadata Metadata, spec PluginSpecV1, creds *PluginCredentialsV1) *PluginV1 {
	p := &PluginV1{
		Metadata: metadata,
		Spec:     spec,
	}
	if creds != nil {
		p.SetCredentials(creds)
	}

	return p
}

// CheckAndSetDefaults checks validity of all parameters and sets defaults.
func (p *PluginV1) CheckAndSetDefaults() error {
	p.setStaticFields()

	if err := p.Metadata.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	switch settings := p.Spec.Settings.(type) {
	case *PluginSpecV1_SlackAccessPlugin:
		// Check settings.
		if settings.SlackAccessPlugin == nil {
			return trace.BadParameter("settings must be set")
		}
		if err := settings.SlackAccessPlugin.CheckAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}

		if p.Credentials == nil {
			// TODO: after credential exchange during creation is implemented,
			// this should validate that credentials are not empty
			break
		}
		if p.Credentials.GetOauth2AccessToken() == nil {
			return trace.BadParameter("Slack access plugin can only be used with OAuth2 access token credential type")
		}
		if err := p.Credentials.GetOauth2AccessToken().CheckAndSetDefaults(); err != nil {
			return trace.Wrap(err)
		}
	default:
		return trace.BadParameter("settings are not set or have an unknown type")
	}

	return nil
}

// WithoutSecrets returns an instance of resource without secrets.
func (p *PluginV1) WithoutSecrets() Resource {
	if p.Credentials == nil {
		return p
	}

	p2 := p.Clone().(*PluginV1)
	p2.SetCredentials(nil)
	return p2
}

func (p *PluginV1) setStaticFields() {
	p.Kind = KindPlugin
	p.Version = V1
}

// Clone returns a copy of the Plugin instance
func (p *PluginV1) Clone() Plugin {
	return proto.Clone(p).(*PluginV1)
}

// GetVersion returns resource version
func (p *PluginV1) GetVersion() string {
	return p.Version
}

// GetKind returns resource kind
func (p *PluginV1) GetKind() string {
	return p.Kind
}

// GetSubKind returns resource sub kind
func (p *PluginV1) GetSubKind() string {
	return p.SubKind
}

// SetSubKind sets resource subkind
func (p *PluginV1) SetSubKind(s string) {
	p.SubKind = s
}

// GetResourceID returns resource ID
func (p *PluginV1) GetResourceID() int64 {
	return p.Metadata.ID
}

// SetResourceID sets resource ID
func (p *PluginV1) SetResourceID(id int64) {
	p.Metadata.ID = id
}

// GetMetadata returns object metadata
func (p *PluginV1) GetMetadata() Metadata {
	return p.Metadata
}

// SetMetadata sets object metadata
func (p *PluginV1) SetMetadata(meta Metadata) {
	p.Metadata = meta
}

// Expiry returns expiry time for the object
func (p *PluginV1) Expiry() time.Time {
	return p.Metadata.Expiry()
}

// SetExpiry sets expiry time for the object
func (p *PluginV1) SetExpiry(expires time.Time) {
	p.Metadata.SetExpiry(expires)
}

// GetName returns the name of the User
func (p *PluginV1) GetName() string {
	return p.Metadata.Name
}

// SetName sets the name of the User
func (p *PluginV1) SetName(e string) {
	p.Metadata.Name = e
}

// GetCredentials implements Plugin
func (p *PluginV1) GetCredentials() PluginCredentials {
	return p.Credentials
}

// SetCredentials implements Plugin
func (p *PluginV1) SetCredentials(creds PluginCredentials) error {
	if creds == nil {
		p.Credentials = nil
		return nil
	}
	switch creds := creds.(type) {
	case *PluginCredentialsV1:
		p.Credentials = creds
	default:
		return trace.BadParameter("unsupported plugin credential type %T", creds)
	}
	return nil
}

// GetStatus implements Plugin
func (p *PluginV1) GetStatus() PluginStatus {
	return p.Status
}

// SetStatus implements Plugin
func (p *PluginV1) SetStatus(status PluginStatus) error {
	if status == nil {
		p.Status = PluginStatusV1{}
		return nil
	}
	switch status := status.(type) {
	case PluginStatusV1:
		p.Status = status
	default:
		return trace.BadParameter("unsupported plugin status type %T", status)
	}
	return nil
}

// GetType implements Plugin
func (p *PluginV1) GetType() PluginType {
	switch p.Spec.Settings.(type) {
	case *PluginSpecV1_SlackAccessPlugin:
		return PluginTypeSlack
	default:
		return PluginTypeUnknown
	}
}

// CheckAndSetDefaults validates and set the default values
func (s *PluginSlackAccessSettings) CheckAndSetDefaults() error {
	if s.FallbackChannel == "" {
		return trace.BadParameter("fallback_channel must be set")
	}

	return nil
}

// CheckAndSetDefaults validates and set the default values
func (c *PluginOAuth2AuthorizationCodeCredentials) CheckAndSetDefaults() error {
	if c.AuthorizationCode == "" {
		return trace.BadParameter("authorization_code must be set")
	}
	if c.RedirectUri == "" {
		return trace.BadParameter("redirect_uri must be set")
	}

	return nil
}

// CheckAndSetDefaults validates and set the default values
func (c *PluginOAuth2AccessTokenCredentials) CheckAndSetDefaults() error {
	if c.AccessToken == "" {
		return trace.BadParameter("access_token must be set")
	}
	if c.RefreshToken == "" {
		return trace.BadParameter("refresh_token must be set")
	}
	c.Expires = c.Expires.UTC()

	return nil
}

// GetCode returns the status code
func (c PluginStatusV1) GetCode() PluginStatusCode {
	return c.Code
}
