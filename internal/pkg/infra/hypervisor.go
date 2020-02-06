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

import (
	"context"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

type Hypervisor struct {
	Name       string
	CRIRuntime criapi.RuntimeServiceClient
	CRIImage   criapi.ImageServiceClient
}

func NewHypervisor(name string, criRuntime criapi.RuntimeServiceClient, criImage criapi.ImageServiceClient) Hypervisor {
	return Hypervisor{
		Name:       name,
		CRIRuntime: criRuntime,
		CRIImage:   criImage,
	}
}

func (hypervisor *Hypervisor) PullImage(image string) error {
	_, err := hypervisor.CRIImage.PullImage(context.Background(), &criapi.PullImageRequest{
		Image: &criapi.ImageSpec{
			Image: image,
		},
	})
	return err
}

func (hypervisor *Hypervisor) RunPod() error {
	podSandboxConfig := criapi.PodSandboxConfig{
		Metadata: &criapi.PodSandboxMetadata{
			Name:      "test",
			Uid:       "test",
			Namespace: "test",
		},
	}
	podSandboxResponse, err := hypervisor.CRIRuntime.RunPodSandbox(
		context.Background(),
		&criapi.RunPodSandboxRequest{
			Config: &podSandboxConfig,
		},
	)
	if err != nil {
		return err
	}
	podSandboxId := podSandboxResponse.PodSandboxId
	containerResponse, err := hypervisor.CRIRuntime.CreateContainer(
		context.Background(),
		&criapi.CreateContainerRequest{
			PodSandboxId: podSandboxId,
			Config: &criapi.ContainerConfig{
				Metadata: &criapi.ContainerMetadata{
					Name: "test",
				},
				Image: &criapi.ImageSpec{
					Image: "k8s.gcr.io/kube-apiserver:v1.17.0",
				},
			},
			SandboxConfig: &podSandboxConfig,
		},
	)
	if err != nil {
		return err
	}
	containerId := containerResponse.ContainerId
	_, err = hypervisor.CRIRuntime.StartContainer(
		context.Background(),
		&criapi.StartContainerRequest{
			ContainerId: containerId,
		},
	)
	return err
}
