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

package controllers

import (
	"context"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"k8s.io/klog"
	clientapi "sigs.k8s.io/controller-runtime/pkg/client"
)

func newClusterReconciler(ctx context.Context, client clientapi.Client, cluster *cluster.Cluster, components ...*component.Component) (*reconciler.ClusterReconciler, error) {
	hypervisorMap, err := listHypervisors(ctx, client)
	if err != nil {
		klog.Errorf("could not list hypervisors: %v", err)
		return nil, err
	}
	componentList := components
	if len(components) == 0 {
		var err error
		componentList, err = listClusterComponents(ctx, client, cluster.Namespace, cluster.Name)
		if err != nil {
			klog.Errorf("could not list components: %v", err)
			return nil, err
		}
	}
	return &reconciler.ClusterReconciler{
		HypervisorMap: hypervisorMap,
		ClusterMap:    clusterapi.Map{cluster.Name: cluster},
		ComponentList: componentList,
	}, nil
}
