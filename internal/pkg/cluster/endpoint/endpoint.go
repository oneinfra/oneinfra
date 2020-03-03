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

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
)

// IngressComponent returns the ingress component for the given cluster
func IngressComponent(components component.List, cluster *cluster.Cluster) (*component.Component, error) {
	for _, componentObj := range components {
		if componentObj.ClusterName == cluster.Name && componentObj.Role == component.ControlPlaneIngressRole {
			return componentObj, nil
		}
	}
	return nil, errors.Errorf("could not find ingress component for cluster %q", cluster.Name)
}

// Endpoint returns the endpoint URI for the given cluster
func Endpoint(components component.List, cluster *cluster.Cluster, hypervisors infra.HypervisorMap, endpointHostOverride string) (string, error) {
	ingressComponent, err := IngressComponent(components, cluster)
	if err != nil {
		return "", nil
	}
	apiserverHostPort, ok := ingressComponent.AllocatedHostPorts["apiserver"]
	if !ok {
		return "", errors.New("apiserver host port not found")
	}
	var endpoint string
	if len(endpointHostOverride) > 0 {
		endpoint = fmt.Sprintf("https://%s", net.JoinHostPort(endpointHostOverride, strconv.Itoa(apiserverHostPort)))
	} else {
		hypervisor, ok := hypervisors[ingressComponent.HypervisorName]
		if !ok {
			return "", errors.Errorf("hypervisor %q not found", ingressComponent.HypervisorName)
		}
		endpoint = fmt.Sprintf("https://%s", net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort)))
	}
	return endpoint, nil
}
