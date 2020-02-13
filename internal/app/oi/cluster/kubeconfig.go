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

package cluster

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"oneinfra.ereslibre.es/m/internal/pkg/manifests"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

// KubeConfig generates a kubeconfig for cluster clusterName
func KubeConfig(clusterName, endpointHostOverride string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	clusters := manifests.RetrieveClusters(string(stdin))
	nodes := manifests.RetrieveNodes(string(stdin))

	cluster, ok := clusters[clusterName]
	if !ok {
		return errors.Errorf("cluster %q not found", clusterName)
	}

	var firstNode *node.Node
	for _, nodeObj := range nodes {
		if nodeObj.ClusterName == cluster.Name && nodeObj.Role == node.ControlPlaneIngressRole {
			firstNode = nodeObj
			break
		}
	}
	if firstNode == nil {
		return errors.Errorf("could not find any control plane ingress role node assigned to cluster %q", cluster.Name)
	}

	var endpoint string
	if len(endpointHostOverride) > 0 {
		endpoint = fmt.Sprintf("https://%s:%d", endpointHostOverride, firstNode.HostPort)
	} else {
		hypervisors := manifests.RetrieveHypervisors(string(stdin))
		hypervisor, ok := hypervisors[firstNode.HypervisorName]
		if !ok {
			return errors.Errorf("hypervisor %q not found", firstNode.HypervisorName)
		}
		endpoint = fmt.Sprintf("https://%s:%d", hypervisor.IPAddress, firstNode.HostPort)
	}

	kubeConfig, err := cluster.KubeConfig(endpoint)
	if err != nil {
		return err
	}

	fmt.Print(kubeConfig)

	return nil
}
