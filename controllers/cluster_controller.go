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
	"fmt"
	"time"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterController manages Component resources from Cluster resources
type ClusterController struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles the Component resources that belong to a
// Cluster resource
func (r *ClusterController) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cluster, err := getCluster(ctx, r, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Errorf("could not get cluster %q: %v", req, err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	desiredControlPlaneReplicas := cluster.ControlPlaneReplicas
	currentClusterComponents, err := listClusterComponents(ctx, r, cluster.Namespace, cluster.Name)
	if err != nil {
		klog.Errorf("could not list cluster %q components: %v", req, err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Reconcile cluster deletion

	if cluster.DeletionTimestamp != nil {
		return r.reconcileDeletion(ctx, req, currentClusterComponents)
	}

	// Reconcile control plane components

	currentUndeletedControlPlaneReplicas := componentapi.List{}
	for _, controlPlaneReplica := range currentClusterComponents {
		if controlPlaneReplica.Role != componentapi.ControlPlaneRole {
			continue
		}
		if controlPlaneReplica.DeletionTimestamp != nil {
			continue
		}
		currentUndeletedControlPlaneReplicas = append(
			currentUndeletedControlPlaneReplicas,
			controlPlaneReplica,
		)
	}

	if desiredControlPlaneReplicas > len(currentUndeletedControlPlaneReplicas) {
		missingReplicaCount := desiredControlPlaneReplicas - len(currentUndeletedControlPlaneReplicas)
		for i := 0; i < missingReplicaCount; i++ {
			component := componentapi.NewComponent(
				cluster.Namespace,
				cluster.Name,
				fmt.Sprintf("%s-control-plane-", cluster.Name),
				componentapi.ControlPlaneRole,
			)
			if err := r.Create(ctx, component.Export()); err != nil {
				klog.Error("could not create a component")
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
		}
	} else {
		excessReplicaCount := len(currentUndeletedControlPlaneReplicas) - desiredControlPlaneReplicas
		for i := 0; i < excessReplicaCount; i++ {
			component := currentUndeletedControlPlaneReplicas[i]
			if err != nil {
				klog.Error("could not get a component sample")
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
			if err := r.Delete(ctx, component.Export()); err != nil {
				klog.Error("could not delete excess component")
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
		}
	}

	// Reconcile control plane ingress components. At the moment, one
	// control plane ingress (and only one) is allowed per cluster.

	currentUndeletedControlPlaneIngressReplicas := componentapi.List{}
	for _, controlPlaneIngressReplica := range currentClusterComponents {
		if controlPlaneIngressReplica.Role != componentapi.ControlPlaneIngressRole {
			continue
		}
		if controlPlaneIngressReplica.DeletionTimestamp != nil {
			continue
		}
		currentUndeletedControlPlaneIngressReplicas = append(
			currentUndeletedControlPlaneIngressReplicas,
			controlPlaneIngressReplica,
		)
	}

	if len(currentUndeletedControlPlaneIngressReplicas) == 0 {
		component := componentapi.NewComponent(
			cluster.Namespace,
			cluster.Name,
			fmt.Sprintf("%s-control-plane-ingress-", cluster.Name),
			componentapi.ControlPlaneIngressRole,
		)
		if err := r.Create(ctx, component.Export()); err != nil {
			klog.Error("could not create a component")
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	} else if len(currentUndeletedControlPlaneIngressReplicas) > 1 {
		excessReplicaCount := len(currentUndeletedControlPlaneIngressReplicas) - 1
		for i := 0; i < excessReplicaCount; i++ {
			component := currentUndeletedControlPlaneIngressReplicas[i]
			if err != nil {
				klog.Error("could not get a component sample")
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
			if err := r.Delete(ctx, component.Export()); err != nil {
				klog.Error("could not delete excess component")
				return ctrl.Result{RequeueAfter: time.Minute}, nil
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterController) reconcileDeletion(ctx context.Context, req ctrl.Request, components component.List) (ctrl.Result, error) {
	for _, component := range components {
		if err := r.Delete(ctx, component.Export()); err != nil {
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the cluster controller with mgr manager
func (r *ClusterController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("cluster-controller").
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}
