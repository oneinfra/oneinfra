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

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
)

// HypervisorSpec defines the desired state of Hypervisor
type HypervisorSpec struct {
	// LocalCRIEndpoint is the unix socket where this hypervisor is
	// reachable. This is only intended for development and testing
	// purposes. On production environments RemoteCRIEndpoint should be
	// used. Either a LocalCRIEndpoint or a RemoteCRIEndpoint has to be
	// provided.
	//
	// +optional
	LocalCRIEndpoint *LocalHypervisorCRIEndpoint `json:"localCRIEndpoint,omitempty"`

	// RemoteCRIEndpoint is the TCP address where this hypervisor is
	// reachable. Either a LocalCRIEndpoint or a RemoteCRIEndpoint has
	// to be provided.
	//
	// +optional
	RemoteCRIEndpoint *RemoteHypervisorCRIEndpoint `json:"remoteCRIEndpoint,omitempty"`

	// Public hypervisors will be scheduled cluster ingress components,
	// whereas private hypervisors will be scheduled the control plane
	// components themselves.
	Public bool `json:"public"`

	// IPAddress of this hypervisor. Public hypervisors must have a
	// publicly reachable IP address.
	IPAddress string `json:"ipAddress,omitempty"`

	// PortRange is the port range to be used for allocating exposed
	// components.
	PortRange HypervisorPortRange `json:"portRange,omitempty"`
}

// FileMap is a map of file paths as keys and their sum as values
type FileMap map[string]string

// ComponentFileMap is a map of filemaps, with component as keys, and
// filemaps as values
type ComponentFileMap map[string]FileMap

// ClusterFileMap is a map of component filemaps, with clusters as
// keys, and component filemaps as values
type ClusterFileMap map[string]ComponentFileMap

// LocalHypervisorCRIEndpoint represents a local hypervisor CRI endpoint (unix socket)
type LocalHypervisorCRIEndpoint struct {
	// CRIEndpoint is the unix socket path
	CRIEndpoint string `json:"criEndpointPath,omitempty"`
}

// RemoteHypervisorCRIEndpoint represents a remote hypervisor CRI endpoint (tcp with client certificate authentication)
type RemoteHypervisorCRIEndpoint struct {
	// CRIEndpoint is the address where this CRI endpoint is listening
	CRIEndpoint string `json:"criEndpointURI,omitempty"`

	// CACertificate is the CA certificate to validate the connection
	// against
	CACertificate string `json:"caCertificate,omitempty"`

	// ClientCertificate is the client certificate that will be used to
	// authenticate requests
	ClientCertificate *commonv1alpha1.Certificate `json:"clientCertificate,omitempty"`
}

// HypervisorStatus defines the observed state of Hypervisor
type HypervisorStatus struct {
	// AllocatedPorts is a list of hypervisor allocated ports
	AllocatedPorts []HypervisorPortAllocation `json:"allocatedPorts,omitempty"`
	// FreedPorts is a list of ports available for usage, freed when
	// components have been deleted
	FreedPorts []int          `json:"freedPorts,omitempty"`
	Files      ClusterFileMap `json:"files,omitempty"`
}

// HypervisorPortRange represents a port range
type HypervisorPortRange struct {
	Low  int `json:"low,omitempty"`
	High int `json:"high,omitempty"`
}

// HypervisorPortAllocation represents a port allocation in an hypervisor
type HypervisorPortAllocation struct {
	Cluster   string `json:"cluster,omitempty"`
	Component string `json:"component,omitempty"`
	Port      int    `json:"port,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Hypervisor is the Schema for the hypervisors API
type Hypervisor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HypervisorSpec   `json:"spec,omitempty"`
	Status HypervisorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HypervisorList contains a list of Hypervisor
type HypervisorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Hypervisor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Hypervisor{}, &HypervisorList{})
}
