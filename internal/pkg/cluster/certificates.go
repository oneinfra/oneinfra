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
)

// CertificateAuthorities represent global certificate authorities
type CertificateAuthorities struct {
	APIServerClient   *certificates.Certificate
	CertificateSigner *certificates.Certificate
	Kubelet           *certificates.Certificate
	KubeletClient     *certificates.Certificate
	EtcdClient        *certificates.Certificate
	EtcdPeer          *certificates.Certificate
}

func newCertificateAuthoritiesFromv1alpha1(certificateAuthorities *clusterv1alpha1.CertificateAuthorities) *CertificateAuthorities {
	return &CertificateAuthorities{
		APIServerClient:   certificates.NewCertificateFromv1alpha1(certificateAuthorities.APIServerClient),
		CertificateSigner: certificates.NewCertificateFromv1alpha1(certificateAuthorities.CertificateSigner),
		Kubelet:           certificates.NewCertificateFromv1alpha1(certificateAuthorities.Kubelet),
		KubeletClient:     certificates.NewCertificateFromv1alpha1(certificateAuthorities.KubeletClient),
		EtcdClient:        certificates.NewCertificateFromv1alpha1(certificateAuthorities.EtcdClient),
		EtcdPeer:          certificates.NewCertificateFromv1alpha1(certificateAuthorities.EtcdPeer),
	}
}

// Export exports these set of certificate authorities to a versioned certificate authority set
func (certificateAuthorities *CertificateAuthorities) Export() *clusterv1alpha1.CertificateAuthorities {
	if certificateAuthorities == nil {
		return nil
	}
	return &clusterv1alpha1.CertificateAuthorities{
		APIServerClient:   certificateAuthorities.APIServerClient.Export(),
		CertificateSigner: certificateAuthorities.CertificateSigner.Export(),
		Kubelet:           certificateAuthorities.Kubelet.Export(),
		KubeletClient:     certificateAuthorities.KubeletClient.Export(),
		EtcdClient:        certificateAuthorities.EtcdClient.Export(),
		EtcdPeer:          certificateAuthorities.EtcdPeer.Export(),
	}
}
