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
		"podman", "run", "-d", "--privileged",
		"-v", fmt.Sprintf("%s:/var/run/containerd", containerdDirectory(name, clusterName)),
		"--name", fmt.Sprintf("%s-%s", clusterName, name),
		"docker.io/oneinfra/containerd:latest",
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
