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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/manifests"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

// Inject injects a node with name nodeName that belongs to cluster
// with name clusterName
func Inject(nodeName, clusterName, role string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	if len(hypervisors) == 0 {
		return errors.New("empty list of hypervisors")
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	nodes := manifests.RetrieveNodes(string(stdin))

	var injectedNodeRole node.Role
	var hypervisorList infra.HypervisorList
	switch role {
	case "controlplane":
		injectedNodeRole = node.ControlPlaneRole
		hypervisorList = hypervisors.PrivateList()
	case "controlplane-ingress":
		injectedNodeRole = node.ControlPlaneIngressRole
		hypervisorList = hypervisors.PublicList()
	default:
		return errors.Errorf("unknown role %q", role)
	}

	injectedNode, err := node.NewNodeWithRandomHypervisor(clusterName, nodeName, injectedNodeRole, hypervisorList)
	if err != nil {
		return err
	}
	injectedNodeList := node.List{injectedNode}

	res := ""

	if hypervisorsSpecs, err := hypervisors.Specs(); err == nil {
		res += hypervisorsSpecs
	}

	if clustersSpecs, err := clusters.Specs(); err == nil {
		res += clustersSpecs
	}

	if nodesSpecs, err := nodes.Specs(); err == nil {
		res += nodesSpecs
	}

	if injectedNodeSpecs, err := injectedNodeList.Specs(); err == nil {
		res += injectedNodeSpecs
	}

	fmt.Print(res)
	return nil
}
