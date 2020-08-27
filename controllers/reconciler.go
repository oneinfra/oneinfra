/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
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

package controllers

import (
	"context"

	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	clusterreconciler "github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentreconciler "github.com/oneinfra/oneinfra/internal/pkg/component/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	clientapi "sigs.k8s.io/controller-runtime/pkg/client"
)

func newComponentReconciler(ctx context.Context, client clientapi.Client, cluster *cluster.Cluster, hypervisorConnectionPool *infra.HypervisorConnectionPool, components ...*component.Component) (*componentreconciler.ComponentReconciler, error) {
	hypervisorMap, err := listHypervisors(ctx, client, hypervisorConnectionPool)
	if err != nil {
		klog.Errorf("could not list hypervisors: %v", err)
		return nil, err
	}
	componentList, err := listClusterComponents(ctx, client, cluster.Namespace, cluster.Name)
	if err != nil {
		klog.Errorf("could not list components: %v", err)
		return nil, err
	}
	return componentreconciler.NewComponentReconciler(
		hypervisorMap,
		clusterapi.Map{cluster.Name: cluster},
		componentList,
	), nil
}

func newClusterReconciler(ctx context.Context, client clientapi.Client, cluster *clusterapi.Cluster) (*clusterreconciler.ClusterReconciler, error) {
	hypervisorMap, err := listHypervisors(ctx, client, nil)
	if err != nil {
		klog.Errorf("could not list hypervisors: %v", err)
		return nil, err
	}
	componentList, err := listClusterComponents(ctx, client, cluster.Namespace, cluster.Name)
	if err != nil {
		klog.Errorf("could not list components: %v", err)
		return nil, err
	}
	return clusterreconciler.NewClusterReconciler(
		hypervisorMap,
		clusterapi.Map{cluster.Name: cluster},
		componentList,
	), nil
}
