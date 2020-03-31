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

package cluster

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// ComponentReconciler reconciles a Component object
type ComponentReconciler struct {
	client.Client
	Log               logr.Logger
	Scheme            *runtime.Scheme
	clusterReconciler *reconciler.ClusterReconciler
}

// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=components,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=components/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.oneinfra.ereslibre.es,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.oneinfra.ereslibre.es,resources=hypervisors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.oneinfra.ereslibre.es,resources=hypervisors/status,verbs=get;update;patch

// Reconcile reconciles the component resources
func (r *ComponentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("component", req.NamespacedName)

	if err := r.refreshClusterReconciler(ctx); err != nil {
		klog.Errorf("could not refresh cluster reconciler: %v", err)
	}

	if err := r.scheduleComponents(ctx); err != nil {
		klog.Errorf("could not schedule some components: %v", err)
	}

	if err := r.clusterReconciler.Reconcile(); err != nil {
		klog.Errorf("failed to reconcile clusters: %v", err)
	}

	r.updateHypervisors(ctx)
	r.updateClusters(ctx)
	r.updateComponents(ctx)

	return ctrl.Result{}, nil
}

func (r *ComponentReconciler) refreshClusterReconciler(ctx context.Context) error {
	hypervisorMap, err := r.listHypervisors(ctx)
	if err != nil {
		klog.Error(err, "could not list hypervisors")
	}
	clusterMap, err := r.listClusters(ctx)
	if err != nil {
		klog.Error(err, "could not list clusters")
	}
	componentList, err := r.listComponents(ctx)
	if err != nil {
		klog.Error(err, "could not list components")
	}
	r.clusterReconciler = reconciler.NewClusterReconciler(hypervisorMap, clusterMap, componentList)
	return nil
}

func (r *ComponentReconciler) listHypervisors(ctx context.Context) (infra.HypervisorMap, error) {
	var hypervisorList infrav1alpha1.HypervisorList
	if err := r.List(ctx, &hypervisorList); err != nil {
		return infra.HypervisorMap{}, err
	}
	res := infra.HypervisorMap{}
	for _, hypervisor := range hypervisorList.Items {
		internalHypervisor, err := infra.NewHypervisorFromv1alpha1(&hypervisor)
		if err != nil {
			klog.Error(err, "could not convert hypervisor to internal type")
			continue
		}
		res[internalHypervisor.Name] = internalHypervisor
	}
	return res, nil
}

func (r *ComponentReconciler) listClusters(ctx context.Context) (clusterapi.Map, error) {
	var clusterList clusterv1alpha1.ClusterList
	if err := r.List(ctx, &clusterList); err != nil {
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

func (r *ComponentReconciler) listComponents(ctx context.Context) (componentapi.List, error) {
	var componentList clusterv1alpha1.ComponentList
	if err := r.List(ctx, &componentList); err != nil {
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

func (r *ComponentReconciler) scheduleComponents(ctx context.Context) error {
	privateHypervisors := r.clusterReconciler.HypervisorMap.PrivateList()
	publicHypervisors := r.clusterReconciler.HypervisorMap.PublicList()
	for _, component := range r.clusterReconciler.ComponentList {
		if component.HypervisorName != "" {
			continue
		}
		switch component.Role {
		case componentapi.ControlPlaneRole:
			hypervisor, err := privateHypervisors.Sample()
			if err != nil {
				klog.Errorf("could not assign a private hypervisor to component %q", component.Name)
				continue
			}
			component.HypervisorName = hypervisor.Name
		case componentapi.ControlPlaneIngressRole:
			hypervisor, err := publicHypervisors.Sample()
			if err != nil {
				klog.Errorf("could not assign a public hypervisor to component %q", component.Name)
				continue
			}
			component.HypervisorName = hypervisor.Name
		}
		if err := r.Update(ctx, component.Export()); err != nil {
			klog.Errorf("could not update component %q spec: %v", component.Name, err)
			continue
		}
		if err := component.RefreshCachedSpecs(); err != nil {
			klog.Errorf("could not refresh component %q cached specs: %v", component.Name, err)
		}
	}
	return nil
}

func (r *ComponentReconciler) updateHypervisors(ctx context.Context) {
	for _, hypervisor := range r.clusterReconciler.HypervisorMap {
		isDirty, err := hypervisor.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if hypervisor %q is dirty", hypervisor.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, hypervisor.Export()); err != nil {
				klog.Errorf("could not update hypervisor %q status: %v", hypervisor.Name, err)
			}
		}
	}
}

func (r *ComponentReconciler) updateClusters(ctx context.Context) {
	for _, cluster := range r.clusterReconciler.ClusterMap {
		isDirty, err := cluster.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if cluster %q is dirty", cluster.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, cluster.Export()); err != nil {
				klog.Errorf("could not update cluster %q status: %v", cluster.Name, err)
			}
		}
	}
}

func (r *ComponentReconciler) updateComponents(ctx context.Context) {
	for _, component := range r.clusterReconciler.ComponentList {
		isDirty, err := component.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if component %q is dirty", component.Name)
			continue
		}
		if isDirty {
			if err := r.Status().Update(ctx, component.Export()); err != nil {
				klog.Errorf("could not update component %q status: %v", component.Name, err)
			}
		}
	}
}

// SetupWithManager sets up the component reconciler with mgr manager
func (r *ComponentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha1.Component{}).
		Complete(r)
}
