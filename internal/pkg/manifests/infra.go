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

package manifests

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
	yamlutils "oneinfra.ereslibre.es/m/internal/pkg/yaml"
)

// RetrieveHypervisors returns an hypervisor map from the given manifests
func RetrieveHypervisors(manifests string) infra.HypervisorMap {
	klog.V(1).Info("retrieving hypervisors from manifests")
	hypervisors := infra.HypervisorMap{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := infrav1alpha1.AddToScheme(scheme); err != nil {
		return infra.HypervisorMap{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		hypervisor := infrav1alpha1.Hypervisor{}
		if _, _, err := serializer.Decode([]byte(document), nil, &hypervisor); err != nil || hypervisor.TypeMeta.Kind != "Hypervisor" {
			continue
		}
		internalHypervisor, err := infra.NewHypervisorFromv1alpha1(&hypervisor)
		if err != nil {
			continue
		}
		hypervisors[internalHypervisor.Name] = internalHypervisor
	}
	return hypervisors
}

// RetrieveClusters returns a cluster list from the given manifests
func RetrieveClusters(manifests string) cluster.Map {
	klog.V(1).Info("retrieving clusters from manifests")
	clusters := cluster.Map{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return cluster.Map{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		clusterObj := clusterv1alpha1.Cluster{}
		if _, _, err := serializer.Decode([]byte(document), nil, &clusterObj); err != nil || clusterObj.TypeMeta.Kind != "Cluster" {
			continue
		}
		internalCluster, err := cluster.NewClusterFromv1alpha1(&clusterObj)
		if err != nil {
			continue
		}
		clusters[clusterObj.ObjectMeta.Name] = internalCluster
	}
	return clusters
}

// RetrieveNodes returns a node list from the given manifests
func RetrieveNodes(manifests string) node.List {
	klog.V(1).Info("retrieving nodes from manifests")
	nodes := node.List{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return node.List{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		nodeObj := clusterv1alpha1.Node{}
		if _, _, err := serializer.Decode([]byte(document), nil, &nodeObj); err != nil || nodeObj.TypeMeta.Kind != "Node" {
			continue
		}
		internalNode, err := node.NewNodeFromv1alpha1(&nodeObj)
		if err != nil {
			continue
		}
		nodes = append(nodes, internalNode)
	}
	return nodes
}
