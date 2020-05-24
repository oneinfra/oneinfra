/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package cluster

import (
	"fmt"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
	"k8s.io/klog"
)

// Inject injects a cluster with name componentName
func Inject(clusterName, kubernetesVersion string, controlPlaneReplicas int, vpnEnabled bool, vpnCIDR string, apiServerExtraSANs []string) error {
	return manifests.WithStdinResources(
		func(hypervisors infra.HypervisorMap, clusters cluster.Map, components component.List) (component.List, error) {
			newCluster, err := cluster.NewCluster(clusterName, kubernetesVersion, controlPlaneReplicas, vpnEnabled, vpnCIDR, apiServerExtraSANs)
			if err != nil {
				return component.List{}, err
			}
			clusters[clusterName] = newCluster
			privateHypervisorList := hypervisors.PrivateList()
			for i := 1; i <= newCluster.ControlPlaneReplicas; i++ {
				component, err := component.NewComponentWithRandomHypervisor(
					clusterName,
					fmt.Sprintf("%s-control-plane-%d", clusterName, i),
					component.ControlPlaneRole,
					privateHypervisorList,
				)
				if err != nil {
					klog.Fatalf("could not create new component: %v", err)
				}
				components = append(components, component)
			}
			publicHypervisorList := hypervisors.PublicList()
			component, err := component.NewComponentWithRandomHypervisor(
				clusterName,
				fmt.Sprintf("%s-control-plane-ingress", clusterName),
				component.ControlPlaneIngressRole,
				publicHypervisorList,
			)
			if err != nil {
				klog.Fatalf("could not create new ingress component: %v", err)
			}
			components = append(components, component)
			return components, nil
		},
	)
}
