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

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
)

// Cluster represents a cluster
type Cluster struct {
	Name                   string
	CertificateAuthorities *CertificateAuthorities
	EtcdServer             *EtcdServer
	APIServer              *KubeAPIServer
	StorageClientEndpoints []string
	StoragePeerEndpoints   []string
}

// Map represents a map of clusters
type Map map[string]*Cluster

// NewCluster returns a cluster with name clusterName
func NewCluster(clusterName string, etcdServerExtraSANs, apiServerExtraSANs []string) (*Cluster, error) {
	res := Cluster{Name: clusterName}
	if err := res.generateCertificates(etcdServerExtraSANs, apiServerExtraSANs); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewClusterFromv1alpha1 returns a cluster based on a versioned cluster
func NewClusterFromv1alpha1(cluster *clusterv1alpha1.Cluster) (*Cluster, error) {
	res := Cluster{
		Name: cluster.ObjectMeta.Name,
		CertificateAuthorities: &CertificateAuthorities{
			APIServerClient:   newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.APIServerClient),
			CertificateSigner: newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.CertificateSigner),
			Kubelet:           newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.Kubelet),
			EtcdClient:        newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.EtcdClient),
			EtcdPeer:          newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.EtcdPeer),
		},
		APIServer: &KubeAPIServer{
			CA:                       newCertificateAuthorityFromv1alpha1(cluster.Spec.APIServer.CA),
			TLSCert:                  cluster.Spec.APIServer.TLSCert,
			TLSPrivateKey:            cluster.Spec.APIServer.TLSPrivateKey,
			ServiceAccountPublicKey:  cluster.Spec.APIServer.ServiceAccount.PublicKey,
			ServiceAccountPrivateKey: cluster.Spec.APIServer.ServiceAccount.PrivateKey,
			ExtraSANs:                cluster.Spec.APIServer.ExtraSANs,
		},
		EtcdServer: &EtcdServer{
			CA:            newCertificateAuthorityFromv1alpha1(cluster.Spec.EtcdServer.CA),
			TLSCert:       cluster.Spec.EtcdServer.TLSCert,
			TLSPrivateKey: cluster.Spec.EtcdServer.TLSPrivateKey,
			ExtraSANs:     cluster.Spec.EtcdServer.ExtraSANs,
		},
		StorageClientEndpoints: cluster.Status.StorageClientEndpoints,
		StoragePeerEndpoints:   cluster.Status.StoragePeerEndpoints,
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
				APIServerClient: clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.CertificateAuthorities.APIServerClient.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.APIServerClient.PrivateKey,
				},
				CertificateSigner: clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.CertificateAuthorities.CertificateSigner.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.CertificateSigner.PrivateKey,
				},
				Kubelet: clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.CertificateAuthorities.Kubelet.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.Kubelet.PrivateKey,
				},
				EtcdClient: clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.CertificateAuthorities.EtcdClient.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.EtcdClient.PrivateKey,
				},
				EtcdPeer: clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.CertificateAuthorities.EtcdPeer.Certificate,
					PrivateKey:  cluster.CertificateAuthorities.EtcdPeer.PrivateKey,
				},
			},
			APIServer: clusterv1alpha1.KubeAPIServer{
				CA: &clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.APIServer.CA.Certificate,
					PrivateKey:  cluster.APIServer.CA.PrivateKey,
				},
				TLSCert:       cluster.APIServer.TLSCert,
				TLSPrivateKey: cluster.APIServer.TLSPrivateKey,
				ServiceAccount: clusterv1alpha1.KeyPair{
					PublicKey:  cluster.APIServer.ServiceAccountPublicKey,
					PrivateKey: cluster.APIServer.ServiceAccountPrivateKey,
				},
				ExtraSANs: cluster.APIServer.ExtraSANs,
			},
			EtcdServer: clusterv1alpha1.EtcdServer{
				CA: &clusterv1alpha1.CertificateAuthority{
					Certificate: cluster.EtcdServer.CA.Certificate,
					PrivateKey:  cluster.EtcdServer.CA.PrivateKey,
				},
				TLSCert:       cluster.EtcdServer.TLSCert,
				TLSPrivateKey: cluster.EtcdServer.TLSPrivateKey,
				ExtraSANs:     cluster.EtcdServer.ExtraSANs,
			},
		},
		Status: clusterv1alpha1.ClusterStatus{
			StorageClientEndpoints: cluster.StorageClientEndpoints,
			StoragePeerEndpoints:   cluster.StoragePeerEndpoints,
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

func (cluster *Cluster) generateCertificates(etcdServerExtraSANs []string, apiServerExtraSANs []string) error {
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
