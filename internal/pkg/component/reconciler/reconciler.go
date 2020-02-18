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
	"k8s.io/klog"

	"oneinfra.ereslibre.es/m/internal/pkg/component"
	"oneinfra.ereslibre.es/m/internal/pkg/component/components"
	"oneinfra.ereslibre.es/m/internal/pkg/inquirer"
)

// Reconcile reconciles the component
func Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	klog.V(1).Infof("reconciling component %q with role %q", inquirer.Component().Name, inquirer.Component().Role)
	var componentObj components.Component
	switch inquirer.Component().Role {
	case component.ControlPlaneRole:
		componentObj = &components.ControlPlane{}
	case component.ControlPlaneIngressRole:
		componentObj = &components.ControlPlaneIngress{}
	}
	return componentObj.Reconcile(inquirer)
}
