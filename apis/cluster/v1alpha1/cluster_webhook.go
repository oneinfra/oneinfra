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
	"github.com/moby/moby/pkg/namesgenerator"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
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
	cluster.defaultVPNCIDR()
	cluster.defaultJoinChallenge()
}

func (cluster *Cluster) defaultKubernetesVersion() {
	if cluster.Spec.KubernetesVersion == "" || cluster.Spec.KubernetesVersion == "default" {
		cluster.Spec.KubernetesVersion = constants.ReleaseData.DefaultKubernetesVersion
	}
}

func (cluster *Cluster) defaultVPNCIDR() {
	if cluster.Spec.VPNCIDR == "" {
		cluster.Spec.VPNCIDR = "10.0.0.0/8"
	}
}

func (cluster *Cluster) defaultJoinChallenge() {
	if cluster.Spec.JoinChallenge == "" {
		cluster.Spec.JoinChallenge = namesgenerator.GetRandomName(0)
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
