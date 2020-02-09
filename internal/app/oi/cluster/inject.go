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
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/manifests"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

func Inject(clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	res := ""
	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	if len(hypervisors) == 0 {
		return errors.New("empty list of hypervisors")
	}

	if hypervisorsSpecs, err := hypervisors.Specs(); err == nil {
		res += hypervisorsSpecs
	}

	clusters := manifests.RetrieveClusters(string(stdin), node.NodeList{})
	if clustersSpecs, err := clusters.Specs(); err == nil {
		res += clustersSpecs
	}

	injectedCluster := cluster.ClusterList{
		cluster.NewCluster(clusterName),
	}
	if injectedClusterSpecs, err := injectedCluster.Specs(); err == nil {
		res += injectedClusterSpecs
	}

	nodes := manifests.RetrieveNodes(string(stdin), hypervisors)
	if nodesSpecs, err := nodes.Specs(); err == nil {
		res += nodesSpecs
	}

	fmt.Print(res)
	return nil
}
