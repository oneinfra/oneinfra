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

package infra

// Pod represents a pod
type Pod struct {
	Name       string
	Containers []Container
	Ports      map[int]int
}

// Container represents a container
type Container struct {
	Name    string
	Image   string
	Command []string
	Args    []string
	Mounts  map[string]string
}

// NewPod returns a pod with name and containers
func NewPod(name string, containers []Container, ports map[int]int) Pod {
	return Pod{
		Name:       name,
		Containers: containers,
		Ports:      ports,
	}
}

// NewSingleContainerPod returns a pod with name, and a single
// container with name, image, command, args, mounts and ports
func NewSingleContainerPod(name, image string, command []string, args []string, mounts map[string]string, ports map[int]int) Pod {
	return Pod{
		Name: name,
		Containers: []Container{
			{
				Name:    name,
				Image:   image,
				Command: command,
				Args:    args,
				Mounts:  mounts,
			},
		},
		Ports: ports,
	}
}
