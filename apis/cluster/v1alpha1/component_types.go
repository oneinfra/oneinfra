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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
)

// Role defines the role of this component
type Role string

const (
	// ControlPlaneRole is the role used for a Control Plane instance
	ControlPlaneRole Role = "control-plane"

	// ControlPlaneIngressRole is the role used for Control Plane ingress
	ControlPlaneIngressRole Role = "control-plane-ingress"
)

// ComponentSpec defines the desired state of Component
type ComponentSpec struct {
	// +optional
	Hypervisor string `json:"hypervisor,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	Role       Role   `json:"role,omitempty"`
}

// ComponentStatus defines the observed state of Component
type ComponentStatus struct {
	AllocatedHostPorts []ComponentHostPortAllocation         `json:"allocatedHostPorts,omitempty"`
	ClientCertificates map[string]commonv1alpha1.Certificate `json:"clientCertificates,omitempty"`
	ServerCertificates map[string]commonv1alpha1.Certificate `json:"serverCertificates,omitempty"`
	InputEndpoints     map[string]string                     `json:"inputEndpoints,omitempty"`
	OutputEndpoints    map[string]string                     `json:"outputEndpoints,omitempty"`
	Conditions         commonv1alpha1.ConditionList          `json:"conditions,omitempty"`
}

// ComponentHostPortAllocation represents a port allocation in a component
type ComponentHostPortAllocation struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port,omitempty"`
}

// +genclient
// +genclient:noStatus
// +genclient:onlyVerbs=list,watch,get,delete,deleteCollection
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=`.spec.cluster`
// +kubebuilder:printcolumn:name="Role",type=string,JSONPath=`.spec.role`
// +kubebuilder:printcolumn:name="Hypervisor",type=string,JSONPath=`.spec.hypervisor`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Component is the Schema for the components API
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentSpec   `json:"spec,omitempty"`
	Status ComponentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComponentList contains a list of Component
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Component{}, &ComponentList{})
}
