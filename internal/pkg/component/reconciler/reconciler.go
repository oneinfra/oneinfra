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

package reconciler

import (
	"github.com/pkg/errors"
	"k8s.io/klog"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/component/components"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"github.com/oneinfra/oneinfra/internal/pkg/utils"
)

// PreReconcile pre-reconciles the component
func PreReconcile(inquirer inquirer.ReconcilerInquirer) error {
	klog.V(1).Infof("pre-reconciling component %q with role %q", inquirer.Component().Name, inquirer.Component().Role)
	component := retrieveComponent(inquirer)
	if component == nil {
		return errors.Errorf("could not retrieve a specific component instance for component %q", inquirer.Component().Name)
	}
	return component.PreReconcile(inquirer)
}

// Reconcile reconciles the component
func Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	klog.V(1).Infof("reconciling component %q with role %q", inquirer.Component().Name, inquirer.Component().Role)
	component := retrieveComponent(inquirer)
	if component == nil {
		return errors.Errorf("could not retrieve a specific component instance for component %q", inquirer.Component().Name)
	}
	inquirer.Component().Conditions.SetCondition(
		componentapi.ReconcileStarted,
		conditions.ConditionTrue,
	)
	res := component.Reconcile(inquirer)
	if res == nil {
		inquirer.Component().Conditions.SetCondition(
			componentapi.ReconcileSucceeded,
			conditions.ConditionTrue,
		)
	} else {
		inquirer.Component().Conditions.SetCondition(
			componentapi.ReconcileSucceeded,
			conditions.ConditionFalse,
		)
	}
	return res
}

// ReconcileDeletion reconciles the component deletion
func ReconcileDeletion(inquirer inquirer.ReconcilerInquirer) error {
	klog.V(1).Infof("reconciling component %q with role %q deletion", inquirer.Component().Name, inquirer.Component().Role)
	component := retrieveComponent(inquirer)
	if component == nil {
		return errors.Errorf("could not retrieve a specific component instance for component %q", inquirer.Component().Name)
	}
	var res error
	if inquirer.Component().HypervisorName != "" {
		res = component.ReconcileDeletion(inquirer)
	} else {
		res = nil
	}
	if res == nil {
		inquirer.Component().Finalizers = utils.RemoveElementsFromList(
			inquirer.Component().Finalizers,
			constants.OneInfraCleanupFinalizer,
		)
	}
	return res
}

func retrieveComponent(inquirer inquirer.ReconcilerInquirer) components.Component {
	switch inquirer.Component().Role {
	case componentapi.ControlPlaneRole:
		return &components.ControlPlane{}
	case componentapi.ControlPlaneIngressRole:
		return &components.ControlPlaneIngress{}
	}
	klog.V(1).Infof("could not retrieve a specific component instance for component %q", inquirer.Component().Name)
	return nil
}
