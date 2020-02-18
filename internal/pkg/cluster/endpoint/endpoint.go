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

package endpoint

import (
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"

	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

// IngressNode returns the ingress node for the given cluster
func IngressNode(nodes node.List, cluster *cluster.Cluster) (*node.Node, error) {
	for _, nodeObj := range nodes {
		if nodeObj.ClusterName == cluster.Name && nodeObj.Role == node.ControlPlaneIngressRole {
			return nodeObj, nil
		}
	}
	return nil, errors.Errorf("could not find ingress node for cluster %q", cluster.Name)
}

// Endpoint returns the endpoint URI for the given cluster
func Endpoint(nodes node.List, cluster *cluster.Cluster, hypervisors infra.HypervisorMap, endpointHostOverride string) (string, error) {
	ingressNode, err := IngressNode(nodes, cluster)
	if err != nil {
		return "", nil
	}
	apiserverHostPort, ok := ingressNode.AllocatedHostPorts["apiserver"]
	if !ok {
		return "", errors.New("apiserver host port not found")
	}
	var endpoint string
	if len(endpointHostOverride) > 0 {
		endpoint = fmt.Sprintf("https://%s", net.JoinHostPort(endpointHostOverride, strconv.Itoa(apiserverHostPort)))
	} else {
		hypervisor, ok := hypervisors[ingressNode.HypervisorName]
		if !ok {
			return "", errors.Errorf("hypervisor %q not found", ingressNode.HypervisorName)
		}
		endpoint = fmt.Sprintf("https://%s", net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort)))
	}
	return endpoint, nil
}
