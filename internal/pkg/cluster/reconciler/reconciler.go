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

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentreconciler "github.com/oneinfra/oneinfra/internal/pkg/component/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// ClusterReconciler represents a cluster reconciler
type ClusterReconciler struct {
	hypervisorMap infra.HypervisorMap
	clusterMap    cluster.Map
	componentList component.List
}

// NewClusterReconciler creates a cluster reconciler with the provided hypervisors, clusters and components
func NewClusterReconciler(hypervisorMap infra.HypervisorMap, clusterMap cluster.Map, componentList component.List) *ClusterReconciler {
	return &ClusterReconciler{
		hypervisorMap: hypervisorMap,
		clusterMap:    clusterMap,
		componentList: componentList,
	}
}

// Reconcile reconciles all components known to this cluster reconciler
func (clusterReconciler *ClusterReconciler) Reconcile() error {
	klog.V(1).Info("starting reconciliation process")
	for _, componentObj := range clusterReconciler.componentList {
		err := componentreconciler.Reconcile(
			&ClusterReconcilerInquirer{
				component:         componentObj,
				clusterReconciler: clusterReconciler,
			},
		)
		if err != nil {
			klog.Errorf("failed to reconcile component %q: %v", componentObj.Name, err)
		}
	}
	return nil
}

// Specs returns the versioned specs for all resources
func (clusterReconciler *ClusterReconciler) Specs() (string, error) {
	res := ""
	hypervisors, err := clusterReconciler.hypervisorMap.Specs()
	if err != nil {
		return "", nil
	}
	res += hypervisors
	clusters, err := clusterReconciler.clusterMap.Specs()
	if err != nil {
		return "", nil
	}
	res += clusters
	components, err := clusterReconciler.componentList.Specs()
	if err != nil {
		return "", nil
	}
	res += components
	return res, nil
}
