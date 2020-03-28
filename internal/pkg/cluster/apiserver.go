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
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
)

// KubeAPIServer represents the kube-apiserver component
type KubeAPIServer struct {
	CA                       *certificates.Certificate
	ServiceAccountPublicKey  string
	ServiceAccountPrivateKey string
	ExtraSANs                []string
}

func newKubeAPIServer(apiServerExtraSANs []string) (*KubeAPIServer, error) {
	certificateAuthority, err := certificates.NewCertificateAuthority("apiserver-authority")
	if err != nil {
		return nil, err
	}
	kubeAPIServer := KubeAPIServer{
		CA:        certificateAuthority,
		ExtraSANs: apiServerExtraSANs,
	}
	serviceAccountKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
	if err != nil {
		return nil, err
	}
	kubeAPIServer.ServiceAccountPublicKey = serviceAccountKey.PublicKey
	kubeAPIServer.ServiceAccountPrivateKey = serviceAccountKey.PrivateKey
	return &kubeAPIServer, nil
}
