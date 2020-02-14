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
	"oneinfra.ereslibre.es/m/internal/pkg/node"
	"oneinfra.ereslibre.es/m/internal/pkg/node/inquirer"
)

// Reconcile reconciles the node
func Reconcile(nodeObj *node.Node, inquirer inquirer.ReconcilerInquirer) error {
	klog.V(1).Infof("reconciling node %q with role %q", nodeObj.Name, nodeObj.Role)
	var componentObj component.Component
	switch nodeObj.Role {
	case node.ControlPlaneRole:
		componentObj = &component.ControlPlane{}
	case node.ControlPlaneIngressRole:
		componentObj = &component.ControlPlaneIngress{}
	}
	return componentObj.Reconcile(inquirer)
}
