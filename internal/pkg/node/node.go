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

package node

import (
	"fmt"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

// Node represents a Control Plane node
type Node struct {
	Name              string
	HypervisorName    string
	ClusterName       string
	RequestedHostPort int
	HostPort          int
}

// List represents a list of nodes
type List []*Node

// NewNodeWithRandomHypervisor creates a node with a random hypervisor from the provided hypervisorList
func NewNodeWithRandomHypervisor(clusterName, nodeName string, hypervisorList infra.HypervisorList) (*Node, error) {
	hypervisor, err := hypervisorList.Sample()
	if err != nil {
		return nil, err
	}
	assignedPort, err := hypervisor.RequestPort(clusterName, nodeName)
	if err != nil {
		return nil, err
	}
	return &Node{
		Name:           nodeName,
		HypervisorName: hypervisor.Name,
		ClusterName:    clusterName,
		HostPort:       assignedPort,
	}, nil
}

// NewNodeFromv1alpha1 returns a node based on a versioned node
func NewNodeFromv1alpha1(node *clusterv1alpha1.Node) (*Node, error) {
	res := Node{
		Name:           node.ObjectMeta.Name,
		HypervisorName: node.Spec.Hypervisor,
		ClusterName:    node.Spec.Cluster,
		HostPort:       node.Status.HostPort,
	}
	if node.Spec.HostPort != nil {
		res.RequestedHostPort = *node.Spec.HostPort
	}
	return &res, nil
}

// Reconcile reconciles the node
func (node *Node) Reconcile(hypervisor *infra.Hypervisor, cluster *cluster.Cluster) error {
	controlPlane := ControlPlane{}
	return controlPlane.Reconcile(hypervisor, cluster, node)
}

// Export exports the node to a versioned node
func (node *Node) Export() *clusterv1alpha1.Node {
	res := &clusterv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: node.Name,
		},
		Spec: clusterv1alpha1.NodeSpec{
			Hypervisor: node.HypervisorName,
			Cluster:    node.ClusterName,
			Role:       clusterv1alpha1.ControlPlaneRole,
		},
	}
	if node.RequestedHostPort > 0 {
		res.Spec.HostPort = &node.RequestedHostPort
	}
	if node.HostPort > 0 {
		res.Status.HostPort = node.HostPort
	}
	return res
}

// Specs returns the versioned specs of this node
func (node *Node) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, clusterv1alpha1.GroupVersion)
	nodeObject := node.Export()
	if encodedNode, err := runtime.Encode(encoder, nodeObject); err == nil {
		return string(encodedNode), nil
	}
	return "", errors.Errorf("could not encode node %q", node.Name)
}

// Specs returns the versioned specs of all nodes in this list
func (list List) Specs() (string, error) {
	res := ""
	for _, node := range list {
		nodeSpec, err := node.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", nodeSpec)
	}
	return res, nil
}
