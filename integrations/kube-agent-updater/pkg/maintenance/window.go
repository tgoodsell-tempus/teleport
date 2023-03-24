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
	"encoding/json"
	"fmt"
	"time"

	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
	v1 "k8s.io/api/core/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const maintenanceScheduleKeyName = "agent-maintenance-schedule"

// windowTrigger allows a maintenance to start if we are in a planned
// maintenance window. Maintenance windows are discovered by the agent and
// written to a secret (shared for all the agents). If the secret is stale or
// missing the trigger will assume the agent is not working properly and allow
// maintenance.
type windowTrigger struct {
	name string
	kclient.Client
	clock clockwork.Clock
}

// Name returns the trigger name.
func (w windowTrigger) Name() string {
	return w.name
}

// CanStart implements maintenance.Trigger and checks if we are in a
// maintenance window.
func (w windowTrigger) CanStart(ctx context.Context, object kclient.Object) (bool, error) {
	log := ctrllog.FromContext(ctx).V(1)
	secretName := fmt.Sprintf("%s-shared-state", object.GetName())
	var secret v1.Secret
	err := w.Get(ctx, kclient.ObjectKey{Namespace: object.GetNamespace(), Name: secretName}, &secret)
	if err != nil {
		return false, trace.Wrap(err)
	}
	rawData, ok := secret.Data[maintenanceScheduleKeyName]
	if !ok {
		return false, trace.Errorf("secret %s does not have key %s", secretName, maintenanceScheduleKeyName)
	}
	var maintenanceSchedule kubeScheduleRepr
	err = json.Unmarshal(rawData, &maintenanceSchedule)
	if err != nil {
		return false, trace.WrapWithMessage(err, "failed to unmarshall schedule")
	}
	now := w.clock.Now()
	if !maintenanceSchedule.isValid(now) {
		return false, trace.Errorf("maintenance schedule is stale or invalid")
	}
	for _, window := range maintenanceSchedule.Windows {
		if window.inWindow(now) {
			log.Info("maintenance window active", "start", window.Start, "end", window.Stop)
			return true, nil
		}
	}
	return false, nil
}

// Default defines what to do in case of failure. The windowTrigger should
// trigger a maintenance if it fails to evaluate the next maintenance windows.
// Not having a sane and up-to-date secret means the agent might not work as
// intended.
func (w windowTrigger) Default() bool {
	return true
}

// kubeSchedulerRepr is the structure containing the maintenance schedule
// sent by the agent through a Kubernetes secret.
type kubeScheduleRepr struct {
	Windows []windowRepr `json:"windows"`
}

// isValid checks if the schedule is valid. A schedule is considered invalid if
// it has no upcoming or ongoing maintenance window, or if it contains a window
// whose start is after its end. This could happen if the agent looses
// connectivity to its cluster or if we have a bug in the window calculation.
// In this case we don't want to honor the schedule and will consider the
// agent is not working properly.
func (s kubeScheduleRepr) isValid(now time.Time) bool {
	valid := false
	for _, window := range s.Windows {
		if window.Start.After(window.Stop) {
			return false
		}
		if window.Stop.After(now) {
			valid = true
		}
	}
	return valid
}

type windowRepr struct {
	Start time.Time `json:"start"`
	Stop  time.Time `json:"stop"`
}

// inWindow checks if a given time is in the window.
func (w windowRepr) inWindow(now time.Time) bool {
	return now.After(w.Start) && now.Before(w.Stop)
}

// NewWindowTrigger returns a new Trigger validating if the agent is within its
// maintenance window.
func NewWindowTrigger(name string, client kclient.Client) Trigger {
	return windowTrigger{
		name:   name,
		Client: client,
		clock:  clockwork.NewRealClock(),
	}
}
