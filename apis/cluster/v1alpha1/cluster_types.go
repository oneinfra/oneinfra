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

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	CertificateAuthorities CertificateAuthorities `json:"certificateAuthorities,omitempty"`
	EtcdServer             EtcdServer             `json:"etcdServer,omitempty"`
	APIServer              KubeAPIServer          `json:"apiServer,omitempty"`
	VPNCIDR                string                 `json:"vpnCIDR,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	StorageClientEndpoints []string  `json:"storageClientEndpoints,omitempty"`
	StoragePeerEndpoints   []string  `json:"storagePeerEndpoints,omitempty"`
	VPNPeers               []VPNPeer `json:"vpnPeers,omitempty"`
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
	APIServerClient   commonv1alpha1.Certificate `json:"apiServerClient,omitempty"`
	CertificateSigner commonv1alpha1.Certificate `json:"certificateSigner,omitempty"`
	Kubelet           commonv1alpha1.Certificate `json:"kubelet,omitempty"`
	EtcdClient        commonv1alpha1.Certificate `json:"etcdClient,omitempty"`
	EtcdPeer          commonv1alpha1.Certificate `json:"etcdPeer,omitempty"`
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  string `json:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey,omitempty"`
}

// KubeAPIServer represents a kube apiserver
type KubeAPIServer struct {
	CA             *commonv1alpha1.Certificate `json:"ca,omitempty"`
	TLSCert        string                      `json:"tlsCert,omitempty"`
	TLSPrivateKey  string                      `json:"tlsPrivateKey,omitempty"`
	ServiceAccount KeyPair                     `json:"serviceAccount,omitempty"`
	ExtraSANs      []string                    `json:"extraSANs,omitempty"`
}

// EtcdServer represents an etcd server
type EtcdServer struct {
	CA            *commonv1alpha1.Certificate `json:"ca,omitempty"`
	TLSCert       string                      `json:"tlsCert,omitempty"`
	TLSPrivateKey string                      `json:"tlsPrivateKey,omitempty"`
	ExtraSANs     []string                    `json:"extraSANs,omitempty"`
}

// +kubebuilder:object:root=true

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
