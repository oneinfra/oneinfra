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

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
	yamlutils "oneinfra.ereslibre.es/m/internal/pkg/yaml"
)

func RetrieveHypervisors(manifests string) infra.HypervisorMap {
	hypervisors := infra.HypervisorMap{}
	documents := yamlutils.SplitDocuments(manifests)
	for _, document := range documents {
		scheme := runtime.NewScheme()
		if err := infrav1alpha1.AddToScheme(scheme); err != nil {
			continue
		}
		hypervisor := infrav1alpha1.Hypervisor{}
		gvk := infrav1alpha1.GroupVersion.WithKind("Hypervisor")
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
		if _, _, err := serializer.Decode([]byte(document), &gvk, &hypervisor); err != nil {
			continue
		}
		internalHypervisor, err := infra.HypervisorFromv1alpha1(&hypervisor)
		if err != nil {
			continue
		}
		hypervisors[internalHypervisor.Name] = internalHypervisor
	}
	return hypervisors
}

func RetrieveClusters(manifests string, nodes node.NodeList) cluster.ClusterList {
	clusters := cluster.ClusterList{}
	documents := yamlutils.SplitDocuments(manifests)
	for _, document := range documents {
		scheme := runtime.NewScheme()
		if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
			continue
		}
		clusterObj := clusterv1alpha1.Cluster{}
		gvk := clusterv1alpha1.GroupVersion.WithKind("Cluster")
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
		if _, _, err := serializer.Decode([]byte(document), &gvk, &clusterObj); err != nil {
			continue
		}
		internalCluster, err := cluster.ClusterWithNodesFromv1alpha1(&clusterObj, nodes)
		if err != nil {
			continue
		}
		clusters = append(clusters, internalCluster)
	}
	return clusters
}

func RetrieveNodes(manifests string, hypervisors infra.HypervisorMap) node.NodeList {
	nodes := node.NodeList{}
	documents := yamlutils.SplitDocuments(manifests)
	for _, document := range documents {
		scheme := runtime.NewScheme()
		if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
			continue
		}
		nodeObj := clusterv1alpha1.Node{}
		gvk := clusterv1alpha1.GroupVersion.WithKind("Node")
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
		if _, _, err := serializer.Decode([]byte(document), &gvk, &nodeObj); err != nil {
			continue
		}
		if hypervisor, ok := hypervisors[nodeObj.Spec.Hypervisor]; ok {
			internalNode, err := node.NodeWithHypervisorFromv1alpha1(&nodeObj, hypervisor)
			if err != nil {
				continue
			}
			nodes = append(nodes, internalNode)
		}
	}
	return nodes
}
