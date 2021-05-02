/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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

package cluster

import (
	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// KubeAPIServer represents the kube-apiserver component
type KubeAPIServer struct {
	CA             *certificates.Certificate
	ServiceAccount *crypto.KeyPair
	ExtraSANs      []string
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
	kubeAPIServer.ServiceAccount = serviceAccountKey
	return &kubeAPIServer, nil
}

func newKubeAPIServerFromv1alpha1(kubeAPIServer *clusterv1alpha1.KubeAPIServer) (*KubeAPIServer, error) {
	apiServerServiceAccountKey, err := crypto.NewKeyPairFromv1alpha1(
		kubeAPIServer.ServiceAccount,
	)
	if err != nil {
		return nil, err
	}
	return &KubeAPIServer{
		CA:             certificates.NewCertificateFromv1alpha1(kubeAPIServer.CA),
		ServiceAccount: apiServerServiceAccountKey,
		ExtraSANs:      kubeAPIServer.ExtraSANs,
	}, nil
}

// Export exports this kube-apiserver to a versioned kube-apiserver
func (kubeAPIServer *KubeAPIServer) Export() *clusterv1alpha1.KubeAPIServer {
	if kubeAPIServer == nil {
		return nil
	}
	return &clusterv1alpha1.KubeAPIServer{
		CA:             kubeAPIServer.CA.Export(),
		ServiceAccount: kubeAPIServer.ServiceAccount.Export(),
		ExtraSANs:      kubeAPIServer.ExtraSANs,
	}
}
