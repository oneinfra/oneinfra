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

package cluster

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"net"

	"github.com/pkg/errors"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientset "k8s.io/client-go/kubernetes"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
)

const (
	// ReconcileStarted represents a condition type signaling whether a
	// reconcile has been started
	ReconcileStarted conditions.ConditionType = "ReconcileStarted"
	// ReconcileSucceeded represents a condition type signaling that a
	// reconcile has succeeded
	ReconcileSucceeded conditions.ConditionType = "ReconcileSucceeded"
)

// Cluster represents a cluster
type Cluster struct {
	Name                   string
	Namespace              string
	ResourceVersion        string
	Labels                 map[string]string
	Annotations            map[string]string
	Finalizers             []string
	DeletionTimestamp      *metav1.Time
	KubernetesVersion      string
	CertificateAuthorities *CertificateAuthorities
	EtcdServer             *EtcdServer
	APIServer              *KubeAPIServer
	StorageClientEndpoints []string
	StoragePeerEndpoints   []string
	VPNCIDR                *net.IPNet
	VPNPeers               VPNPeerMap
	APIServerEndpoint      string
	JoinKey                *crypto.KeyPair
	DesiredJoinTokens      []string
	CurrentJoinTokens      []string
	Conditions             conditions.ConditionList
	clientSet              clientset.Interface
	extensionsClientSet    apiextensionsclientset.Interface
	loadedContentsHash     string
}

// Map represents a map of clusters
type Map map[string]*Cluster

// NewCluster returns an internal cluster
func NewCluster(clusterName, kubernetesVersion, vpnCIDR string, apiServerExtraSANs []string) (*Cluster, error) {
	_, vpnCIDRNet, err := net.ParseCIDR(vpnCIDR)
	if err != nil {
		return nil, err
	}
	res := Cluster{
		Name:              clusterName,
		KubernetesVersion: kubernetesVersion,
		VPNCIDR:           vpnCIDRNet,
		VPNPeers:          VPNPeerMap{},
	}
	if err := res.InitializeCertificatesAndKeys(); err != nil {
		return nil, err
	}
	if len(apiServerExtraSANs) > 0 {
		res.APIServer = &KubeAPIServer{
			ExtraSANs: apiServerExtraSANs,
		}
	}
	return &res, nil
}

// NewClusterFromv1alpha1 returns a cluster based on a versioned cluster
func NewClusterFromv1alpha1(cluster *clusterv1alpha1.Cluster) (*Cluster, error) {
	joinKey, err := crypto.NewKeyPairFromv1alpha1(cluster.Spec.JoinKey)
	if err != nil {
		return nil, err
	}
	if cluster.Spec.CertificateAuthorities == nil {
		cluster.Spec.CertificateAuthorities = &clusterv1alpha1.CertificateAuthorities{}
	}
	if cluster.Spec.EtcdServer == nil {
		cluster.Spec.EtcdServer = &clusterv1alpha1.EtcdServer{}
	}
	if cluster.Spec.APIServer == nil {
		cluster.Spec.APIServer = &clusterv1alpha1.KubeAPIServer{}
	}
	kubeAPIServer, err := newKubeAPIServerFromv1alpha1(cluster.Spec.APIServer)
	if err != nil {
		return nil, err
	}
	res := Cluster{
		Name:                   cluster.Name,
		Namespace:              cluster.Namespace,
		ResourceVersion:        cluster.ResourceVersion,
		Labels:                 cluster.Labels,
		Annotations:            cluster.Annotations,
		Finalizers:             cluster.Finalizers,
		DeletionTimestamp:      cluster.DeletionTimestamp,
		KubernetesVersion:      cluster.Spec.KubernetesVersion,
		CertificateAuthorities: newCertificateAuthoritiesFromv1alpha1(cluster.Spec.CertificateAuthorities),
		EtcdServer:             newEtcdServerFromv1alpha1(cluster.Spec.EtcdServer),
		APIServer:              kubeAPIServer,
		StorageClientEndpoints: cluster.Status.StorageClientEndpoints,
		StoragePeerEndpoints:   cluster.Status.StoragePeerEndpoints,
		VPNCIDR:                newVPNCIDRFromv1alpha1(cluster.Spec.VPNCIDR),
		VPNPeers:               newVPNPeersFromv1alpha1(cluster.Status.VPNPeers),
		APIServerEndpoint:      cluster.Status.APIServerEndpoint,
		JoinKey:                joinKey,
		DesiredJoinTokens:      cluster.Spec.JoinTokens,
		CurrentJoinTokens:      cluster.Status.JoinTokens,
		Conditions:             conditions.NewConditionListFromv1alpha1(cluster.Status.Conditions),
	}
	if err := res.RefreshCachedSpecs(); err != nil {
		return nil, err
	}
	return &res, nil
}

// Export exports the cluster to a versioned cluster
func (cluster *Cluster) Export() *clusterv1alpha1.Cluster {
	return &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              cluster.Name,
			Namespace:         cluster.Namespace,
			ResourceVersion:   cluster.ResourceVersion,
			Labels:            cluster.Labels,
			Annotations:       cluster.Annotations,
			Finalizers:        cluster.Finalizers,
			DeletionTimestamp: cluster.DeletionTimestamp,
		},
		Spec: clusterv1alpha1.ClusterSpec{
			KubernetesVersion:      cluster.KubernetesVersion,
			CertificateAuthorities: cluster.CertificateAuthorities.Export(),
			EtcdServer:             cluster.EtcdServer.Export(),
			APIServer:              cluster.APIServer.Export(),
			VPNCIDR:                cluster.VPNCIDR.String(),
			JoinKey:                cluster.JoinKey.Export(),
			JoinTokens:             cluster.DesiredJoinTokens,
		},
		Status: clusterv1alpha1.ClusterStatus{
			StorageClientEndpoints: cluster.StorageClientEndpoints,
			StoragePeerEndpoints:   cluster.StoragePeerEndpoints,
			VPNPeers:               cluster.VPNPeers.Export(),
			APIServerEndpoint:      cluster.APIServerEndpoint,
			JoinTokens:             cluster.CurrentJoinTokens,
			Conditions:             cluster.Conditions.Export(),
		},
	}
}

// RefreshCachedSpecs refreshes the cached spec
func (cluster *Cluster) RefreshCachedSpecs() error {
	specs, err := cluster.Specs()
	if err != nil {
		return err
	}
	cluster.loadedContentsHash = fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return nil
}

// IsDirty returns whether this cluster is dirty compared to when it
// was loaded
func (cluster *Cluster) IsDirty() (bool, error) {
	specs, err := cluster.Specs()
	if err != nil {
		return false, err
	}
	currentContentsHash := fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return cluster.loadedContentsHash != currentContentsHash, nil
}

// Specs returns the versioned specs of this cluster
func (cluster *Cluster) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, clusterv1alpha1.GroupVersion)
	clusterObject := cluster.Export()
	if encodedCluster, err := runtime.Encode(encoder, clusterObject); err == nil {
		return string(encodedCluster), nil
	}
	return "", errors.Errorf("could not encode cluster %q", cluster.Name)
}

// GenerateVPNPeer generates a new VPN peer with name peerName
func (cluster *Cluster) GenerateVPNPeer(peerName string) (*VPNPeer, error) {
	if vpnPeer, err := cluster.VPNPeer(peerName); err == nil {
		return vpnPeer, nil
	}
	controlPlaneIngressVPNIP, err := cluster.requestVPNIP()
	if err != nil {
		return nil, err
	}
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	var ipAddressNet net.IPNet
	ipAddress := net.ParseIP(controlPlaneIngressVPNIP)
	if len(ipAddress) == net.IPv6len {
		ipAddressNet = net.IPNet{IP: ipAddress, Mask: net.CIDRMask(128, 128)}
	} else {
		ipAddressNet = net.IPNet{IP: ipAddress, Mask: net.CIDRMask(32, 32)}
	}
	vpnPeer := &VPNPeer{
		Name:       peerName,
		Address:    ipAddressNet.String(),
		PrivateKey: privateKey.String(),
		PublicKey:  privateKey.PublicKey().String(),
	}
	cluster.VPNPeers[peerName] = vpnPeer
	return vpnPeer, nil
}

// VPNPeer returns the VPN peer with the provided name
func (cluster *Cluster) VPNPeer(name string) (*VPNPeer, error) {
	if vpnPeer, exists := cluster.VPNPeers[name]; exists {
		return vpnPeer, nil
	}
	return nil, errors.Errorf("vpn peer %q not found", name)
}

// HasUninitializedCertificates returns whether this cluster has
// uninitialized certificates
func (cluster *Cluster) HasUninitializedCertificates() bool {
	_, hasUninitializedCertificates := cluster.Labels[constants.OneInfraClusterUninitializedCertificates]
	return hasUninitializedCertificates
}

// requestVPNIP requests a VPN from the VPN CIDR
func (cluster *Cluster) requestVPNIP() (string, error) {
	assignedIP := big.NewInt(int64(len(cluster.VPNPeers) + 1))
	vpnNetwork := big.NewInt(0).SetBytes(cluster.VPNCIDR.IP.To16())
	vpnAssignedIP := vpnNetwork.Add(vpnNetwork, assignedIP)
	vpnAssignedIPSlice := vpnAssignedIP.Bytes()[2:]
	if len(vpnAssignedIP.Bytes()) == net.IPv6len {
		vpnAssignedIPSlice = vpnAssignedIP.Bytes()
	}
	if !cluster.VPNCIDR.Contains(net.IP(vpnAssignedIPSlice)) {
		return "", errors.Errorf("not enough IP addresses to assign in the %q CIDR", cluster.VPNCIDR)
	}
	return net.IP(vpnAssignedIPSlice).String(), nil
}

// Specs returns the versioned specs of all clusters in this map
func (clusterMap Map) Specs() (string, error) {
	res := ""
	for _, cluster := range clusterMap {
		clusterSpec, err := cluster.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", clusterSpec)
	}
	return res, nil
}
