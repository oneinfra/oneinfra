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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// ComponentReconciler reconciles a Component object
type ComponentReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	ConnectionPool infra.HypervisorConnectionPool
}

// Reconcile reconciles the component resources
func (r *ComponentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	component, err := getComponent(ctx, r, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Errorf("could not get component %q: %v", req, err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	cluster, err := getCluster(
		ctx,
		r,
		reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: component.Namespace,
				Name:      component.ClusterName,
			},
		},
	)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.Errorf("could not get cluster %q: %v", component.ClusterName, err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	clusterReconciler, err := newClusterReconciler(ctx, r, cluster, &r.ConnectionPool)
	if err != nil {
		klog.Errorf("could not create a cluster reconciler: %v", err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	res := ctrl.Result{}

	if component.DeletionTimestamp == nil {
		// If the owning cluster has uninitialized certificates it's not
		// safe for us to reconcile this component yet
		if cluster.HasUninitializedCertificates() {
			return ctrl.Result{Requeue: true}, nil
		}
		if err := clusterReconciler.PreReconcile(); err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
		if err := clusterReconciler.UpdateResources(ctx, r); err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
		if err := clusterReconciler.Reconcile(); err != nil {
			klog.Errorf("failed to reconcile cluster %q: %v", req, err)
			res = ctrl.Result{Requeue: true}
		}
	} else {
		component := clusterReconciler.ComponentList.WithName(component.Name)
		if err := clusterReconciler.ReconcileDeletions(component); err != nil {
			klog.Errorf("failed to reconcile component deletion %q in cluster %q: %v", req, cluster.Name, err)
			res = ctrl.Result{Requeue: true}
		} else {
			if component != nil {
				if err := r.Update(ctx, component.Export()); err != nil {
					klog.Errorf("could not update component %q: %v", component.Name, err)
					res = ctrl.Result{Requeue: true}
				}
			} else {
				res = ctrl.Result{Requeue: true}
			}
		}
	}

	if err := clusterReconciler.UpdateResources(ctx, r); err != nil {
		res = ctrl.Result{Requeue: true}
	}

	return res, nil
}

// SetupWithManager sets up the component reconciler with mgr manager
func (r *ComponentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("component-controller").
		For(&clusterv1alpha1.Component{}).
		Complete(r)
}
