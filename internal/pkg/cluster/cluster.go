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
	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

// Cluster represents a cluster
type Cluster struct {
	Name                   string
	CertificateAuthorities *CertificateAuthorities
	EtcdServer             *EtcdServer
	APIServer              *KubeAPIServer
	StorageClientEndpoints []string
	StoragePeerEndpoints   []string
	VPNCIDR                *net.IPNet
	VPNPeers               VPNPeerMap
	APIServerEndpoint      string
	JoinKey                *certificates.KeyPair
	DesiredJoinTokens      []string
	CurrentJoinTokens      []string
	clientSet              clientset.Interface
	extensionsClientSet    apiextensionsclientset.Interface
}

// Map represents a map of clusters
type Map map[string]*Cluster

// NewCluster returns a cluster with name clusterName
func NewCluster(clusterName, vpnCIDR string, etcdServerExtraSANs, apiServerExtraSANs []string) (*Cluster, error) {
	_, vpnCIDRNet, err := net.ParseCIDR(vpnCIDR)
	if err != nil {
		return nil, err
	}
	joinKey, err := certificates.NewPrivateKey(constants.DefaultKeyBitSize)
	if err != nil {
		return nil, err
	}
	res := Cluster{
		Name:     clusterName,
		VPNCIDR:  vpnCIDRNet,
		VPNPeers: VPNPeerMap{},
		JoinKey:  joinKey,
	}
	if err := res.generateCertificates(etcdServerExtraSANs, apiServerExtraSANs); err != nil {
		return nil, err
	}
	if _, err := res.GenerateVPNPeer(constants.OneInfraControlPlaneIngressVPNPeerName); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewClusterFromv1alpha1 returns a cluster based on a versioned cluster
func NewClusterFromv1alpha1(cluster *clusterv1alpha1.Cluster) (*Cluster, error) {
	joinKey, err := certificates.NewKeyPairFromv1alpha1(cluster.Spec.JoinKey)
	if err != nil {
		return nil, err
	}
	res := Cluster{
		Name: cluster.ObjectMeta.Name,
		CertificateAuthorities: &CertificateAuthorities{
			APIServerClient:   certificates.NewCertificateFromv1alpha1(&cluster.Spec.CertificateAuthorities.APIServerClient),
			CertificateSigner: certificates.NewCertificateFromv1alpha1(&cluster.Spec.CertificateAuthorities.CertificateSigner),
			Kubelet:           certificates.NewCertificateFromv1alpha1(&cluster.Spec.CertificateAuthorities.Kubelet),
			EtcdClient:        certificates.NewCertificateFromv1alpha1(&cluster.Spec.CertificateAuthorities.EtcdClient),
			EtcdPeer:          certificates.NewCertificateFromv1alpha1(&cluster.Spec.CertificateAuthorities.EtcdPeer),
		},
		APIServer: &KubeAPIServer{
			CA:                       certificates.NewCertificateFromv1alpha1(cluster.Spec.APIServer.CA),
			TLSCert:                  cluster.Spec.APIServer.TLSCert,
			TLSPrivateKey:            cluster.Spec.APIServer.TLSPrivateKey,
			ServiceAccountPublicKey:  cluster.Spec.APIServer.ServiceAccount.PublicKey,
			ServiceAccountPrivateKey: cluster.Spec.APIServer.ServiceAccount.PrivateKey,
			ExtraSANs:                cluster.Spec.APIServer.ExtraSANs,
		},
		EtcdServer: &EtcdServer{
			CA:            certificates.NewCertificateFromv1alpha1(cluster.Spec.EtcdServer.CA),
			TLSCert:       cluster.Spec.EtcdServer.TLSCert,
			TLSPrivateKey: cluster.Spec.EtcdServer.TLSPrivateKey,
			ExtraSANs:     cluster.Spec.EtcdServer.ExtraSANs,
		},
		StorageClientEndpoints: cluster.Status.StorageClientEndpoints,
		StoragePeerEndpoints:   cluster.Status.StoragePeerEndpoints,
		VPNCIDR:                newVPNCIDRFromv1alpha1(cluster.Spec.VPNCIDR),
		VPNPeers:               newVPNPeersFromv1alpha1(cluster.Status.VPNPeers),
		APIServerEndpoint:      cluster.Status.APIServerEndpoint,
		JoinKey:                joinKey,
		DesiredJoinTokens:      cluster.Spec.JoinTokens,
		CurrentJoinTokens:      cluster.Status.JoinTokens,
	}
	return &res, nil
}

// Export exports the cluster to a versioned cluster
func (cluster *Cluster) Export() *clusterv1alpha1.Cluster {
	return &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.Name,
		},
		Spec: clusterv1alpha1.ClusterSpec{
			CertificateAuthorities: clusterv1alpha1.CertificateAuthorities{
				APIServerClient: commonv1alpha1.Certificate{
					Certificate: cluster.CertificateAuthorities.APIServerClient.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.APIServerClient.PrivateKey,
				},
				CertificateSigner: commonv1alpha1.Certificate{
					Certificate: cluster.CertificateAuthorities.CertificateSigner.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.CertificateSigner.PrivateKey,
				},
				Kubelet: commonv1alpha1.Certificate{
					Certificate: cluster.CertificateAuthorities.Kubelet.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.Kubelet.PrivateKey,
				},
				EtcdClient: commonv1alpha1.Certificate{
					Certificate: cluster.CertificateAuthorities.EtcdClient.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.EtcdClient.PrivateKey,
				},
				EtcdPeer: commonv1alpha1.Certificate{
					Certificate: cluster.CertificateAuthorities.EtcdPeer.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.EtcdPeer.PrivateKey,
				},
			},
			APIServer: clusterv1alpha1.KubeAPIServer{
				CA: &commonv1alpha1.Certificate{
					Certificate: cluster.APIServer.CA.Certificate,
					PrivateKey:  cluster.APIServer.CA.PrivateKey,
				},
				TLSCert:       cluster.APIServer.TLSCert,
				TLSPrivateKey: cluster.APIServer.TLSPrivateKey,
				ServiceAccount: &commonv1alpha1.KeyPair{
					PublicKey:  cluster.APIServer.ServiceAccountPublicKey,
					PrivateKey: cluster.APIServer.ServiceAccountPrivateKey,
				},
				ExtraSANs: cluster.APIServer.ExtraSANs,
			},
			EtcdServer: clusterv1alpha1.EtcdServer{
				CA: &commonv1alpha1.Certificate{
					Certificate: cluster.EtcdServer.CA.Certificate,
					PrivateKey:  cluster.EtcdServer.CA.PrivateKey,
				},
				TLSCert:       cluster.EtcdServer.TLSCert,
				TLSPrivateKey: cluster.EtcdServer.TLSPrivateKey,
				ExtraSANs:     cluster.EtcdServer.ExtraSANs,
			},
			VPNCIDR:    cluster.VPNCIDR.String(),
			JoinKey:    cluster.JoinKey.Export(),
			JoinTokens: cluster.DesiredJoinTokens,
		},
		Status: clusterv1alpha1.ClusterStatus{
			StorageClientEndpoints: cluster.StorageClientEndpoints,
			StoragePeerEndpoints:   cluster.StoragePeerEndpoints,
			VPNPeers:               cluster.VPNPeers.Export(),
			APIServerEndpoint:      cluster.APIServerEndpoint,
			JoinTokens:             cluster.CurrentJoinTokens,
		},
	}
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

func (cluster *Cluster) generateCertificates(etcdServerExtraSANs, apiServerExtraSANs []string) error {
	certificateAuthorities, err := newCertificateAuthorities()
	if err != nil {
		return err
	}
	cluster.CertificateAuthorities = certificateAuthorities
	etcdServer, err := newEtcdServer(etcdServerExtraSANs)
	if err != nil {
		return err
	}
	cluster.EtcdServer = etcdServer
	kubeAPIServer, err := newKubeAPIServer(apiServerExtraSANs)
	if err != nil {
		return err
	}
	cluster.APIServer = kubeAPIServer
	return nil
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
