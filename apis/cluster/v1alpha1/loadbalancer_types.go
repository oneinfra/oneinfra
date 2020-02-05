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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LoadBalancerSpec defines the desired state of LoadBalancer
type LoadBalancerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of LoadBalancer. Edit LoadBalancer_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// LoadBalancerStatus defines the observed state of LoadBalancer
type LoadBalancerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// LoadBalancer is the Schema for the loadbalancers API
type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerSpec   `json:"spec,omitempty"`
	Status LoadBalancerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerList contains a list of LoadBalancer
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoadBalancer{}, &LoadBalancerList{})
}
