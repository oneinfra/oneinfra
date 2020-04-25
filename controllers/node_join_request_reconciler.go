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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
)

// NodeJoinRequestReconciler reconciles node join requests on clusters
// that contain at least one join token
type NodeJoinRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles the node join requests for the given cluster
func (r *NodeJoinRequestReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cluster, err := getCluster(ctx, r, req)
	if err != nil {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if len(cluster.CurrentJoinTokens) == 0 {
		return ctrl.Result{}, nil
	}

	klog.Infof("reconciling node join requests for cluster %s/%s", cluster.Namespace, cluster.Name)

	if err := cluster.ReconcileNodeJoinRequests(); err != nil {
		klog.Errorf("failed to reconcile node join requests for cluster %s/%s", cluster.Namespace, cluster.Name)
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the node join request reconciler with mgr manager
func (r *NodeJoinRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("node-join-request-reconciler").
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}
