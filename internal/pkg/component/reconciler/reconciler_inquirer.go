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
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// HypervisorMap returns the hypervisor map known to this component reconciler
func (componentReconciler *ComponentReconciler) HypervisorMap() infra.HypervisorMap {
	return componentReconciler.hypervisorMap
}

// ClusterMap returns the cluster map known to this component reconciler
func (componentReconciler *ComponentReconciler) ClusterMap() clusterapi.Map {
	return componentReconciler.clusterMap
}

// ComponentList returns the component list known to this component reconciler
func (componentReconciler *ComponentReconciler) ComponentList() componentapi.List {
	return componentReconciler.componentList
}
