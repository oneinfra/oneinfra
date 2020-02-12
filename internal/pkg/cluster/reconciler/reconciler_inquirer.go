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

// ClusterReconcilerInquirer represents the cluster reconciler that
// allows to retrieve information about the cluster
type ClusterReconcilerInquirer struct {
	node              *node.Node
	clusterReconciler *ClusterReconciler
}

// Node returns the current node in reconciliation
func (inquirer *ClusterReconcilerInquirer) Node() *node.Node {
	return inquirer.node
}

// Hypervisor returns the current hypervisor in reconciliation
func (inquirer *ClusterReconcilerInquirer) Hypervisor() *infra.Hypervisor {
	return inquirer.NodeHypervisor(inquirer.node)
}

// Cluster returns the current cluster in reconciliation
func (inquirer *ClusterReconcilerInquirer) Cluster() *cluster.Cluster {
	return inquirer.clusterReconciler.clusterMap[inquirer.Node().ClusterName]
}

// ClusterNodes returns a list of nodes with the provided role for the
// current cluster in reconciliation
func (inquirer *ClusterReconcilerInquirer) ClusterNodes(role node.Role) node.List {
	return inquirer.clusterReconciler.nodeList.WithRole(role)
}

// NodeHypervisor returns the hypervisor where the provided node is
// located
func (inquirer *ClusterReconcilerInquirer) NodeHypervisor(node *node.Node) *infra.Hypervisor {
	return inquirer.clusterReconciler.hypervisorMap[node.HypervisorName]
}
