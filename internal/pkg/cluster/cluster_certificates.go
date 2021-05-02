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
	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// InitializeCertificatesAndKeys initializes those certificates and
// keys that are not set in this cluster
func (cluster *Cluster) InitializeCertificatesAndKeys() error {
	klog.Info("generating keys and certificates")
	if err := cluster.initializeCertificateAuthorities(); err != nil {
		return err
	}
	if err := cluster.initializeEtcdServerCertificateAuthority(); err != nil {
		return err
	}
	if err := cluster.initializeAPIServerCertificateAuthority(); err != nil {
		return err
	}
	return cluster.initializeJoinKey()
}

func (cluster *Cluster) initializeCertificateAuthorities() error {
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
	if cluster.CertificateAuthorities.KubeletClient == nil {
		kubeletClientAuthority, err := certificates.NewCertificateAuthority("kubelet-client-authority")
		if err != nil {
			return err
		}
		cluster.CertificateAuthorities.KubeletClient = kubeletClientAuthority
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

func (cluster *Cluster) initializeEtcdServerCertificateAuthority() error {
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

func (cluster *Cluster) initializeAPIServerCertificateAuthority() error {
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

func (cluster *Cluster) initializeJoinKey() error {
	if cluster.JoinKey == nil {
		joinKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
		if err != nil {
			return err
		}
		cluster.JoinKey = joinKey
	}
	return nil
}
