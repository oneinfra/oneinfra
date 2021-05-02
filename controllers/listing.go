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

package controllers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	clientapi "sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

func listHypervisors(ctx context.Context, client clientapi.Client, connectionPool *infra.HypervisorConnectionPool) (infra.HypervisorMap, error) {
	var hypervisorList infrav1alpha1.HypervisorList
	if err := client.List(ctx, &hypervisorList); err != nil {
		return infra.HypervisorMap{}, err
	}
	res := infra.HypervisorMap{}
	for _, hypervisor := range hypervisorList.Items {
		internalHypervisor, err := infra.NewHypervisorFromv1alpha1(&hypervisor, connectionPool)
		if err != nil {
			klog.Errorf("could not convert hypervisor to internal type: %v", err)
			continue
		}
		res[internalHypervisor.Name] = internalHypervisor
	}
	return res, nil
}

func listClusters(ctx context.Context, client clientapi.Client) (clusterapi.Map, error) {
	var clusterList clusterv1alpha1.ClusterList
	if err := client.List(ctx, &clusterList); err != nil {
		return clusterapi.Map{}, err
	}
	res := clusterapi.Map{}
	for _, cluster := range clusterList.Items {
		internalCluster, err := clusterapi.NewClusterFromv1alpha1(&cluster)
		if err != nil {
			continue
		}
		res[internalCluster.Name] = internalCluster
	}
	return res, nil
}

func getComponent(ctx context.Context, client clientapi.Client, req ctrl.Request) (*component.Component, error) {
	var component clusterv1alpha1.Component
	if err := client.Get(ctx, req.NamespacedName, &component); err != nil {
		return nil, err
	}
	internalComponent, err := componentapi.NewComponentFromv1alpha1(&component)
	if err != nil {
		return nil, err
	}
	return internalComponent, nil
}

func getCluster(ctx context.Context, client clientapi.Client, req ctrl.Request) (*clusterapi.Cluster, error) {
	var cluster clusterv1alpha1.Cluster
	if err := client.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return nil, err
	}
	internalCluster, err := clusterapi.NewClusterFromv1alpha1(&cluster)
	if err != nil {
		return nil, err
	}
	return internalCluster, nil
}

func listComponents(ctx context.Context, client clientapi.Client) (componentapi.List, error) {
	var componentList clusterv1alpha1.ComponentList
	if err := client.List(ctx, &componentList); err != nil {
		return componentapi.List{}, err
	}
	res := componentapi.List{}
	for _, component := range componentList.Items {
		internalComponent, err := componentapi.NewComponentFromv1alpha1(&component)
		if err != nil {
			continue
		}
		res = append(res, internalComponent)
	}
	return res, nil
}

func listClusterComponents(ctx context.Context, client clientapi.Client, clusterNamespace, clusterName string) (componentapi.List, error) {
	var componentList clusterv1alpha1.ComponentList
	err := client.List(
		ctx,
		&componentList,
		&clientapi.ListOptions{
			Raw: &metav1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.namespace=%s", clusterNamespace),
				LabelSelector: fmt.Sprintf("%s=%s", constants.OneInfraClusterNameLabelName, clusterName),
			},
		},
	)
	if err != nil {
		return componentapi.List{}, err
	}
	res := componentapi.List{}
	for _, component := range componentList.Items {
		internalComponent, err := componentapi.NewComponentFromv1alpha1(&component)
		if err != nil {
			continue
		}
		res = append(res, internalComponent)
	}
	return res, nil
}
