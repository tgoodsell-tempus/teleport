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

package maintenance

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gravitational/trace"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gravitational/teleport/integrations/kube-agent-updater/pkg/basichttp"
	"github.com/gravitational/teleport/integrations/kube-agent-updater/pkg/cache"
	"github.com/gravitational/teleport/integrations/kube-agent-updater/pkg/constants"
)

// basicHTTPMaintenanceClient retrieves whether the target version represents a
// critical update from an HTTP endpoint. It should not be invoked immediately
// and must be wrapped in a cache layer in order to avoid spamming the version
// server in case of reconciliation errors.
// use BasicHTTPMaintenanceTrigger if you need to check if an update is critical.
type basicHTTPMaintenanceClient struct {
	baseURL *url.URL
	client  *basichttp.Client
}

// Get sends an HTTP GET request and returns whether the current target version
// represents a critical update.
func (b *basicHTTPMaintenanceClient) Get(ctx context.Context) (bool, error) {
	versionURL := b.baseURL.JoinPath(constants.MaintenancePath)
	body, err := b.client.GetContent(ctx, *versionURL)
	if err != nil {
		return false, trace.Wrap(err)
	}
	// Validating early that the payload can be converted to a boolean allows to
	// gracefully catch connectivity error caused by mitm infrastructure such as
	// corporate proxies.
	result, err := stringToBool(strings.TrimSpace(string(body)))
	return result, trace.Wrap(err)
}

// BasicHTTPMaintenanceTrigger gets the critical status from an HTTP response
// containing only a truthy or falsy string.
// This is used typically to trigger emergency maintenances from a file hosted
// in a s3 bucket or raw file served over HTTP.
// BasicHTTPMaintenanceTrigger uses basicHTTPMaintenanceClient and wraps it in a cache
// in order to mitigate the impact of too frequent reconciliations.
// The structure implements the maintenance.Trigger interface.
type BasicHTTPMaintenanceTrigger struct {
	name         string
	cachedGetter func(context.Context) (bool, error)
}

// Name implements maintenance.Triggernd returns the trigger name for logging
// and debugging pursposes.
func (g BasicHTTPMaintenanceTrigger) Name() string {
	return g.name
}

// Default returns what to do if the trigger can't be evaluated.
// BasicHTTPMaintenanceTrigger should fail open, so the function returns true.
func (g BasicHTTPMaintenanceTrigger) Default() bool {
	return false
}

// CanStart implements maintenance.Trigger
func (g BasicHTTPMaintenanceTrigger) CanStart(ctx context.Context, _ client.Object) (bool, error) {
	result, err := g.cachedGetter(ctx)
	return result, trace.Wrap(err)
}

// NewBasicHTTPMaintenanceTrigger builds and return a Trigger checking a public HTTP endpoint.
func NewBasicHTTPMaintenanceTrigger(name string, baseURL *url.URL) Trigger {
	client := &http.Client{
		Timeout: constants.HTTPTimeout,
	}
	httpMaintenanceClient := &basicHTTPMaintenanceClient{
		baseURL: baseURL,
		client:  &basichttp.Client{Client: client},
	}

	return BasicHTTPMaintenanceTrigger{name, cache.NewTimedMemoize[bool](httpMaintenanceClient.Get, constants.CacheDuration).Get}
}

func stringToBool(input string) (bool, error) {
	switch {
	case strings.EqualFold("true", input), strings.EqualFold("yes", input):
		return true, nil
	case strings.EqualFold("false", input), strings.EqualFold("no", input):
		return false, nil
	default:
		return false, trace.BadParameter("cannot convert input to boolean: %s", input)
	}
}
