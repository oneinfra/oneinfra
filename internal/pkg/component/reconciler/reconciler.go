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

package reconciler

import (
	"errors"

	"k8s.io/klog/v2"

	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/component/components"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/utils"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// ComponentReconciler represents a component reconciler
type ComponentReconciler struct {
	hypervisorMap infra.HypervisorMap
	clusterMap    clusterapi.Map
	componentList componentapi.List
}

// NewComponentReconciler creates a component reconciler with the provided hypervisors, clusters and components
func NewComponentReconciler(hypervisorMap infra.HypervisorMap, clusterMap clusterapi.Map, componentList componentapi.List) *ComponentReconciler {
	return &ComponentReconciler{
		hypervisorMap: hypervisorMap,
		clusterMap:    clusterMap,
		componentList: componentList,
	}
}

// PreReconcile pre-reconciles the provided components
func (componentReconciler *ComponentReconciler) PreReconcile(componentsToPreReconcile ...*componentapi.Component) reconciler.ReconcileErrors {
	if len(componentsToPreReconcile) == 0 {
		componentsToPreReconcile = componentReconciler.componentList
	}
	reconcileErrors := reconciler.ReconcileErrors{}
	for _, component := range componentsToPreReconcile {
		klog.V(1).Infof("pre-reconciling component %q with role %q", component.Name, component.Role)
		componentToReconcile := retrieveComponent(component)
		if componentToReconcile == nil {
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				errors.New("could not retrieve a specific component instance"),
			)
			continue
		}
		err := componentToReconcile.PreReconcile(
			&reconciler.Inquirer{
				ReconciledComponent: component,
				Reconciler:          componentReconciler,
			},
		)
		if err != nil {
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				err,
			)
		}
	}
	if len(reconcileErrors) == 0 {
		return nil
	}
	return reconcileErrors
}

// Reconcile reconciles the provided components
func (componentReconciler *ComponentReconciler) Reconcile(componentsToReconcile ...*componentapi.Component) reconciler.ReconcileErrors {
	if len(componentsToReconcile) == 0 {
		componentsToReconcile = componentReconciler.componentList
	}
	reconcileErrors := reconciler.ReconcileErrors{}
	for _, component := range componentsToReconcile {
		klog.V(1).Infof("reconciling component %q with role %q", component.Name, component.Role)
		componentToReconcile := retrieveComponent(component)
		if componentToReconcile == nil {
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				errors.New("could not retrieve a specific component instance"),
			)
			continue
		}
		component.Conditions.SetCondition(
			componentapi.ReconcileStarted,
			conditions.ConditionTrue,
		)
		err := componentToReconcile.Reconcile(
			&reconciler.Inquirer{
				ReconciledComponent: component,
				Reconciler:          componentReconciler,
			},
		)
		if err == nil {
			component.Conditions.SetCondition(
				componentapi.ReconcileSucceeded,
				conditions.ConditionTrue,
			)
		} else {
			component.Conditions.SetCondition(
				componentapi.ReconcileSucceeded,
				conditions.ConditionFalse,
			)
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				err,
			)
		}
	}
	if len(reconcileErrors) == 0 {
		return nil
	}
	return reconcileErrors
}

// ReconcileDeletion reconciles the deletion of the provided components
func (componentReconciler *ComponentReconciler) ReconcileDeletion(componentsToDelete ...*componentapi.Component) reconciler.ReconcileErrors {
	reconcileErrors := reconciler.ReconcileErrors{}
	for _, component := range componentsToDelete {
		klog.V(1).Infof("reconciling component %q deletion with role %q", component.Name, component.Role)
		componentToReconcile := retrieveComponent(component)
		if componentToReconcile == nil {
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				errors.New("could not retrieve a specific component instance"),
			)
			continue
		}
		err := componentToReconcile.ReconcileDeletion(
			&reconciler.Inquirer{
				ReconciledComponent: component,
				Reconciler:          componentReconciler,
			},
		)
		if err == nil {
			component.Finalizers = utils.RemoveElementsFromList(
				component.Finalizers,
				constants.OneInfraCleanupFinalizer,
			)
		} else {
			reconcileErrors.AddComponentError(
				component.Namespace,
				component.ClusterName,
				component.Name,
				err,
			)
		}
	}
	if len(reconcileErrors) == 0 {
		return nil
	}
	return reconcileErrors
}

func retrieveComponent(component *componentapi.Component) components.Component {
	switch component.Role {
	case componentapi.ControlPlaneRole:
		return &components.ControlPlane{}
	case componentapi.ControlPlaneIngressRole:
		return &components.ControlPlaneIngress{}
	}
	klog.V(1).Infof("could not retrieve a specific component instance for component %q", component.Name)
	return nil
}
