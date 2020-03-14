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

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"

	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
)

// RESTClientFromKubeConfig creates a rest client from a kubeconfig file
func RESTClientFromKubeConfig(kubeConfig string, groupVersion *schema.GroupVersion, scheme *runtime.Scheme) (*restclient.RESTClient, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	restConfig.APIPath = "/apis"
	restConfig.GroupVersion = groupVersion
	restConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme).WithoutConversion()
	restConfig.UserAgent = rest.DefaultKubernetesUserAgent()
	return restclient.RESTClientFor(restConfig)
}

// KubeConfigWithToken returns a kubeconfig with token auth
func KubeConfigWithToken(clusterName, apiServerEndpoint, caCertificate, token string) (string, error) {
	kubeConfig := kubeConfigObjectWithToken(clusterName, apiServerEndpoint, caCertificate, token)
	return marshalKubeConfig(clusterName, kubeConfig)
}

// AdminKubeConfig returns a kubeconfig file for the current cluster
func (cluster *Cluster) AdminKubeConfig() (string, error) {
	return cluster.KubeConfig("kubernetes-admin", []string{"system:masters"})
}

// KubeConfig returns a kube config with a client certificate with the
// given common name and organization
func (cluster *Cluster) KubeConfig(commonName string, organization []string) (string, error) {
	kubeConfig, err := cluster.kubeConfigObject(cluster.APIServerEndpoint, commonName, organization)
	if err != nil {
		return "", err
	}
	return marshalKubeConfig(cluster.Name, kubeConfig)
}

// KubeConfigWithEndpoint returns a kube config with a client
// certificate with the given common name and organization, pointing
// to the provided API endpoint
func (cluster *Cluster) KubeConfigWithEndpoint(apiServerEndpoint, commonName string, organization []string) (string, error) {
	kubeConfig, err := cluster.kubeConfigObject(apiServerEndpoint, commonName, organization)
	if err != nil {
		return "", err
	}
	return marshalKubeConfig(cluster.Name, kubeConfig)
}

// KubernetesExtensionsClient returns an extensions clientset for the current cluster
func (cluster *Cluster) KubernetesExtensionsClient() (apiextensionsclientset.Interface, error) {
	if cluster.extensionsClientSet != nil {
		return cluster.extensionsClientSet, nil
	}
	kubeConfig, err := cluster.AdminKubeConfig()
	if err != nil {
		return nil, err
	}
	restClient, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	clientSet, err := apiextensionsclientset.NewForConfig(restClient)
	if err != nil {
		return nil, err
	}
	cluster.extensionsClientSet = clientSet
	return clientSet, nil
}

// RESTClient returns a REST client for the current cluster
func (cluster *Cluster) RESTClient(groupVersion *schema.GroupVersion, scheme *runtime.Scheme) (*restclient.RESTClient, error) {
	kubeConfig, err := cluster.AdminKubeConfig()
	if err != nil {
		return nil, err
	}
	return RESTClientFromKubeConfig(kubeConfig, groupVersion, scheme)
}

// KubernetesClient returns a kubernetes clientset for the current cluster
func (cluster *Cluster) KubernetesClient() (clientset.Interface, error) {
	if cluster.clientSet != nil {
		return cluster.clientSet, nil
	}
	kubeConfig, err := cluster.AdminKubeConfig()
	if err != nil {
		return nil, err
	}
	restClient, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, err
	}
	clientSet, err := clientset.NewForConfig(restClient)
	if err != nil {
		return nil, err
	}
	cluster.clientSet = clientSet
	return clientSet, nil
}

// KubeConfigWithClientCertificate returns a kubeconfig for the current cluster using the provided client certificate
func (cluster *Cluster) KubeConfigWithClientCertificate(apiServerEndpoint string, clientCertificate *certificates.Certificate) (string, error) {
	kubeConfig := kubeConfigObjectWithClientCertificate(cluster.Name, apiServerEndpoint, cluster.APIServer.CA.Certificate, clientCertificate.Certificate, clientCertificate.PrivateKey)
	return marshalKubeConfig(cluster.Name, kubeConfig)
}

func (cluster *Cluster) kubeConfigObject(apiServerEndpoint, commonName string, organization []string) (*v1.Config, error) {
	clientCertificate, clientCertificatePrivateKey, err := cluster.CertificateAuthorities.APIServerClient.CreateCertificate(commonName, organization, []string{})
	if err != nil {
		return nil, err
	}
	return kubeConfigObjectWithClientCertificate(cluster.Name, apiServerEndpoint, cluster.APIServer.CA.Certificate, clientCertificate, clientCertificatePrivateKey), nil
}

func marshalKubeConfig(clusterName string, kubeConfig *v1.Config) (string, error) {
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, v1.SchemeGroupVersion)
	if encodedKubeConfig, err := runtime.Encode(encoder, kubeConfig); err == nil {
		return string(encodedKubeConfig), nil
	}
	return "", errors.Errorf("could not create a kubeconfig for cluster %q", clusterName)
}

func kubeConfigCommon(clusterName, apiServerEndpoint, caCertificate string) *v1.Config {
	return &v1.Config{
		Clusters: []v1.NamedCluster{
			{
				Name: clusterName,
				Cluster: v1.Cluster{
					Server:                   apiServerEndpoint,
					CertificateAuthorityData: []byte(caCertificate),
				},
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: clusterName,
				Context: v1.Context{
					Cluster:  clusterName,
					AuthInfo: clusterName,
				},
			},
		},
		CurrentContext: clusterName,
	}
}

func kubeConfigObjectWithToken(clusterName, apiServerEndpoint, caCertificate, token string) *v1.Config {
	kubeConfig := kubeConfigCommon(clusterName, apiServerEndpoint, caCertificate)
	kubeConfig.AuthInfos = []v1.NamedAuthInfo{
		{
			Name: clusterName,
			AuthInfo: v1.AuthInfo{
				Token: token,
			},
		},
	}
	return kubeConfig
}

func kubeConfigObjectWithClientCertificate(clusterName, apiServerEndpoint, caCertificate, clientCertificate, clientCertificatePrivateKey string) *v1.Config {
	kubeConfig := kubeConfigCommon(clusterName, apiServerEndpoint, caCertificate)
	kubeConfig.AuthInfos = []v1.NamedAuthInfo{
		{
			Name: clusterName,
			AuthInfo: v1.AuthInfo{
				ClientCertificateData: []byte(clientCertificate),
				ClientKeyData:         []byte(clientCertificatePrivateKey),
			},
		},
	}
	return kubeConfig
}
