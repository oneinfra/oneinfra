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

package localhypervisorset

import (
	"fmt"

	localhypervisorsetpkg "github.com/oneinfra/oneinfra/internal/pkg/local-hypervisor-set"
)

// Create creates an hypervisor set with name hypervisorSetName, with
// private size privateHypervisorSetSize and public size
// publicHypervisorSetSize
func Create(hypervisorSetName, kubernetesVersion string, privateHypervisorSetSize, publicHypervisorSetSize int, remote bool) error {
	hypervisorSet := localhypervisorsetpkg.NewHypervisorSet(hypervisorSetName, kubernetesVersion, privateHypervisorSetSize, publicHypervisorSetSize, remote)
	if err := hypervisorSet.Create(); err != nil {
		return err
	}
	if err := hypervisorSet.Wait(); err != nil {
		return err
	}
	if remote {
		if err := hypervisorSet.StartRemoteCRIEndpoints(); err != nil {
			return err
		}
	}
	fmt.Print(hypervisorSet.Specs())
	return nil
}
