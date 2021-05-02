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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
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
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
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
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		klog.Errorf("could not get cluster %q: %v", component.ClusterName, err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if component.DeletionTimestamp == nil {
		// If the owning cluster has uninitialized certificates it's not
		// safe for us to reconcile this component yet
		if cluster.HasUninitializedCertificates() {
			return ctrl.Result{Requeue: true}, nil
		}
		err := retry.RetryOnConflict(
			retry.DefaultRetry,
			func() error {
				componentReconciler, err := newComponentReconciler(ctx, r, cluster, &r.ConnectionPool)
				if err != nil {
					klog.Errorf("could not create a component reconciler: %v", err)
					return err
				}
				component = componentReconciler.ComponentList().WithName(component.Name)
				if err := componentReconciler.PreReconcile(component); err != nil {
					return err
				}
				isDirty, err := component.IsDirty()
				if err != nil {
					return err
				}
				if isDirty {
					return reconciler.UpdateResources(ctx, componentReconciler, r)
				}
				return nil
			},
		)
		if err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
	}

	componentReconciler, err := newComponentReconciler(ctx, r, cluster, &r.ConnectionPool)
	if err != nil {
		klog.Errorf("could not create a component reconciler: %v", err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	component = componentReconciler.ComponentList().WithName(component.Name)

	res := ctrl.Result{}

	if cluster.DeletionTimestamp == nil && component.Role != componentapi.ControlPlaneIngressRole {
		ingressComponents := componentReconciler.ComponentList().WithRole(componentapi.ControlPlaneIngressRole)
		if err := componentReconciler.Reconcile(ingressComponents...); err != nil {
			res = ctrl.Result{Requeue: true}
		}
	}

	if component.DeletionTimestamp == nil {
		if err := componentReconciler.Reconcile(component); err != nil {
			klog.Errorf("failed to reconcile component %q: %v", req, err)
			res = ctrl.Result{Requeue: true}
		} else {
			isDirty, err := component.IsDirty()
			if err != nil {
				return ctrl.Result{Requeue: true}, nil
			}
			if isDirty {
				if err := r.Status().Update(ctx, component.Export()); err != nil {
					klog.Errorf("could not update component %q: %v", component.Name, err)
					res = ctrl.Result{Requeue: true}
				}
			}
		}
	} else {
		if err := componentReconciler.ReconcileDeletion(component); err != nil {
			klog.Errorf("failed to reconcile component deletion %q in cluster %q: %v", req, cluster.Name, err)
			res = ctrl.Result{Requeue: true}
		} else {
			if err := r.Update(ctx, component.Export()); err != nil {
				klog.Errorf("could not update component %q: %v", component.Name, err)
				res = ctrl.Result{Requeue: true}
			}
		}
	}

	component.RefreshCachedSpecs()

	if err := reconciler.UpdateResources(ctx, componentReconciler, r); err != nil {
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
