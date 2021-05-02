/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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

package localhypervisorset

import (
	"os/exec"

	"github.com/pkg/errors"
)

const (
	kindDefaultNetworkName string = "kind"
)

// NetworkName returns a defaulted network, or the provided if exists
func NetworkName(networkName string) (string, error) {
	if networkName == "" {
		if networkExists(kindDefaultNetworkName) {
			return kindDefaultNetworkName, nil
		}
	} else if !networkExists(networkName) {
		return "", errors.Errorf("could not find requested docker network %q, please make sure it exists", networkName)
	}
	return "", nil
}

func networkExists(networkName string) bool {
	return exec.Command("docker", "network", "inspect", networkName).Run() == nil
}
