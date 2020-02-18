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

package node

import (
	"fmt"
)

// List represents a list of nodes
type List []*Node

// WithRole returns a subset of the current list matching the given
// role.
func (list List) WithRole(role Role) List {
	res := List{}
	for _, node := range list {
		if node.Role == role {
			res = append(res, node)
		}
	}
	return res
}

// WithCluster returns a subset of the current list matching the given
// cluster.
func (list List) WithCluster(clusterName string) List {
	res := List{}
	for _, node := range list {
		if node.ClusterName == clusterName {
			res = append(res, node)
		}
	}
	return res
}

// Specs returns the versioned specs of all nodes in this list.
func (list List) Specs() (string, error) {
	res := ""
	for _, node := range list {
		nodeSpec, err := node.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", nodeSpec)
	}
	return res, nil
}
