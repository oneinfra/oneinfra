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

const (
	// Issued represents a join request Issued condition
	Issued commonv1alpha1.ConditionType = "Issued"
)

// NodeJoinRequestSpec defines the desired state of NodeJoinRequest
type NodeJoinRequestSpec struct {
	SymmetricKey             string `json:"symmetricKey,omitempty"`
	APIServerEndpoint        string `json:"apiServerEndpoint,omitempty"`
	ContainerRuntimeEndpoint string `json:"containerRuntimeEndpoint,omitempty"`
	ImageServiceEndpoint     string `json:"imageServiceEndpoint,omitempty"`
}

// NodeJoinRequestStatus defines the observed state of NodeJoinRequest
type NodeJoinRequestStatus struct {
	KubernetesVersion        string                       `json:"kubernetesVersion,omitempty"`
	VPNAddress               string                       `json:"vpnAddress,omitempty"`
	VPNPeer                  string                       `json:"vpnPeer,omitempty"`
	KubeConfig               string                       `json:"kubeConfig,omitempty"`
	KubeletConfig            string                       `json:"kubeletConfig,omitempty"`
	KubeletServerCertificate string                       `json:"kubeletServerCertificate,omitempty"`
	KubeletServerPrivateKey  string                       `json:"kubeletServerPrivateKey,omitempty"`
	Conditions               commonv1alpha1.ConditionList `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// NodeJoinRequest is the Schema for the nodejoinrequests API
type NodeJoinRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeJoinRequestSpec   `json:"spec,omitempty"`
	Status NodeJoinRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeJoinRequestList contains a list of NodeJoinRequest
type NodeJoinRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeJoinRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeJoinRequest{}, &NodeJoinRequestList{})
}
