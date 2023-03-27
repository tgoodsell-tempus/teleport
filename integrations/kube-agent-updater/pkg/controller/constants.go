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

package controller

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	// Teleport container name in the `teleport-kube-agent` Helm chart
	teleportContainerName = "teleport"
	defaultRequeue        = 30 * time.Minute
	reconciliationTimeout = 2 * time.Minute
	kubeClientTimeout     = 1 * time.Minute
	// skipReconciliationAnnotation is inspired by the tenant-operator one
	// (from the Teleport Cloud) but namespaced under `teleport.dev`
	skipReconciliationAnnotation = "teleport.dev/skipreconcile"
)

var (
	requeueLater = ctrl.Result{
		Requeue:      true,
		RequeueAfter: defaultRequeue,
	}
	requeueNow = ctrl.Result{
		Requeue:      true,
		RequeueAfter: 0,
	}
)
