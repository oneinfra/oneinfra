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
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

type ClusterReconciler struct {
	hypervisorMap infra.HypervisorMap
	clusterMap    cluster.Map
	nodeList      node.List
}

func NewClusterReconciler(hypervisorMap infra.HypervisorMap, clusterMap cluster.Map, nodeList node.List) *ClusterReconciler {
	return &ClusterReconciler{
		hypervisorMap: hypervisorMap,
		clusterMap:    clusterMap,
		nodeList:      nodeList,
	}
}

func (clusterReconciler *ClusterReconciler) Reconcile() error {
	for _, node := range clusterReconciler.nodeList {
		if _, ok := clusterReconciler.hypervisorMap[node.HypervisorName]; !ok {
			continue
		}
		if _, ok := clusterReconciler.clusterMap[node.ClusterName]; !ok {
			continue
		}
		node.Reconcile(
			clusterReconciler.hypervisorMap[node.HypervisorName],
			clusterReconciler.clusterMap[node.ClusterName],
		)
	}
	return nil
}
