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
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
	releasecomponents "github.com/oneinfra/oneinfra/internal/pkg/release-components"
)

// KubernetesVersion returns the Kubernetes version for the given
// cluster
func KubernetesVersion(clusterName string) (string, error) {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	clusters := manifests.RetrieveClusters(string(stdin))

	cluster, exists := clusters[clusterName]
	if !exists {
		return "", errors.Errorf("cluster %q not found", clusterName)
	}

	return cluster.KubernetesVersion, nil
}

// ComponentVersion returns the component version for the given
// cluster and component
func ComponentVersion(inputManifests, clusterName string, component releasecomponents.KubernetesComponent) (string, error) {
	clusters := manifests.RetrieveClusters(string(inputManifests))

	cluster, exists := clusters[clusterName]
	if !exists {
		return "", errors.Errorf("cluster %q not found", clusterName)
	}

	componentVersion, err := constants.KubernetesComponentVersion(cluster.KubernetesVersion, component)
	if err != nil {
		return "", err
	}

	return componentVersion, nil
}

// TestComponentVersion returns the test component version for the
// given cluster and component
func TestComponentVersion(inputManifests, clusterName string, testComponent releasecomponents.KubernetesTestComponent) (string, error) {
	clusters := manifests.RetrieveClusters(string(inputManifests))

	cluster, exists := clusters[clusterName]
	if !exists {
		return "", errors.Errorf("cluster %q not found", clusterName)
	}

	testComponentVersion, err := constants.KubernetesTestComponentVersion(cluster.KubernetesVersion, testComponent)
	if err != nil {
		return "", err
	}

	return testComponentVersion, nil
}
