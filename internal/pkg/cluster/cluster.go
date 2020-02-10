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
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

// Cluster represents a cluster
type Cluster struct {
	name                   string
	certificateAuthorities *certificateAuthorities
	apiServer              *kubeAPIServer
	nodes                  []*node.Node
}

// Map represents a map of clusters
type Map map[string]*Cluster

// List represents a list of clusters
type List []*Cluster

// NewCluster returns a cluster with name clusterName
func NewCluster(clusterName string) (*Cluster, error) {
	res := Cluster{name: clusterName}
	if err := res.generateCertificates(); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewClusterWithNodesFromv1alpha1 returns a cluster based on a versioned cluster
func NewClusterWithNodesFromv1alpha1(cluster *clusterv1alpha1.Cluster, nodes node.List) (*Cluster, error) {
	res := Cluster{
		name: cluster.ObjectMeta.Name,
		certificateAuthorities: &certificateAuthorities{
			apiServerClient:   newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.APIServerClient),
			certificateSigner: newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.CertificateSigner),
			kubelet:           newCertificateAuthorityFromv1alpha1(&cluster.Spec.CertificateAuthorities.Kubelet),
		},
		apiServer: &kubeAPIServer{
			ca:            newCertificateAuthorityFromv1alpha1(cluster.Spec.APIServer.CA),
			tlsCert:       cluster.Spec.APIServer.TLSCert,
			tlsPrivateKey: cluster.Spec.APIServer.TLSPrivateKey,
		},
		nodes: []*node.Node{},
	}
	for _, node := range nodes {
		if node.ClusterName == res.name {
			res.nodes = append(res.nodes, node)
		}
	}
	return &res, nil
}

// Reconcile reconciles the cluster
func (cluster *Cluster) Reconcile() error {
	for _, node := range cluster.nodes {
		if err := node.Reconcile(); err != nil {
			return err
		}
	}
	return nil
}

// Export exports the cluster to a versioned cluster
func (cluster *Cluster) Export() *clusterv1alpha1.Cluster {
	return &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.name,
		},
		Spec: clusterv1alpha1.ClusterSpec{
			CertificateAuthorities: clusterv1alpha1.CertificateAuthorities{
				APIServerClient: clusterv1alpha1.CertificateAuthority{
					CACertificate: cluster.certificateAuthorities.apiServerClient.caCertificateContents,
					CAPrivateKey:  cluster.certificateAuthorities.apiServerClient.caPrivateKeyContents,
				},
				CertificateSigner: clusterv1alpha1.CertificateAuthority{
					CACertificate: cluster.certificateAuthorities.certificateSigner.caCertificateContents,
					CAPrivateKey:  cluster.certificateAuthorities.certificateSigner.caPrivateKeyContents,
				},
				Kubelet: clusterv1alpha1.CertificateAuthority{
					CACertificate: cluster.certificateAuthorities.kubelet.caCertificateContents,
					CAPrivateKey:  cluster.certificateAuthorities.kubelet.caPrivateKeyContents,
				},
			},
			APIServer: clusterv1alpha1.KubeAPIServer{
				CA: &clusterv1alpha1.CertificateAuthority{
					CACertificate: cluster.apiServer.ca.caCertificateContents,
					CAPrivateKey:  cluster.apiServer.ca.caPrivateKeyContents,
				},
				TLSCert:       cluster.apiServer.tlsCert,
				TLSPrivateKey: cluster.apiServer.tlsPrivateKey,
			},
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
	return "", errors.Errorf("could not encode cluster %q", cluster.name)
}

func (cluster *Cluster) generateCertificates() error {
	certificateAuthorities, err := newCertificateAuthorities()
	if err != nil {
		return err
	}
	cluster.certificateAuthorities = certificateAuthorities
	kubeAPIServer, err := newKubeAPIServer()
	if err != nil {
		return err
	}
	cluster.apiServer = kubeAPIServer
	return nil
}

// Specs returns the versioned specs of all nodes in this list
func (list List) Specs() (string, error) {
	res := ""
	for _, cluster := range list {
		clusterSpec, err := cluster.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", clusterSpec)
	}
	return res, nil
}

// Reconcile reconciles all clusters in this list
func (list List) Reconcile() error {
	for _, cluster := range list {
		if err := cluster.Reconcile(); err != nil {
			return err
		}
	}
	return nil
}
