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

package component

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// Inject injects a component with name componentName that belongs to cluster
// with name clusterName
func Inject(componentName, clusterName, role string) error {
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

	var injectedComponentRole component.Role
	var hypervisorList infra.HypervisorList
	switch role {
	case "control-plane":
		injectedComponentRole = component.ControlPlaneRole
		hypervisorList = hypervisors.PrivateList()
	case "control-plane-ingress":
		injectedComponentRole = component.ControlPlaneIngressRole
		hypervisorList = hypervisors.PublicList()
	default:
		return errors.Errorf("unknown role %q", role)
	}

	injectedComponent, err := component.NewComponentWithRandomHypervisor(clusterName, componentName, injectedComponentRole, hypervisorList)
	if err != nil {
		return err
	}
	injectedComponentList := component.List{injectedComponent}

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

	if injectedComponentSpecs, err := injectedComponentList.Specs(); err == nil {
		res += injectedComponentSpecs
	}

	fmt.Print(res)

	return nil
}
