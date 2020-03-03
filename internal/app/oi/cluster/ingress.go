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

	"github.com/oneinfra/oneinfra/m/internal/pkg/cluster/endpoint"
	"github.com/oneinfra/oneinfra/m/internal/pkg/manifests"
)

// IngressComponentName prints the ingress component name for this cluster
func IngressComponentName(clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	cluster, ok := clusters[clusterName]
	if !ok {
		return errors.Errorf("cluster %q not found", clusterName)
	}

	ingressComponent, err := endpoint.IngressComponent(components, cluster)
	if err != nil {
		return err
	}

	fmt.Println(ingressComponent.HypervisorName)

	return nil
}
