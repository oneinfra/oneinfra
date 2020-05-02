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

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// The number of control plane replicas this cluster will
	// manage. One control plane replica if not provided.
	//
	// +optional
	ControlPlaneReplicas int `json:"controlPlaneReplicas,omitempty"`

	// +optional
	CertificateAuthorities *CertificateAuthorities `json:"certificateAuthorities,omitempty"`

	// +optional
	EtcdServer *EtcdServer `json:"etcdServer,omitempty"`

	// +optional
	APIServer *KubeAPIServer `json:"apiServer,omitempty"`

	// +optional
	VPN *VPN `json:"vpn,omitempty"`

	// +optional
	JoinKey *commonv1alpha1.KeyPair `json:"joinKey,omitempty"`

	// +optional
	JoinTokens []string `json:"joinTokens,omitempty"`

	// +optional
	Networking *ClusterNetworking `json:"networking,omitempty"`
}

// VPN defines the VPN configuration for this cluster
type VPN struct {
	// Whether a VPN should be set for this cluster. Disabled by default.
	//
	// +optional
	Enabled bool `json:"enabled"`

	// The VPN CIDR for this cluster. 10.0.0.0/8 by default.
	//
	// +optional
	CIDR *string `json:"CIDR,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	ClientCertificates     map[string]commonv1alpha1.Certificate `json:"clientCertificates,omitempty"`
	StorageClientEndpoints map[string]string                     `json:"storageClientEndpoints,omitempty"`
	StoragePeerEndpoints   map[string]string                     `json:"storagePeerEndpoints,omitempty"`
	VPNPeers               []VPNPeer                             `json:"vpnPeers,omitempty"`
	APIServerEndpoint      string                                `json:"apiServerEndpoint,omitempty"`
	JoinTokens             []string                              `json:"joinTokens,omitempty"`
	Conditions             commonv1alpha1.ConditionList          `json:"conditions,omitempty"`
}

// VPNPeer represents a VPN peer
type VPNPeer struct {
	Name       string `json:"name,omitempty"`
	Address    string `json:"address,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
}

// CertificateAuthorities represents a set of Certificate Authorities
type CertificateAuthorities struct {
	// +optional
	APIServerClient *commonv1alpha1.Certificate `json:"apiServerClient,omitempty"`
	// +optional
	CertificateSigner *commonv1alpha1.Certificate `json:"certificateSigner,omitempty"`
	// +optional
	Kubelet *commonv1alpha1.Certificate `json:"kubelet,omitempty"`
	// +optional
	KubeletClient *commonv1alpha1.Certificate `json:"kubeletClient,omitempty"`
	// +optional
	EtcdClient *commonv1alpha1.Certificate `json:"etcdClient,omitempty"`
	// +optional
	EtcdPeer *commonv1alpha1.Certificate `json:"etcdPeer,omitempty"`
}

// KubeAPIServer represents a kube apiserver
type KubeAPIServer struct {
	// +optional
	CA *commonv1alpha1.Certificate `json:"ca,omitempty"`
	// +optional
	ServiceAccount *commonv1alpha1.KeyPair `json:"serviceAccount,omitempty"`
	// +optional
	ExtraSANs []string `json:"extraSANs,omitempty"`
}

// EtcdServer represents an etcd server
type EtcdServer struct {
	// +optional
	CA *commonv1alpha1.Certificate `json:"ca,omitempty"`
}

// ClusterNetworking represents the cluster networking settings
type ClusterNetworking struct {
	ClusterCIDR          string `json:"clusterCIDR,omitempty"`
	ServiceCIDR          string `json:"serviceCIDR,omitempty"`
	NodeCIDRMaskSize     int    `json:"nodeCIDRMaskSize,omitempty"`
	NodeCIDRMaskSizeIPv4 int    `json:"nodeCIDRMaskSizeIPv4,omitempty"`
	NodeCIDRMaskSizeIPv6 int    `json:"nodeCIDRMaskSizeIPv6,omitempty"`
}

// +genclient
// +genclient:noStatus
// +kubebuilder:printcolumn:name="Kubernetes version",type=string,JSONPath=`.spec.kubernetesVersion`
// +kubebuilder:printcolumn:name="API server endpoint",type=string,JSONPath=`.status.apiServerEndpoint`
// +kubebuilder:printcolumn:name="VPN",type=boolean,JSONPath=`.spec.vpn.enabled`
// +kubebuilder:printcolumn:name="VPN CIDR",type=string,JSONPath=`.spec.vpn.CIDR`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
