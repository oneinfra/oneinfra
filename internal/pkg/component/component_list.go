/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package component

import (
	"errors"
	"fmt"
	"math/rand"
)

// List represents a list of components
type List []*Component

// WithName returns a component matching the provided name
func (list List) WithName(name string) *Component {
	for _, component := range list {
		if component.Name == name {
			return component
		}
	}
	return nil
}

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
func (list List) WithCluster(clusterNamespace, clusterName string) List {
	res := List{}
	for _, component := range list {
		if component.Namespace == clusterNamespace && component.ClusterName == clusterName {
			res = append(res, component)
		}
	}
	return res
}

// Sample returns a random component from the current list
func (list List) Sample() (*Component, error) {
	if len(list) == 0 {
		return nil, errors.New("no components available")
	}
	return list[rand.Intn(len(list))], nil
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

// AllWithHypervisorAssigned returns whether all components in this
// list have an hypervisor assigned
func (list List) AllWithHypervisorAssigned() bool {
	for _, component := range list {
		if component.HypervisorName == "" {
			return false
		}
	}
	return true
}
