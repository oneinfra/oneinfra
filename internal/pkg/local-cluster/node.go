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
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	infraapiv1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
)

type Node struct {
	Name       string
	Cluster    *Cluster
	criRuntime criapi.RuntimeServiceClient
	criImage   criapi.ImageServiceClient
}

func (node *Node) Create() error {
	if err := node.createRuntimeDirectory(); err != nil {
		return err
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	return exec.Command(
		"docker", "run", "-d", "--privileged",
		"--name", fmt.Sprintf("%s-%s", node.Cluster.Name, node.Name),
		"-v", fmt.Sprintf("%s:%s", node.runtimeDirectory(), node.localContainerdSockDirectory()),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_UID=%s", currentUser.Uid),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_GID=%s", currentUser.Gid),
		"-e", fmt.Sprintf("CONTAINER_RUNTIME_ENDPOINT=%s", node.localContainerdSockPath()),
		"-e", fmt.Sprintf("IMAGE_SERVICE_ENDPOINT=%s", node.localContainerdSockPath()),
		"oneinfra/containerd:latest",
	).Run()
}

func (node *Node) Destroy() error {
	exec.Command(
		"docker", "rm", "-f", fmt.Sprintf("%s-%s", node.Cluster.Name, node.Name),
	).Run()
	return os.RemoveAll(node.runtimeDirectory())
}

func (node *Node) localContainerdSockDirectory() string {
	return "/containerd-socket"
}

func (node *Node) localContainerdSockPath() string {
	return fmt.Sprintf("unix://%s/containerd.sock", node.localContainerdSockDirectory())
}

func (node *Node) containerdSockPath() string {
	return fmt.Sprintf("passthrough:///unix://%s", filepath.Join(node.runtimeDirectory(), "containerd.sock"))
}

func (node *Node) CRIRuntime() (criapi.RuntimeServiceClient, error) {
	if node.criRuntime != nil {
		return node.criRuntime, nil
	}
	conn, err := grpc.Dial(node.containerdSockPath(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	node.criRuntime = criapi.NewRuntimeServiceClient(conn)
	return node.criRuntime, nil
}

func (node *Node) CRIImage() (criapi.ImageServiceClient, error) {
	if node.criImage != nil {
		return node.criImage, nil
	}
	conn, err := grpc.Dial(node.containerdSockPath(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	node.criImage = criapi.NewImageServiceClient(conn)
	return node.criImage, nil
}

func (node *Node) Version(ctx context.Context) (*criapi.VersionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	cri, err := node.CRIRuntime()
	if err != nil {
		return nil, err
	}
	return cri.Version(ctx, &criapi.VersionRequest{})
}

func (node *Node) createRuntimeDirectory() error {
	return os.MkdirAll(node.runtimeDirectory(), 0755)
}

func (node *Node) runtimeDirectory() string {
	return filepath.Join(node.Cluster.directory(), node.Name)
}

func (node *Node) Export() infraapiv1alpha1.Hypervisor {
	return infraapiv1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name: node.Name,
		},
		Spec: infraapiv1alpha1.HypervisorSpec{
			CRIRuntimeEndpoint: node.containerdSockPath(),
		},
	}
}
