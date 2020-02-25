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
)

// Inject injects a cluster with name componentName
func Inject(clusterName string, etcdServerExtraSANs, apiServerExtraSANs []string) error {
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

	clusters := manifests.RetrieveClusters(string(stdin))
	if clustersSpecs, err := clusters.Specs(); err == nil {
		res += clustersSpecs
	}

	etcdServerExtraSANsMap := map[string]struct{}{}
	for _, etcdServerExtraSAN := range etcdServerExtraSANs {
		etcdServerExtraSANsMap[etcdServerExtraSAN] = struct{}{}
	}
	for _, ipAddress := range hypervisors.List().IPAddresses() {
		etcdServerExtraSANsMap[ipAddress] = struct{}{}
	}
	finalEtcdServerExtraSANs := []string{}
	for etcdServerExtraSAN := range etcdServerExtraSANsMap {
		finalEtcdServerExtraSANs = append(finalEtcdServerExtraSANs, etcdServerExtraSAN)
	}

	apiServerExtraSANsMap := map[string]struct{}{}
	for _, apiServerExtraSAN := range apiServerExtraSANs {
		apiServerExtraSANsMap[apiServerExtraSAN] = struct{}{}
	}
	for _, publicIPAddress := range hypervisors.PublicList().IPAddresses() {
		apiServerExtraSANsMap[publicIPAddress] = struct{}{}
	}
	finalAPIServerExtraSANs := []string{}
	for apiServerExtraSAN := range apiServerExtraSANsMap {
		finalAPIServerExtraSANs = append(finalAPIServerExtraSANs, apiServerExtraSAN)
	}

	newCluster, err := cluster.NewCluster(clusterName, finalEtcdServerExtraSANs, finalAPIServerExtraSANs)
	if err != nil {
		return err
	}
	injectedCluster := cluster.Map{clusterName: newCluster}
	if injectedClusterSpecs, err := injectedCluster.Specs(); err == nil {
		res += injectedClusterSpecs
	}

	components := manifests.RetrieveComponents(string(stdin))
	if componentsSpecs, err := components.Specs(); err == nil {
		res += componentsSpecs
	}

	fmt.Print(res)
	return nil
}
