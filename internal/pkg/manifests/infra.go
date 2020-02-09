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
	"sigs.k8s.io/yaml"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
	yamlutils "oneinfra.ereslibre.es/m/internal/pkg/yaml"
)

func RetrieveHypervisors(manifests string) infra.HypervisorMap {
	hypervisors := infra.HypervisorMap{}
	documents := yamlutils.SplitDocuments(manifests)
	for _, document := range documents {
		hypervisor := infrav1alpha1.Hypervisor{}
		if err := yaml.Unmarshal([]byte(document), &hypervisor); err != nil {
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

func RetrieveNodes(manifests string, hypervisors infra.HypervisorMap) node.NodeList {
	nodes := node.NodeList{}
	documents := yamlutils.SplitDocuments(manifests)
	for _, document := range documents {
		nodeObj := clusterv1alpha1.Node{}
		if err := yaml.Unmarshal([]byte(document), &nodeObj); err != nil {
			continue
		}
		if hypervisor, ok := hypervisors[nodeObj.Spec.Hypervisor]; ok {
			internalNode, err := node.NodeFromv1alpha1WithHypervisor(&nodeObj, hypervisor)
			if err != nil {
				continue
			}
			nodes = append(nodes, internalNode)
		}
	}
	return nodes
}
