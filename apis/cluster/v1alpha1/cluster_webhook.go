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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

// SetupWebhookWithManager registers this web hook on the given
// manager instance
func (cluster *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(cluster).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-cluster-oneinfra-ereslibre-es-v1alpha1-cluster,mutating=true,failurePolicy=fail,groups=cluster.oneinfra.ereslibre.es,resources=clusters,verbs=create;update,versions=v1alpha1,name=mcluster.kb.io

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (cluster *Cluster) Default() {
	clusterlog.Info("default", "name", cluster.Name)
	cluster.defaultKubernetesVersion()
	cluster.defaultCertificateAuthorities()
	cluster.defaultEtcdServer()
	cluster.defaultAPIServer()
	cluster.defaultVPNCIDR()
	cluster.defaultJoinKey()
}

func (cluster *Cluster) defaultKubernetesVersion() {
	if cluster.Spec.KubernetesVersion == "" || cluster.Spec.KubernetesVersion == "default" {
		cluster.Spec.KubernetesVersion = constants.ReleaseData.DefaultKubernetesVersion
	}
}

func (cluster *Cluster) defaultCertificateAuthorities() {
	if cluster.Spec.CertificateAuthorities == nil {
		cluster.Spec.CertificateAuthorities = &CertificateAuthorities{}
	}
	if cluster.Spec.CertificateAuthorities.APIServerClient == nil {
		apiserverClientAuthority, err := certificates.NewCertificateAuthority("apiserver-client-authority")
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.CertificateAuthorities.APIServerClient = apiserverClientAuthority.Export()
	}
	if cluster.Spec.CertificateAuthorities.CertificateSigner == nil {
		certificateSignerAuthority, err := certificates.NewCertificateAuthority("certificate-signer-authority")
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.CertificateAuthorities.CertificateSigner = certificateSignerAuthority.Export()
	}
	if cluster.Spec.CertificateAuthorities.Kubelet == nil {
		kubeletAuthority, err := certificates.NewCertificateAuthority("kubelet-authority")
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.CertificateAuthorities.Kubelet = kubeletAuthority.Export()
	}
	if cluster.Spec.CertificateAuthorities.EtcdClient == nil {
		etcdClientAuthority, err := certificates.NewCertificateAuthority("etcd-client-authority")
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.CertificateAuthorities.EtcdClient = etcdClientAuthority.Export()
	}
	if cluster.Spec.CertificateAuthorities.EtcdPeer == nil {
		etcdPeerAuthority, err := certificates.NewCertificateAuthority("etcd-peer-authority")
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.CertificateAuthorities.EtcdPeer = etcdPeerAuthority.Export()
	}
}

func (cluster *Cluster) defaultEtcdServer() {
	if cluster.Spec.EtcdServer == nil {
	}
}

func (cluster *Cluster) defaultAPIServer() {
	if cluster.Spec.APIServer == nil {
	}
}

func (cluster *Cluster) defaultVPNCIDR() {
	if cluster.Spec.VPNCIDR == "" {
		cluster.Spec.VPNCIDR = "10.0.0.0/8"
	}
}

func (cluster *Cluster) defaultJoinKey() {
	if cluster.Spec.JoinKey == nil {
		joinKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
		if err != nil {
			klog.Error(err)
			return
		}
		cluster.Spec.JoinKey = joinKey.Export()
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-oneinfra-ereslibre-es-v1alpha1-cluster,mutating=false,failurePolicy=fail,groups=cluster.oneinfra.ereslibre.es,resources=clusters,versions=v1alpha1,name=vcluster.kb.io

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cluster *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", cluster.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cluster *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", cluster.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cluster *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", cluster.Name)
	return nil
}
