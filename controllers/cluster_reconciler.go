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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	clusterreconciler "github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=components,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=components/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.oneinfra.ereslibre.es,resources=hypervisors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.oneinfra.ereslibre.es,resources=hypervisors/status,verbs=get;update;patch

// Reconcile reconciles the cluster resources
func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cluster, err := getCluster(ctx, r, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Errorf("could not get cluster %q: %v", req, err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	clusterReconciler, err := newClusterReconciler(ctx, r, cluster)
	if err != nil {
		klog.Errorf("could not create a cluster reconciler: %v", err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	clusterMap := clusterReconciler.ClusterMap()
	cluster = clusterMap[cluster.Name]

	res := ctrl.Result{}

	if cluster.DeletionTimestamp == nil {
		if err := clusterReconciler.Reconcile(clusterreconciler.OptionalReconcile{}, cluster); err != nil {
			klog.Errorf("failed to reconcile cluster %q: %v", req, err)
			res = ctrl.Result{Requeue: true}
		} else {
			if err := r.Status().Update(ctx, cluster.Export()); err != nil {
				klog.Errorf("could not update cluster %q: %v", cluster.Name, err)
				res = ctrl.Result{Requeue: true}
			}
		}
	} else {
		if err := clusterReconciler.ReconcileDeletion(cluster); err != nil {
			klog.Errorf("failed to reconcile cluster %q deletion: %v", req, err)
			res = ctrl.Result{Requeue: true}
		} else {
			if cluster != nil {
				if err := r.Update(ctx, cluster.Export()); err != nil {
					klog.Errorf("could not update cluster %q: %v", cluster.Name, err)
					res = ctrl.Result{Requeue: true}
				}
			} else {
				res = ctrl.Result{Requeue: true}
			}
		}
	}

	cluster.RefreshCachedSpecs()

	if err := reconciler.UpdateResources(ctx, clusterReconciler, r); err != nil {
		res = ctrl.Result{Requeue: true}
	}

	return res, nil
}

// SetupWithManager sets up the cluster reconciler with mgr manager
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("cluster-reconciler").
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}
