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

// KubeAPIServer represents the kube-apiserver component
type KubeAPIServer struct {
	CA            *CertificateAuthority
	TLSCert       string
	TLSPrivateKey string
}

func newKubeAPIServer() (*KubeAPIServer, error) {
	// TODO: allow no CA generation, provided cert and key
	certificateAuthority, err := newCertificateAuthority()
	if err != nil {
		return nil, err
	}
	kubeAPIServer := KubeAPIServer{
		CA: certificateAuthority,
	}
	tlsCert, tlsKey, err := certificateAuthority.createCertificate()
	if err != nil {
		return nil, err
	}
	kubeAPIServer.TLSCert = tlsCert
	kubeAPIServer.TLSPrivateKey = tlsKey
	return &kubeAPIServer, nil
}
