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

// ReconcileCertificateAuthorities reconciles certificate authorities
// in this cluster, generating those missing
func (cluster *Cluster) ReconcileCertificateAuthorities() error {
	if cluster.CertificateAuthorities == nil {
		cluster.CertificateAuthorities = &CertificateAuthorities{}
	}
	if cluster.CertificateAuthorities.APIServerClient == nil {
		apiserverClientAuthority, err := certificates.NewCertificateAuthority("apiserver-client-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.APIServerClient = apiserverClientAuthority
	}
	if cluster.CertificateAuthorities.CertificateSigner == nil {
		certificateSignerAuthority, err := certificates.NewCertificateAuthority("certificate-signer-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.CertificateSigner = certificateSignerAuthority
	}
	if cluster.CertificateAuthorities.Kubelet == nil {
		kubeletAuthority, err := certificates.NewCertificateAuthority("kubelet-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.Kubelet = kubeletAuthority
	}
	if cluster.CertificateAuthorities.EtcdClient == nil {
		etcdClientAuthority, err := certificates.NewCertificateAuthority("etcd-client-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.EtcdClient = etcdClientAuthority
	}
	if cluster.CertificateAuthorities.EtcdPeer == nil {
		etcdPeerAuthority, err := certificates.NewCertificateAuthority("etcd-peer-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.EtcdPeer = etcdPeerAuthority
	}
	return nil
}

// ReconcileEtcdServerCertificateAuthority reconciles this cluster
// etcd server authority
func (cluster *Cluster) ReconcileEtcdServerCertificateAuthority() error {
	if cluster.EtcdServer == nil {
		cluster.EtcdServer = &EtcdServer{}
	}
	if cluster.EtcdServer.CA == nil {
		etcdServerCA, err := certificates.NewCertificateAuthority("etcd-authority")
		if err != nil {
			return err
		}
		cluster.EtcdServer.CA = etcdServerCA
	}
	return nil
}

// ReconcileAPIServerCertificateAuthority reconciles this cluster API
// Server certificate authority
func (cluster *Cluster) ReconcileAPIServerCertificateAuthority() error {
	if cluster.APIServer == nil {
		cluster.APIServer = &KubeAPIServer{}
	}
	if cluster.APIServer.CA == nil {
		apiServerCA, err := certificates.NewCertificateAuthority("apiserver-authority")
		if err != nil {
			return err
		}
		cluster.APIServer.CA = apiServerCA
	}
	if cluster.APIServer.ServiceAccount == nil {
		serviceAccountKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
		if err != nil {
			return err
		}
		cluster.APIServer.ServiceAccount = serviceAccountKey
	}
	return nil
}

// ReconcileJoinKey reconciles this cluster join key
func (cluster *Cluster) ReconcileJoinKey() error {
	if cluster.JoinKey == nil {
		joinKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
		if err != nil {
			return err
		}
		cluster.JoinKey = joinKey
	}
	return nil
}
