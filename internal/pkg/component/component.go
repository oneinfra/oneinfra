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

package component

import (
	"fmt"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

// Role defines the role of this component
type Role string

const (
	// ControlPlaneRole is the role used for a Control Plane instance
	ControlPlaneRole Role = "control-plane"
	// ControlPlaneIngressRole is the role used for Control Plane ingress
	ControlPlaneIngressRole Role = "control-plane-ingress"
)

// Component represents a Control Plane component
type Component struct {
	Name               string
	Role               Role
	HypervisorName     string
	ClusterName        string
	AllocatedHostPorts map[string]int
}

// NewComponentWithRandomHypervisor creates a component with a random hypervisor from the provided hypervisorList
func NewComponentWithRandomHypervisor(clusterName, componentName string, role Role, hypervisorList infra.HypervisorList) (*Component, error) {
	hypervisor, err := hypervisorList.Sample()
	if err != nil {
		return nil, err
	}
	return &Component{
		Name:               componentName,
		HypervisorName:     hypervisor.Name,
		ClusterName:        clusterName,
		Role:               role,
		AllocatedHostPorts: map[string]int{},
	}, nil
}

// NewComponentFromv1alpha1 returns a component based on a versioned component
func NewComponentFromv1alpha1(component *clusterv1alpha1.Component) (*Component, error) {
	res := Component{
		Name:           component.ObjectMeta.Name,
		HypervisorName: component.Spec.Hypervisor,
		ClusterName:    component.Spec.Cluster,
	}
	switch component.Spec.Role {
	case clusterv1alpha1.ControlPlaneRole:
		res.Role = ControlPlaneRole
	case clusterv1alpha1.ControlPlaneIngressRole:
		res.Role = ControlPlaneIngressRole
	}
	res.AllocatedHostPorts = map[string]int{}
	for _, hostPort := range component.Status.AllocatedHostPorts {
		res.AllocatedHostPorts[hostPort.Name] = hostPort.Port
	}
	return &res, nil
}

// RequestPort requests a port on the given hypervisor
func (component *Component) RequestPort(hypervisor *infra.Hypervisor, name string) (int, error) {
	if allocatedPort, ok := component.AllocatedHostPorts[name]; ok {
		return allocatedPort, nil
	}
	allocatedPort, err := hypervisor.RequestPort(component.ClusterName, fmt.Sprintf("%s-%s", component.Name, name))
	if err != nil {
		return 0, err
	}
	component.AllocatedHostPorts[name] = allocatedPort
	return allocatedPort, nil
}

// Export exports the component to a versioned component
func (component *Component) Export() *clusterv1alpha1.Component {
	res := &clusterv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name: component.Name,
		},
		Spec: clusterv1alpha1.ComponentSpec{
			Hypervisor: component.HypervisorName,
			Cluster:    component.ClusterName,
		},
	}
	switch component.Role {
	case ControlPlaneRole:
		res.Spec.Role = clusterv1alpha1.ControlPlaneRole
	case ControlPlaneIngressRole:
		res.Spec.Role = clusterv1alpha1.ControlPlaneIngressRole
	}
	res.Status.AllocatedHostPorts = []clusterv1alpha1.ComponentHostPortAllocation{}
	for hostPortName, hostPort := range component.AllocatedHostPorts {
		res.Status.AllocatedHostPorts = append(
			res.Status.AllocatedHostPorts,
			clusterv1alpha1.ComponentHostPortAllocation{
				Name: hostPortName,
				Port: hostPort,
			},
		)
	}
	return res
}

// Specs returns the versioned specs of this component
func (component *Component) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, clusterv1alpha1.GroupVersion)
	componentObject := component.Export()
	if encodedComponent, err := runtime.Encode(encoder, componentObject); err == nil {
		return string(encodedComponent), nil
	}
	return "", errors.Errorf("could not encode component %q", component.Name)
}
