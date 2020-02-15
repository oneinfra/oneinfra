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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Role defines the role of this node
type Role string

const (
	// ControlPlaneRole is the role used for a Control Plane instance
	ControlPlaneRole Role = "control-plane"
	// ControlPlaneIngressRole is the role used for Control Plane ingress
	ControlPlaneIngressRole Role = "control-plane-ingress"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	Hypervisor string `json:"hypervisor,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	Role       Role   `json:"role,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	AllocatedHostPorts []NodeHostPortAllocation `json:"allocatedHostPorts,omitempty"`
}

// NodeHostPortAllocation represents a port allocation in a node
type NodeHostPortAllocation struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port,omitempty"`
}

// +kubebuilder:object:root=true

// Node is the Schema for the nodes API
type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSpec   `json:"spec,omitempty"`
	Status NodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeList contains a list of Node
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Node `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
