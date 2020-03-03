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

package vpnpeer

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/oneinfra/oneinfra/m/internal/pkg/manifests"
)

// Inject injects a VPN peer with name peerName
func Inject(peerName, clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	res := ""

	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	if len(hypervisors) == 0 {
		return errors.New("empty list of hypervisors")
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	if cluster, ok := clusters[clusterName]; ok {
		if err := cluster.GenerateVPNPeer(peerName); err != nil {
			return errors.Wrapf(err, "could not inject VPN peer %q", peerName)
		}
	} else {
		return errors.Errorf("could not find cluster %q", clusterName)
	}

	if hypervisorsSpecs, err := hypervisors.Specs(); err == nil {
		res += hypervisorsSpecs
	}
	if clustersSpecs, err := clusters.Specs(); err == nil {
		res += clustersSpecs
	}
	if componentsSpecs, err := components.Specs(); err == nil {
		res += componentsSpecs
	}

	fmt.Print(res)
	return nil
}
