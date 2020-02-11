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
	"os/user"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

// Hypervisor represents a local hypervisor
type Hypervisor struct {
	Name              string
	HypervisorCluster *HypervisorCluster
	CRIRuntime        string
	CRIImage          string
	ExposedPortRange  string
}

// Create creates the local hypervisor
func (hypervisor *Hypervisor) Create() error {
	if err := hypervisor.createRuntimeDirectory(); err != nil {
		return err
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	return exec.Command(
		"docker", "run", "-d", "--privileged",
		"--name", fmt.Sprintf("%s-%s", hypervisor.HypervisorCluster.Name, hypervisor.Name),
		"-v", fmt.Sprintf("%s:%s", hypervisor.runtimeDirectory(), hypervisor.localContainerdSockDirectory()),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_UID=%s", currentUser.Uid),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_GID=%s", currentUser.Gid),
		"-e", fmt.Sprintf("CONTAINER_RUNTIME_ENDPOINT=%s", hypervisor.localContainerdSockPath()),
		"-e", fmt.Sprintf("IMAGE_SERVICE_ENDPOINT=%s", hypervisor.localContainerdSockPath()),
		"-p", fmt.Sprintf("%s:%s", hypervisor.ExposedPortRange, hypervisor.ExposedPortRange),
		"oneinfra/containerd:latest",
	).Run()
}

// Destroy destroys the current hypervisor
func (hypervisor *Hypervisor) Destroy() error {
	exec.Command(
		"docker", "rm", "-f", fmt.Sprintf("%s-%s", hypervisor.HypervisorCluster.Name, hypervisor.Name),
	).Run()
	return os.RemoveAll(hypervisor.runtimeDirectory())
}

func (hypervisor *Hypervisor) localContainerdSockDirectory() string {
	return "/containerd-socket"
}

func (hypervisor *Hypervisor) localContainerdSockPath() string {
	return fmt.Sprintf("unix://%s/containerd.sock", hypervisor.localContainerdSockDirectory())
}

func (hypervisor *Hypervisor) containerdSockPath() string {
	return fmt.Sprintf("passthrough:///unix://%s", filepath.Join(hypervisor.runtimeDirectory(), "containerd.sock"))
}

func (hypervisor *Hypervisor) createRuntimeDirectory() error {
	return os.MkdirAll(hypervisor.runtimeDirectory(), 0755)
}

func (hypervisor *Hypervisor) runtimeDirectory() string {
	return filepath.Join(hypervisor.HypervisorCluster.directory(), hypervisor.Name)
}

// Export exports the local hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	return &infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name: hypervisor.Name,
		},
		Spec: infrav1alpha1.HypervisorSpec{
			CRIRuntimeEndpoint: hypervisor.containerdSockPath(),
		},
	}
}

// Wait waits for the local hypervisor to be created
func (hypervisor *Hypervisor) Wait() error {
	infraHypervisor := infra.Hypervisor{
		Name:               hypervisor.Name,
		CRIRuntimeEndpoint: hypervisor.containerdSockPath(),
		CRIImageEndpoint:   hypervisor.containerdSockPath(),
	}
	for {
		_, runtimeErr := infraHypervisor.CRIRuntime()
		_, imageErr := infraHypervisor.CRIImage()
		if runtimeErr == nil && imageErr == nil {
			break
		}
	}
	return nil
}
