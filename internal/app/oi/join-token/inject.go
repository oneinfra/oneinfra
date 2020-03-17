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

package jointoken

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	bootstraptokenutil "k8s.io/cluster-bootstrap/token/util"

	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// Inject injects a join token into the provided cluster spec
func Inject(clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	if len(hypervisors) == 0 {
		return errors.New("empty list of hypervisors")
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	cluster, exists := clusters[clusterName]
	if !exists {
		return errors.Errorf("could not find cluster %q", clusterName)
	}

	bootstrapToken, err := bootstraptokenutil.GenerateBootstrapToken()
	cluster.DesiredJoinTokens = append(cluster.DesiredJoinTokens, bootstrapToken)

	res := ""

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
	fmt.Fprintln(os.Stderr, bootstrapToken)

	return nil
}
