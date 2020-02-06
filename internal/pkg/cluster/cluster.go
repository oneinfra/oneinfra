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

	"oneinfra.ereslibre.es/m/internal/pkg/infra"
	localcluster "oneinfra.ereslibre.es/m/internal/pkg/local-cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

func Reconcile() error {
	// TODO: all this is to be removed; POC
	cluster, err := localcluster.LoadCluster("test")
	if err != nil {
		return err
	}
	hypervisors := []infra.Hypervisor{}
	for _, node := range cluster.Nodes {
		runtimeCri, err := node.CRIRuntime()
		if err != nil {
			continue
		}
		imageCri, err := node.CRIImage()
		if err != nil {
			continue
		}
		hypervisors = append(hypervisors, infra.NewHypervisor(node.Name, runtimeCri, imageCri))
	}
	if len(hypervisors) == 0 {
		return errors.New("no hypervisors available")
	}
	hypervisor := hypervisors[0]
	newNode := node.NewNode(&hypervisor)
	return newNode.Reconcile()
}
