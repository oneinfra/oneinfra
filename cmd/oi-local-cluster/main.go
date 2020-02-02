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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func createClusterDirectory(clusterName string) error {
	return os.MkdirAll(clusterDirectory(clusterName), 0755)
}

func clusterDirectory(clusterName string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("oneinfra-cluster-%s", clusterName))
}

func createContainerdDirectory(name, clusterName string) error {
	return os.MkdirAll(containerdDirectory(name, clusterName), 0755)
}

func containerdDirectory(name, clusterName string) string {
	return filepath.Join(clusterDirectory(clusterName), name)
}

func createNode(name, clusterName string) error {
	if err := createContainerdDirectory(name, clusterName); err != nil {
		return err
	}
	cmd := exec.Command(
		"docker", "run", "-d", "--privileged",
		"-v", fmt.Sprintf("%s:/var/run/containerd", containerdDirectory(name, clusterName)),
		"--name", fmt.Sprintf("%s-%s", clusterName, name),
		"oneinfra/containerd:latest",
	)
	return cmd.Run()
}

func createCluster(clusterName string, clusterSize int) error {
	if err := createClusterDirectory(clusterName); err != nil {
		log.Fatalf("could not create cluster directory")
	}
	for i := 0; i < clusterSize; i++ {
		if err := createNode(fmt.Sprintf("node-%d", i), clusterName); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	clusterName := "test"
	if err := createCluster(clusterName, 3); err != nil {
		log.Fatalf("could not create cluster: %v", err)
	}
	fmt.Printf("Cluster created. Cluster configuration present at %s\n", clusterDirectory(clusterName))
}
