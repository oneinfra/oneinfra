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

package reconciler

import (
	"github.com/oneinfra/oneinfra/m/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/m/internal/pkg/component"
	"github.com/oneinfra/oneinfra/m/internal/pkg/infra"
)

// ClusterReconcilerInquirer represents the cluster reconciler that
// allows to retrieve information about the cluster
type ClusterReconcilerInquirer struct {
	component         *component.Component
	clusterReconciler *ClusterReconciler
}

// Component returns the current component in reconciliation
func (inquirer *ClusterReconcilerInquirer) Component() *component.Component {
	return inquirer.component
}

// Hypervisor returns the current hypervisor in reconciliation
func (inquirer *ClusterReconcilerInquirer) Hypervisor() *infra.Hypervisor {
	return inquirer.ComponentHypervisor(inquirer.component)
}

// Cluster returns the current cluster in reconciliation
func (inquirer *ClusterReconcilerInquirer) Cluster() *cluster.Cluster {
	return inquirer.clusterReconciler.clusterMap[inquirer.Component().ClusterName]
}

// ClusterComponents returns a list of components with the provided role for the
// current cluster in reconciliation
func (inquirer *ClusterReconcilerInquirer) ClusterComponents(role component.Role) component.List {
	return inquirer.clusterReconciler.componentList.WithCluster(inquirer.Cluster().Name).WithRole(role)
}

// ComponentHypervisor returns the hypervisor where the provided component is
// located
func (inquirer *ClusterReconcilerInquirer) ComponentHypervisor(component *component.Component) *infra.Hypervisor {
	return inquirer.clusterReconciler.hypervisorMap[component.HypervisorName]
}
