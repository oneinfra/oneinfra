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

package jointoken

import (
	"github.com/pkg/errors"

	bootstraptokenutil "k8s.io/cluster-bootstrap/token/util"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// Inject injects a join token into the provided cluster spec
func Inject(clusterName string) error {
	return manifests.WithStdinResources(
		func(hypervisors infra.HypervisorMap, clusters cluster.Map, components component.List) (component.List, error) {
			cluster, exists := clusters[clusterName]
			if !exists {
				return component.List{}, errors.Errorf("could not find cluster %q", clusterName)
			}

			bootstrapToken, err := bootstraptokenutil.GenerateBootstrapToken()
			if err != nil {
				return component.List{}, err
			}
			cluster.DesiredJoinTokens = append(cluster.DesiredJoinTokens, bootstrapToken)
			return components, nil
		},
	)
}
