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
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// SetupWebhookWithManager registers this web hook on the given
// manager instance
func (component *Component) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(component).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-cluster-oneinfra-ereslibre-es-v1alpha1-component,mutating=true,failurePolicy=fail,groups=cluster.oneinfra.ereslibre.es,resources=components,verbs=create;update,versions=v1alpha1,name=mcomponent.kb.io

var _ webhook.Defaulter = &Component{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (component *Component) Default() {
	klog.Info("default", "name", component.Name)
	component.addFinalizer()
	component.addClusterNameLabel()
}

func (component *Component) addFinalizer() {
	if component.Finalizers == nil {
		component.Finalizers = []string{}
	}
	component.Finalizers = append(
		component.Finalizers,
		constants.OneInfraCleanupFinalizer,
	)
}

func (component *Component) addClusterNameLabel() {
	if component.Labels == nil {
		component.Labels = map[string]string{}
	}
	component.Labels[constants.OneInfraClusterNameLabelName] = component.Spec.Cluster
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-oneinfra-ereslibre-es-v1alpha1-component,mutating=false,failurePolicy=fail,groups=cluster.oneinfra.ereslibre.es,resources=components,versions=v1alpha1,name=vcomponent.kb.io

var _ webhook.Validator = &Component{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (component *Component) ValidateCreate() error {
	klog.Info("validate create", "name", component.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (component *Component) ValidateUpdate(old runtime.Object) error {
	klog.Info("validate update", "name", component.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (component *Component) ValidateDelete() error {
	klog.Info("validate delete", "name", component.Name)
	return nil
}
