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

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ClusterInitializer initializes a Cluster object
type ClusterInitializer struct {
	client.Client
	Scheme            *runtime.Scheme
	clusterReconciler *reconciler.ClusterReconciler
}

// Reconcile initializes the cluster resources
func (r *ClusterInitializer) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cluster, err := getCluster(ctx, r, req)
	if err != nil {
		return ctrl.Result{Requeue: true}, nil
	}

	if !cluster.HasUninitializedCertificates() {
		return ctrl.Result{}, nil
	}

	if err := cluster.InitializeCertificatesAndKeys(); err != nil {
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	delete(cluster.Labels, constants.OneInfraClusterUninitializedCertificates)

	if err := r.Update(ctx, cluster.Export()); err != nil {
		klog.Errorf("could not update cluster %q spec: %v", cluster.Name, err)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the cluster initializer with mgr manager
func (r *ClusterInitializer) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("cluster-initializer").
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}
