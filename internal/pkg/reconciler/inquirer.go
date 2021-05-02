/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package reconciler

import (
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// Inquirer allows to retrieve information about the cluster
type Inquirer struct {
	ReconciledComponent *component.Component
	Reconciler          Reconciler
}

// Component returns the current component in reconciliation
func (inquirer *Inquirer) Component() *component.Component {
	return inquirer.ReconciledComponent
}

// Hypervisor returns the current hypervisor in reconciliation
func (inquirer *Inquirer) Hypervisor() *infra.Hypervisor {
	return inquirer.ComponentHypervisor(inquirer.ReconciledComponent)
}

// Cluster returns the current cluster in reconciliation
func (inquirer *Inquirer) Cluster() *cluster.Cluster {
	clusterMap := inquirer.Reconciler.ClusterMap()
	return clusterMap[inquirer.Component().ClusterName]
}

// ClusterComponents returns a list of components with the provided role for the
// current cluster in reconciliation
func (inquirer *Inquirer) ClusterComponents(role component.Role) component.List {
	componentList := inquirer.Reconciler.ComponentList()
	return componentList.WithCluster(
		inquirer.Cluster().Namespace,
		inquirer.Cluster().Name,
	).WithRole(role)
}

// ComponentHypervisor returns the hypervisor where the provided component is
// located
func (inquirer *Inquirer) ComponentHypervisor(component *component.Component) *infra.Hypervisor {
	hypervisorMap := inquirer.Reconciler.HypervisorMap()
	return hypervisorMap[component.HypervisorName]
}
