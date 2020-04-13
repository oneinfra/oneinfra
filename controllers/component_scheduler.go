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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// ComponentScheduler schedules an unscheduled Component object
type ComponentScheduler struct {
	client.Client
	Scheme        *runtime.Scheme
	hypervisorMap infra.HypervisorMap
}

// Reconcile schedules unscheduled component resources
func (r *ComponentScheduler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var componentList clusterv1alpha1.ComponentList
	if err := r.List(ctx, &componentList); err != nil {
		return ctrl.Result{}, err
	}

	unscheduledComponentList := []clusterv1alpha1.Component{}
	for _, component := range componentList.Items {
		if component.Spec.Hypervisor == "" {
			unscheduledComponentList = append(
				unscheduledComponentList,
				component,
			)
		}
	}

	if len(unscheduledComponentList) == 0 {
		return ctrl.Result{}, nil
	}

	var err error
	r.hypervisorMap, err = listHypervisors(ctx, r, nil)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	privateHypervisors := r.hypervisorMap.PrivateList()
	publicHypervisors := r.hypervisorMap.PublicList()

	res := ctrl.Result{}

	var hypervisorList infra.HypervisorList
	for _, versionedComponent := range unscheduledComponentList {
		component, err := component.NewComponentFromv1alpha1(&versionedComponent)
		if err != nil {
			klog.Errorf("could not convert versioned component to internal component: %v", err)
			continue
		}
		switch component.Role {
		case componentapi.ControlPlaneRole:
			hypervisorList = privateHypervisors
		case componentapi.ControlPlaneIngressRole:
			hypervisorList = publicHypervisors
		}
		scheduledHypervisor, err := hypervisorList.Sample()
		if err != nil {
			if component.Name == req.Name && component.Namespace == req.Namespace {
				res = ctrl.Result{RequeueAfter: time.Minute}
			}
			klog.Errorf("could not assign an hypervisor to component %q", component.Name)
			continue
		}
		component.HypervisorName = scheduledHypervisor.Name
		if err := r.Update(ctx, component.Export()); err != nil {
			res = ctrl.Result{Requeue: true}
			klog.Errorf("could not update component %q spec: %v", component.Name, err)
		}
	}

	return res, nil
}

// SetupWithManager sets up the component reconciler with mgr manager
func (r *ComponentScheduler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("component-scheduler").
		For(&clusterv1alpha1.Component{}).
		Complete(r)
}
