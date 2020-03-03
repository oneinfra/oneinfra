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
	"github.com/pkg/errors"
	"oneinfra.ereslibre.es/m/internal/pkg/certificates"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// KubeConfig returns a kubeconfig for the current cluster
func (cluster *Cluster) KubeConfig(endpoint string) (string, error) {
	kubeConfig, err := cluster.kubeConfigObject(endpoint)
	if err != nil {
		return "", err
	}
	return cluster.marshalKubeConfig(kubeConfig)
}

// KubeConfigWithClientCertificate returns a kubeconfig for the current cluster using the provided client certificate
func (cluster *Cluster) KubeConfigWithClientCertificate(endpoint string, clientCertificate *certificates.Certificate) (string, error) {
	kubeConfig := cluster.kubeConfigObjectWithCertificate(endpoint, clientCertificate.Certificate, clientCertificate.PrivateKey)
	return cluster.marshalKubeConfig(kubeConfig)
}

func (cluster *Cluster) kubeConfigObject(endpoint string) (*v1.Config, error) {
	clientCertificate, clientCertificatePrivateKey, err := cluster.CertificateAuthorities.APIServerClient.CreateCertificate("kubernetes-admin", []string{"system:masters"}, []string{})
	if err != nil {
		return nil, err
	}
	return cluster.kubeConfigObjectWithCertificate(endpoint, clientCertificate, clientCertificatePrivateKey), nil
}

func (cluster *Cluster) marshalKubeConfig(kubeConfig *v1.Config) (string, error) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
	if encodedKubeConfig, err := runtime.Encode(encoder, kubeConfig); err == nil {
		return string(encodedKubeConfig), nil
	}
	return "", errors.Errorf("could not create a kubeconfig for cluster %q", cluster.Name)
}

func (cluster *Cluster) kubeConfigObjectWithCertificate(endpoint, clientCertificate, clientCertificatePrivateKey string) *v1.Config {
	return &v1.Config{
		Clusters: []v1.NamedCluster{
			{
				Name: cluster.Name,
				Cluster: v1.Cluster{
					Server:                   endpoint,
					CertificateAuthorityData: []byte(cluster.APIServer.CA.Certificate),
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: cluster.Name,
				Context: v1.Context{
					Cluster:  cluster.Name,
					AuthInfo: cluster.Name,
				},
			},
		},
		CurrentContext: cluster.Name,
		AuthInfos: []v1.NamedAuthInfo{
			{
				Name: cluster.Name,
				AuthInfo: v1.AuthInfo{
					ClientCertificateData: []byte(clientCertificate),
					ClientKeyData:         []byte(clientCertificatePrivateKey),
				},
			},
		},
	}
}
