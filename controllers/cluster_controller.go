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
	"errors"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	clusterReconciler *reconciler.ClusterReconciler
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

	if err := r.refreshClusterReconciler(ctx, req); err != nil {
		klog.Errorf("could not refresh cluster reconciler: %v", err)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	if r.clusterReconciler == nil {
		return ctrl.Result{}, nil
	}

	res := ctrl.Result{}

	if err := r.clusterReconciler.Reconcile(); err != nil {
		klog.Errorf("failed to reconcile cluster %q: %v", req, err)
		res = ctrl.Result{Requeue: true}
	}

	if err := r.updateHypervisors(ctx); err != nil {
		res = ctrl.Result{Requeue: true}
	}
	if err := r.updateClusters(ctx); err != nil {
		res = ctrl.Result{Requeue: true}
	}
	if err := r.updateComponents(ctx); err != nil {
		res = ctrl.Result{Requeue: true}
	}

	return res, nil
}

func (r *ClusterReconciler) refreshClusterReconciler(ctx context.Context, req ctrl.Request) error {
	hypervisorMap, err := listHypervisors(ctx, r)
	if err != nil {
		klog.Errorf("could not list hypervisors: %v", err)
		return err
	}
	cluster, err := getCluster(ctx, r, req)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.clusterReconciler = nil
			return nil
		}
		klog.Errorf("could not get cluster %q: %v", req, err)
		return err
	}
	componentList, err := listClusterComponents(ctx, r, cluster.Name)
	if err != nil {
		klog.Errorf("could not list components: %v", err)
		return err
	}
	r.clusterReconciler = reconciler.NewClusterReconciler(
		hypervisorMap,
		clusterapi.Map{cluster.Name: cluster},
		componentList,
	)
	return nil
}

func (r *ClusterReconciler) updateHypervisors(ctx context.Context) error {
	someError := false
	for _, hypervisor := range r.clusterReconciler.HypervisorMap {
		isDirty, err := hypervisor.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if hypervisor %q is dirty", hypervisor.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, hypervisor.Export()); err != nil {
				someError = true
				klog.Errorf("could not update hypervisor %q status: %v", hypervisor.Name, err)
			}
		}
	}
	if someError {
		return errors.New("could not update all hypervisors")
	}
	return nil
}

func (r *ClusterReconciler) updateClusters(ctx context.Context) error {
	someError := false
	for _, cluster := range r.clusterReconciler.ClusterMap {
		isDirty, err := cluster.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if cluster %q is dirty", cluster.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, cluster.Export()); err != nil {
				someError = true
				klog.Errorf("could not update cluster %q status: %v", cluster.Name, err)
			}
		}
	}
	if someError {
		return errors.New("could not update all clusters")
	}
	return nil
}

func (r *ClusterReconciler) updateComponents(ctx context.Context) error {
	someError := false
	for _, component := range r.clusterReconciler.ComponentList {
		isDirty, err := component.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if component %q is dirty", component.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, component.Export()); err != nil {
				someError = true
				klog.Errorf("could not update component %q status: %v", component.Name, err)
			}
		}
	}
	if someError {
		return errors.New("could not update all components")
	}
	return nil
}

// SetupWithManager sets up the cluster reconciler with mgr manager
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.Cluster{}).
		Complete(r)
}
