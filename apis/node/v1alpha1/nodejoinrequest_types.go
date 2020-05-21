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

const (
	// Issued represents a join request Issued condition
	Issued commonv1alpha1.ConditionType = "Issued"
)

// NodeJoinRequestSpec defines the desired state of NodeJoinRequest
type NodeJoinRequestSpec struct {
	// Base64 encoded symmetric key, used by `oneinfra` management
	// cluster to cipher joining information. This key must be ciphered
	// with the join public key of the cluster to be joined, and encoded
	// in base64. The public join key of every managed cluster can be
	// found on the `oneinfra-join` ConfigMap present in the
	// `oneinfra-system` namespace.
	SymmetricKey string `json:"symmetricKey,omitempty"`

	// The API Server endpoint for what this join request is for. The
	// generated kubeconfig file for the kubelet will point to this
	// endpoint. If not provided, the default cluster API endpoint will
	// be used.
	//
	// +optional
	APIServerEndpoint string `json:"apiServerEndpoint,omitempty"`

	// The local node container runtime endpoint.
	// (e.g. unix:///run/containerd/containerd.sock)
	ContainerRuntimeEndpoint string `json:"containerRuntimeEndpoint,omitempty"`

	// The local node image service endpoint. It's usually the same as
	// the container runtime endpoint.
	// (e.g. unix:///run/containerd/containerd.sock)
	ImageServiceEndpoint string `json:"imageServiceEndpoint,omitempty"`

	// A list of extra Subject Alternative Names (SAN's) that will be
	// included in the Kubelet server certificate issued by oneinfra.
	//
	// +optional
	ExtraSANs []string `json:"extraSANs,omitempty"`
}

// NodeJoinRequestStatus defines the observed state of NodeJoinRequest
type NodeJoinRequestStatus struct {
	// KubernetesVersion contains the Kubernetes version of the cluster
	// this node is joining to. Filled by `oneinfra`, and ciphered using
	// the provided SymmetricKey in the request spec. Base64 encoded.
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// VPN contains the VPN information for this node join request. Nil
	// if VPN is disabled. Filled by `oneinfra`.
	VPN *VPN `json:"vpn,omitempty"`

	// KubeConfig has the kubeconfig contents that the kubelet should
	// use. Filled by `oneinfra`, and ciphered using the provided
	// SymmetricKey in the request spec. Base64 encoded.
	KubeConfig string `json:"kubeConfig,omitempty"`

	// KubeletConfig has the kubelet configuration contents that the
	// kubelet should use. Filled by `oneinfra`, and ciphered using the
	// provided SymmetricKey in the request spec. Base64 encoded.
	KubeletConfig string `json:"kubeletConfig,omitempty"`

	// KubeletServerCertificate contains the contents of the Kubelet
	// server certificate to be used. Filled by `oneinfra`, and ciphered
	// using the provided SymmetricKey in the request spec. Base64
	// encoded.
	KubeletServerCertificate string `json:"kubeletServerCertificate,omitempty"`

	// KubeletServerPrivateKey contains the contents of the Kubelet
	// server private key to be used. Filled by `oneinfra`, and ciphered
	// using the provided SymmetricKey in the request spec. Base64
	// encoded.
	KubeletServerPrivateKey string `json:"kubeletServerPrivateKey,omitempty"`

	// KubeletClientCACertificate contains the contents of the Kubelet
	// client CA certificate to be used. Filled by `oneinfra`, and ciphered
	// using the provided SymmetricKey in the request spec. Base64
	// encoded.
	KubeletClientCACertificate string `json:"kubeletClientCACertificate,omitempty"`

	// Conditions contains a list of conditions for this
	// request. `oneinfra` will set the `Issued` condition to `True`
	// when this request has all the information set, and available in
	// this `Status` object.
	Conditions commonv1alpha1.ConditionList `json:"conditions,omitempty"`
}

// VPN defines the VPN related information to a node join request.
type VPN struct {
	// CIDR for this cluster. Filled by `oneinfra`, and ciphered using
	// the provided SymmetricKey in the request spec. Base64 encoded.
	CIDR string `json:"CIDR,omitempty"`

	// Address contains the VPN address of this node. Filled by
	// `oneinfra`, and ciphered using the provided SymmetricKey in the
	// request spec. Base64 encoded.
	Address string `json:"address,omitempty"`

	// PeerPrivateKey contains the VPN peer private key. Filled by
	// `oneinfra`, and ciphered using the provided SymmetricKey in the
	// request spec. Base64 encoded.
	PeerPrivateKey string `json:"peerPrivateKey,omitempty"`

	// Endpoint contains the VPN endpoint this peer should connect
	// to. Filled by `oneinfra`, and ciphered using the provided
	// SymmetricKey in the request spec. Base64 encoded.
	Endpoint string `json:"endpoint,omitempty"`

	// EndpointPublicKey contains the VPN endpoint public key. Filled by
	// `oneinfra`, and ciphered using the provided SymmetricKey in the
	// request spec. Base64 encoded.
	EndpointPublicKey string `json:"endpointPublicKey,omitempty"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create,list,watch,get,delete,deleteCollection
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
