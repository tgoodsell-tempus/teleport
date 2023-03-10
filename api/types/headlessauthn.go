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

package types

import (
	"github.com/gravitational/trace"
)

// CheckAndSetDefaults does basic validation and default setting.
func (h *HeadlessAuthentication) CheckAndSetDefaults() error {
	h.setStaticFields()
	if err := h.Metadata.CheckAndSetDefaults(); err != nil {
		return trace.Wrap(err)
	}

	if h.Metadata.Expires == nil || h.Metadata.Expires.IsZero() {
		return trace.BadParameter("headless authentication resource must have non-zero header.metadata.expires")
	}

	if h.Version == "" {
		h.Version = V1
	}
	if h.State == HeadlessAuthenticationState_HEADLESS_AUTHENTICATION_STATE_UNSPECIFIED {
		h.State = HeadlessAuthenticationState_HEADLESS_AUTHENTICATION_STATE_PENDING
	}

	return nil
}

// setStaticFields sets static resource header and metadata fields.
func (h *HeadlessAuthentication) setStaticFields() {
	h.Kind = KindHeadlessAuthentication
}
