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

package node

import (
	"github.com/pkg/errors"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

var (
	components = []ComponentType{
		KubeAPIServerComponent,
		KubeControllerManagerComponent,
		KubeSchedulerComponent,
	}
)

type Node struct {
	Name           string
	HypervisorName string
	Hypervisor     *infra.Hypervisor
}

func NodeFromv1alpha1(node *clusterv1alpha1.Node) (*Node, error) {
	return &Node{
		Name:           node.ObjectMeta.Name,
		HypervisorName: node.Spec.HypervisorName,
	}, nil
}

func (node *Node) Component(componentType ComponentType) (Component, error) {
	switch componentType {
	case KubeAPIServerComponent:
		return &KubeAPIServer{}, nil
	case KubeControllerManagerComponent:
		return &KubeControllerManager{}, nil
	case KubeSchedulerComponent:
		return &KubeScheduler{}, nil
	default:
		return nil, errors.Errorf("unknown component: %d", componentType)
	}
}

func (node *Node) Reconcile() error {
	for _, componentType := range components {
		component, err := node.Component(componentType)
		if err != nil {
			return err
		}
		if err := component.Reconcile(node.Hypervisor); err != nil {
			return err
		}
	}
	return nil
}

func (node *Node) Export() *clusterv1alpha1.Node {
	return nil
}
