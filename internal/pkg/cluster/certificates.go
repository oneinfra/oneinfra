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
	"github.com/oneinfra/oneinfra/m/internal/pkg/certificates"
)

// CertificateAuthorities represent global certificate authorities
type CertificateAuthorities struct {
	APIServerClient   *certificates.Certificate
	CertificateSigner *certificates.Certificate
	Kubelet           *certificates.Certificate
	EtcdClient        *certificates.Certificate
	EtcdPeer          *certificates.Certificate
}

func newCertificateAuthorities() (*CertificateAuthorities, error) {
	apiserverClientAuthority, err := certificates.NewCertificateAuthority("apiserver-client-authority")
	if err != nil {
		return nil, err
	}
	certificateSignerAuthority, err := certificates.NewCertificateAuthority("certificate-signer-authority")
	if err != nil {
		return nil, err
	}
	kubeletAuthority, err := certificates.NewCertificateAuthority("kubelet-authority")
	if err != nil {
		return nil, err
	}
	etcdClientAuthority, err := certificates.NewCertificateAuthority("etcd-client-authority")
	if err != nil {
		return nil, err
	}
	etcdPeerAuthority, err := certificates.NewCertificateAuthority("etcd-peer-authority")
	if err != nil {
		return nil, err
	}
	return &CertificateAuthorities{
		APIServerClient:   apiserverClientAuthority,
		CertificateSigner: certificateSignerAuthority,
		Kubelet:           kubeletAuthority,
		EtcdClient:        etcdClientAuthority,
		EtcdPeer:          etcdPeerAuthority,
	}, nil
}
