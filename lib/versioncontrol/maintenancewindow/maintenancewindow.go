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

package maintenancewindow

import (
	"context"
	"os"
	"path/filepath"

	"github.com/gravitational/teleport"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/backend"
	"github.com/gravitational/teleport/lib/backend/kubernetes"
)

const (
	// kubeSchedKey is the key under which the kube controller schedule is exported
	kubeSchedKey = "agent-maintenance-schedule"

	// unitScheduleFile is the name of the file to which the unit schedule is exported.
	unitScheduleFile = "schedule"

	// unitConfigDir is the configuration directory of the teleport-upgrade unit.
	unitConfigDir = "/etc/teleport-upgrade.d"
)

// Exporter represents a type capable of exporting the maintenance window schedule to an external
// upgrader, such as the teleport-upgrade systemd timer or the kube-updater controller.
type Exporter interface {
	// Kind gets the upgrader kind associated with this exporter.
	Kind() string
	// Sync exports the appropriate maintenance window schedule if one is present, or
	// resets/clears the maintenance window if the schedule response returns no viable scheduling
	// info.
	Sync(ctx context.Context, rsp proto.ExportMaintenanceWindowsResponse) error

	// Reset forcibly clears any previously exported maintenance window values. This should be
	// called if teleport experiences prolonged loss of auth connectivity, which may be an indicator
	// that the control plane has been upgraded s.t. this agent is no longer compatible.
	Reset(ctx context.Context) error
}

// NewExporter sets up a new exporter corresponding to the specified upgrader kind.
func NewExporter(kind string) (Exporter, error) {
	switch kind {
	case types.UpgraderKindKubeController:
		return NewKubeControllerExporter(KubeControllerExporterConfig{})
	case types.UpgraderKindSystemdUnit:
		return NewSystemdUnitExporter(SystemdUnitExporterConfig{})
	default:
		return nil, trace.BadParameter("unsupported upgrader kind: %q", kind)
	}
}

type KubeControllerExporterConfig struct {
	// Backend is an optional backend. Must be an instance of the kuberenets shared-state backend
	// if not nil.
	Backend KubernetesBackend
}

// KubernetesBackend interface for kube shared storage backend.
type KubernetesBackend interface {
	// Put puts value into backend (creates if it does not
	// exists, updates it otherwise)
	Put(ctx context.Context, i backend.Item) (*backend.Lease, error)
	// Get returns a single item or not found error
	Get(ctx context.Context, key []byte) (*backend.Item, error)
}

type kubeExporter struct {
	cfg KubeControllerExporterConfig
}

func NewKubeControllerExporter(cfg KubeControllerExporterConfig) (Exporter, error) {
	if cfg.Backend == nil {
		var err error
		cfg.Backend, err = kubernetes.NewShared()
		if err != nil {
			return nil, trace.Wrap(err)
		}
	}

	return &kubeExporter{cfg: cfg}, nil
}

func (e *kubeExporter) Kind() string {
	return types.UpgraderKindKubeController
}

func (e *kubeExporter) Sync(ctx context.Context, rsp proto.ExportMaintenanceWindowsResponse) error {
	if rsp.KubeControllerSchedule == "" {
		return e.Reset(ctx)
	}

	_, err := e.cfg.Backend.Put(ctx, backend.Item{
		Key:   []byte(kubeSchedKey),
		Value: []byte(rsp.KubeControllerSchedule),
	})

	return trace.Wrap(err)
}

func (e *kubeExporter) Reset(ctx context.Context) error {
	// kube backend doesn't support deletes right now, so just set
	// the key to empty.
	_, err := e.cfg.Backend.Put(ctx, backend.Item{
		Key:   []byte(kubeSchedKey),
		Value: []byte{},
	})

	return trace.Wrap(err)
}

type SystemdUnitExporterConfig struct {
	// ConfigDir is the directory from which the teleport-upgrade periodic loads its
	// configuration parameters. Most notably, the 'schedule' file.
	ConfigDir string
}

type systemdExporter struct {
	cfg SystemdUnitExporterConfig
}

func NewSystemdUnitExporter(cfg SystemdUnitExporterConfig) (Exporter, error) {
	if cfg.ConfigDir == "" {
		cfg.ConfigDir = unitConfigDir
	}

	return &systemdExporter{cfg: cfg}, nil
}

func (e *systemdExporter) Kind() string {
	return types.UpgraderKindSystemdUnit
}

func (e *systemdExporter) Sync(ctx context.Context, rsp proto.ExportMaintenanceWindowsResponse) error {
	if len(rsp.SystemdUnitSchedule) == 0 {
		// treat an empty schedule value as equivalent to a reset
		return e.Reset(ctx)
	}

	// ensure config dir exists
	if err := os.MkdirAll(e.cfg.ConfigDir, teleport.PrivateDirMode); err != nil {
		return trace.Wrap(err)
	}

	// export schedule file
	if err := os.WriteFile(e.scheduleFile(), []byte(rsp.SystemdUnitSchedule), teleport.FileMaskOwnerOnly); err != nil {
		return trace.Errorf("failed to write schedule file: %v", err)
	}

	return nil
}

func (e *systemdExporter) Reset(_ context.Context) error {
	if err := os.Remove(e.scheduleFile()); err != nil {
		if err == os.ErrNotExist {
			return nil
		}

		return trace.Errorf("failed to reset schedule file: %v", err)
	}

	return nil
}

func (e *systemdExporter) scheduleFile() string {
	return filepath.Join(e.cfg.ConfigDir, unitScheduleFile)
}
