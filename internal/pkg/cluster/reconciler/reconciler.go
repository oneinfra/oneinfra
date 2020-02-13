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
	"k8s.io/klog"

	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

// ClusterReconciler represents a cluster reconciler
type ClusterReconciler struct {
	hypervisorMap infra.HypervisorMap
	clusterMap    cluster.Map
	nodeList      node.List
}

// NewClusterReconciler creates a cluster reconciler with the provided hypervisors, clusters and nodes
func NewClusterReconciler(hypervisorMap infra.HypervisorMap, clusterMap cluster.Map, nodeList node.List) *ClusterReconciler {
	return &ClusterReconciler{
		hypervisorMap: hypervisorMap,
		clusterMap:    clusterMap,
		nodeList:      nodeList,
	}
}

// Reconcile reconciles all nodes known to this cluster reconciler
func (clusterReconciler *ClusterReconciler) Reconcile() error {
	klog.V(1).Info("starting reconciliation process")
	clusterNodes := map[string]node.List{}
	for _, nodeObj := range clusterReconciler.nodeList {
		if _, ok := clusterReconciler.hypervisorMap[nodeObj.HypervisorName]; !ok {
			continue
		}
		if _, ok := clusterReconciler.clusterMap[nodeObj.ClusterName]; !ok {
			continue
		}
		if _, ok := clusterNodes[nodeObj.ClusterName]; !ok {
			clusterNodes[nodeObj.ClusterName] = node.List{}
		}
		clusterNodes[nodeObj.ClusterName] = append(clusterNodes[nodeObj.ClusterName], nodeObj)
	}
	for _, node := range clusterReconciler.nodeList {
		node.Reconcile(&ClusterReconcilerInquirer{
			node:              node,
			clusterReconciler: clusterReconciler,
		})
	}
	return nil
}
