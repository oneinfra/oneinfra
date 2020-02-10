/*
Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>

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

package node

import (
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

const (
	kubeControllerManager = "k8s.gcr.io/kube-controller-manager:v1.17.0"
)

// KubeControllerManager represents the kube-controller-manager
type KubeControllerManager struct{}

// Reconcile reconciles the kube-controller-manager
func (kubeControllerManager *KubeControllerManager) Reconcile(hypervisor *infra.Hypervisor, cluster *cluster.Cluster) error {
	return nil
}
