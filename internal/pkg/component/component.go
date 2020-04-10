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

package component

import (
	"crypto/sha1"
	"fmt"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// Role defines the role of this component
type Role string

const (
	// ControlPlaneRole is the role used for a Control Plane instance
	ControlPlaneRole Role = "control-plane"
	// ControlPlaneIngressRole is the role used for Control Plane ingress
	ControlPlaneIngressRole Role = "control-plane-ingress"
)

const (
	// ReconcileStarted represents a condition type signaling whether a
	// reconcile has been started
	ReconcileStarted conditions.ConditionType = "ReconcileStarted"
	// ReconcileSucceeded represents a condition type signaling that a
	// reconcile has succeeded
	ReconcileSucceeded conditions.ConditionType = "ReconcileSucceeded"
)

// Component represents a Control Plane component
type Component struct {
	Name               string
	Namespace          string
	ResourceVersion    string
	Labels             map[string]string
	Annotations        map[string]string
	Finalizers         []string
	DeletionTimestamp  *metav1.Time
	Role               Role
	HypervisorName     string
	ClusterName        string
	AllocatedHostPorts map[string]int
	ClientCertificates map[string]*certificates.Certificate
	ServerCertificates map[string]*certificates.Certificate
	Conditions         conditions.ConditionList
	loadedContentsHash string
}

// NewComponentWithRandomHypervisor creates a component with a random hypervisor from the provided hypervisorList
func NewComponentWithRandomHypervisor(clusterName, componentName string, role Role, hypervisorList infra.HypervisorList) (*Component, error) {
	hypervisor, err := hypervisorList.Sample()
	if err != nil {
		return nil, err
	}
	return &Component{
		Name:               componentName,
		HypervisorName:     hypervisor.Name,
		ClusterName:        clusterName,
		Role:               role,
		AllocatedHostPorts: map[string]int{},
		ClientCertificates: map[string]*certificates.Certificate{},
		ServerCertificates: map[string]*certificates.Certificate{},
	}, nil
}

// NewComponentFromv1alpha1 returns a component based on a versioned component
func NewComponentFromv1alpha1(component *clusterv1alpha1.Component) (*Component, error) {
	res := Component{
		Name:              component.Name,
		Namespace:         component.Namespace,
		ResourceVersion:   component.ResourceVersion,
		Labels:            component.Labels,
		Annotations:       component.Annotations,
		Finalizers:        component.Finalizers,
		DeletionTimestamp: component.DeletionTimestamp,
		HypervisorName:    component.Spec.Hypervisor,
		ClusterName:       component.Spec.Cluster,
		Conditions:        conditions.NewConditionListFromv1alpha1(component.Status.Conditions),
	}
	switch component.Spec.Role {
	case clusterv1alpha1.ControlPlaneRole:
		res.Role = ControlPlaneRole
	case clusterv1alpha1.ControlPlaneIngressRole:
		res.Role = ControlPlaneIngressRole
	}
	res.AllocatedHostPorts = map[string]int{}
	for _, hostPort := range component.Status.AllocatedHostPorts {
		res.AllocatedHostPorts[hostPort.Name] = hostPort.Port
	}
	res.ClientCertificates = map[string]*certificates.Certificate{}
	for clientCertificateName, clientCertificate := range component.Status.ClientCertificates {
		res.ClientCertificates[clientCertificateName] = certificates.NewCertificateFromv1alpha1(&clientCertificate)
	}
	res.ServerCertificates = map[string]*certificates.Certificate{}
	for serverCertificateName, serverCertificate := range component.Status.ServerCertificates {
		res.ServerCertificates[serverCertificateName] = certificates.NewCertificateFromv1alpha1(&serverCertificate)
	}
	if err := res.RefreshCachedSpecs(); err != nil {
		return nil, err
	}
	return &res, nil
}

// RequestPort requests a port on the given hypervisor
func (component *Component) RequestPort(hypervisor *infra.Hypervisor, name string) (int, error) {
	if allocatedPort, exists := component.AllocatedHostPorts[name]; exists {
		return allocatedPort, nil
	}
	allocatedPort, err := hypervisor.RequestPort(component.ClusterName, fmt.Sprintf("%s-%s", component.Name, name))
	if err != nil {
		return 0, err
	}
	component.AllocatedHostPorts[name] = allocatedPort
	return allocatedPort, nil
}

// ClientCertificate returns a client certificate with the given name
func (component *Component) ClientCertificate(ca *certificates.Certificate, name, commonName string, organization []string, extraSANs []string) (*certificates.Certificate, error) {
	// FIXME: not only check for existence, also that contents semantically match
	if clientCertificate, exists := component.ClientCertificates[name]; exists {
		return clientCertificate, nil
	}
	certificate, privateKey, err := ca.CreateCertificate(commonName, organization, extraSANs)
	if err != nil {
		return nil, err
	}
	clientCertificate := &certificates.Certificate{
		Certificate: certificate,
		PrivateKey:  privateKey,
	}
	component.ClientCertificates[name] = clientCertificate
	return clientCertificate, nil
}

// ServerCertificate returns a server certificate with the given name
func (component *Component) ServerCertificate(ca *certificates.Certificate, name, commonName string, organization []string, extraSANs []string) (*certificates.Certificate, error) {
	// FIXME: not only check for existence, also that contents semantically match
	if serverCertificate, exists := component.ServerCertificates[name]; exists {
		return serverCertificate, nil
	}
	certificate, privateKey, err := ca.CreateCertificate(commonName, organization, extraSANs)
	if err != nil {
		return nil, err
	}
	serverCertificate := &certificates.Certificate{
		Certificate: certificate,
		PrivateKey:  privateKey,
	}
	component.ServerCertificates[name] = serverCertificate
	return serverCertificate, nil
}

// KubeConfig returns or generates a new KubeConfig file for the given cluster
func (component *Component) KubeConfig(cluster *cluster.Cluster, apiServerEndpoint, name string) (string, error) {
	clientCertificate, err := component.ClientCertificate(
		cluster.CertificateAuthorities.APIServerClient,
		name,
		"kubernetes-admin",
		[]string{"system:masters"},
		[]string{},
	)
	if err != nil {
		return "", err
	}
	kubeConfig, err := cluster.KubeConfigWithClientCertificate(apiServerEndpoint, clientCertificate)
	if err != nil {
		return "", err
	}
	return kubeConfig, nil
}

// Export exports the component to a versioned component
func (component *Component) Export() *clusterv1alpha1.Component {
	res := &clusterv1alpha1.Component{
		ObjectMeta: metav1.ObjectMeta{
			Name:              component.Name,
			Namespace:         component.Namespace,
			ResourceVersion:   component.ResourceVersion,
			Labels:            component.Labels,
			Annotations:       component.Annotations,
			Finalizers:        component.Finalizers,
			DeletionTimestamp: component.DeletionTimestamp,
		},
		Spec: clusterv1alpha1.ComponentSpec{
			Hypervisor: component.HypervisorName,
			Cluster:    component.ClusterName,
		},
		Status: clusterv1alpha1.ComponentStatus{
			Conditions: component.Conditions.Export(),
		},
	}
	switch component.Role {
	case ControlPlaneRole:
		res.Spec.Role = clusterv1alpha1.ControlPlaneRole
	case ControlPlaneIngressRole:
		res.Spec.Role = clusterv1alpha1.ControlPlaneIngressRole
	}
	res.Status.AllocatedHostPorts = []clusterv1alpha1.ComponentHostPortAllocation{}
	for hostPortName, hostPort := range component.AllocatedHostPorts {
		res.Status.AllocatedHostPorts = append(
			res.Status.AllocatedHostPorts,
			clusterv1alpha1.ComponentHostPortAllocation{
				Name: hostPortName,
				Port: hostPort,
			},
		)
	}
	res.Status.ClientCertificates = map[string]commonv1alpha1.Certificate{}
	for clientCertificateName, clientCertificate := range component.ClientCertificates {
		res.Status.ClientCertificates[clientCertificateName] = *clientCertificate.Export()
	}
	res.Status.ServerCertificates = map[string]commonv1alpha1.Certificate{}
	for serverCertificateName, serverCertificate := range component.ServerCertificates {
		res.Status.ServerCertificates[serverCertificateName] = *serverCertificate.Export()
	}
	return res
}

// RefreshCachedSpecs refreshes the cached spec
func (component *Component) RefreshCachedSpecs() error {
	specs, err := component.Specs()
	if err != nil {
		return err
	}
	component.loadedContentsHash = fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return nil
}

// IsDirty returns whether this component is dirty compared to when it
// was loaded
func (component *Component) IsDirty() (bool, error) {
	specs, err := component.Specs()
	if err != nil {
		return false, err
	}
	currentContentsHash := fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return component.loadedContentsHash != currentContentsHash, nil
}

// Specs returns the versioned specs of this component
func (component *Component) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, clusterv1alpha1.GroupVersion)
	componentObject := component.Export()
	if encodedComponent, err := runtime.Encode(encoder, componentObject); err == nil {
		return string(encodedComponent), nil
	}
	return "", errors.Errorf("could not encode component %q", component.Name)
}
