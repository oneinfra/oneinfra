/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// KubernetesVersion returns the Kubernetes version for the given
// cluster
func KubernetesVersion(clusterName string) (string, error) {
	var kubernetesVersion string
	err := manifests.WithStdinResourcesSilent(
		func(_ infra.HypervisorMap, clusters cluster.Map, _ component.List) error {
			return manifests.WithNamedCluster(clusterName, clusters, func(cluster *cluster.Cluster) error {
				kubernetesVersion = cluster.KubernetesVersion
				return nil
			})
		},
	)
	return kubernetesVersion, err
}
