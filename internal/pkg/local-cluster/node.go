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

package localcluster

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Node struct {
	Name    string
	Cluster *Cluster
}

func (node *Node) Create() error {
	if err := node.createRuntimeDirectory(); err != nil {
		return err
	}
	cmd := exec.Command(
		"docker", "run", "-d", "--privileged",
		"-v", fmt.Sprintf("%s:/var/run/containerd", node.runtimeDirectory()),
		"--name", fmt.Sprintf("%s-%s", node.Cluster.Name, node.Name),
		"oneinfra/containerd:latest",
	)
	return cmd.Run()
}

func (node *Node) createRuntimeDirectory() error {
	return os.MkdirAll(node.runtimeDirectory(), 0755)
}

func (node *Node) runtimeDirectory() string {
	return filepath.Join(node.Cluster.directory(), node.Name)
}
