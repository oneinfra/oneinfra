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

package pod

import (
	"crypto/sha1"
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Privileges represents the privileges that a pod sandbox and a
// container has
type Privileges int

const (
	// PrivilegesUnprivileged represents a pod sandbox or a container
	// with no special privileges
	PrivilegesUnprivileged Privileges = 0
	// PrivilegesPrivileged represents a pod sandbox or a container with
	// privileged access
	PrivilegesPrivileged Privileges = 1
	// PrivilegesNetworkPrivileged represents a pod sandbox or a
	// privileged container with node network. Implies
	// PrivilegesPrivileged.
	PrivilegesNetworkPrivileged Privileges = 3
)

// Pod represents a pod
type Pod struct {
	Name       string
	Containers []Container
	Ports      map[int]int
	Privileges Privileges
}

// Container represents a container
type Container struct {
	Name        string
	Image       string
	Command     []string
	Args        []string
	Env         map[string]string
	Mounts      map[string]string
	Privileges  Privileges
	Annotations map[string]string
}

// NewPod returns a pod with name, containers, ports and privileges
func NewPod(name string, containers []Container, ports map[int]int, privileges Privileges) Pod {
	return Pod{
		Name:       name,
		Containers: containers,
		Ports:      ports,
		Privileges: privileges,
	}
}

// SHA1Sum returns the SHA-1 of the textual YAML representation of
// this pod
func (pod *Pod) SHA1Sum() (string, error) {
	podManifest, err := yaml.Marshal(pod)
	if err != nil {
		return "", errors.Errorf("cannot marshal pod %q: %v", pod.Name, err)
	}
	return fmt.Sprintf("%x", sha1.Sum(podManifest)), nil
}
