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
)

// List represents a list of components
type List []*Component

// WithRole returns a subset of the current list matching the given
// role.
func (list List) WithRole(role Role) List {
	res := List{}
	for _, component := range list {
		if component.Role == role {
			res = append(res, component)
		}
	}
	return res
}

// WithCluster returns a subset of the current list matching the given
// cluster.
func (list List) WithCluster(clusterName string) List {
	res := List{}
	for _, component := range list {
		if component.ClusterName == clusterName {
			res = append(res, component)
		}
	}
	return res
}

// Specs returns the versioned specs of all components in this list.
func (list List) Specs() (string, error) {
	res := ""
	for _, component := range list {
		componentSpec, err := component.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", componentSpec)
	}
	return res, nil
}
