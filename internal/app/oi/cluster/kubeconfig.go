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
func KubeConfig(clusterName string) error {
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

	// FIXME: simplification for now, use LB
	var firstNode *node.Node
	for _, node := range nodes {
		if node.ClusterName == cluster.Name {
			firstNode = node
			break
		}
	}
	if firstNode == nil {
		return errors.Errorf("could not find any node assigned to cluster %q", cluster.Name)
	}

	kubeConfig, err := cluster.KubeConfig(fmt.Sprintf("https://127.0.0.1:%d", firstNode.HostPort))
	if err != nil {
		return err
	}

	fmt.Print(kubeConfig)

	return nil
}
