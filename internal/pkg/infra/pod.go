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

type Pod struct {
	Name       string
	Containers []Container
}

type Container struct {
	Name    string
	Image   string
	Command []string
}

func NewRegularPod(name string, containers []Container) Pod {
	return Pod{
		Name:       name,
		Containers: containers,
	}
}

func NewSingleContainerPod(name, image string, command []string) Pod {
	return Pod{
		Name: name,
		Containers: []Container{
			{
				Name:    name,
				Image:   image,
				Command: command,
			},
		},
	}
}
