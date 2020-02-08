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
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
)

type Hypervisor struct {
	Name               string
	CRIRuntimeEndpoint string
	CRIImageEndpoint   string
	criRuntime         criapi.RuntimeServiceClient
	criImage           criapi.ImageServiceClient
}

func HypervisorFromv1alpha1(hypervisor infrav1alpha1.Hypervisor) (*Hypervisor, error) {
	return &Hypervisor{
		Name:               hypervisor.ObjectMeta.Name,
		CRIRuntimeEndpoint: hypervisor.Spec.CRIRuntimeEndpoint,
		CRIImageEndpoint:   hypervisor.Spec.CRIRuntimeEndpoint,
	}, nil
}

func (hypervisor *Hypervisor) CRIRuntime() (criapi.RuntimeServiceClient, error) {
	if hypervisor.criRuntime != nil {
		return hypervisor.criRuntime, nil
	}
	conn, err := grpc.Dial(hypervisor.CRIRuntimeEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	hypervisor.criRuntime = criapi.NewRuntimeServiceClient(conn)
	return hypervisor.criRuntime, nil
}

func (hypervisor *Hypervisor) CRIImage() (criapi.ImageServiceClient, error) {
	if hypervisor.criImage != nil {
		return hypervisor.criImage, nil
	}
	conn, err := grpc.Dial(hypervisor.CRIImageEndpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	hypervisor.criImage = criapi.NewImageServiceClient(conn)
	return hypervisor.criImage, nil
}

func (hypervisor *Hypervisor) PullImage(image string) error {
	criImage, err := hypervisor.CRIImage()
	if err != nil {
		return err
	}
	_, err = criImage.PullImage(context.Background(), &criapi.PullImageRequest{
		Image: &criapi.ImageSpec{
			Image: image,
		},
	})
	return err
}

func (hypervisor *Hypervisor) PullImages(images ...string) error {
	for _, image := range images {
		if err := hypervisor.PullImage(image); err != nil {
			return err
		}
	}
	return nil
}

func (hypervisor *Hypervisor) RunPod(pod Pod) error {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
	podSandboxConfig := criapi.PodSandboxConfig{
		Metadata: &criapi.PodSandboxMetadata{
			Name:      pod.Name,
			Uid:       uuid.New().String(),
			Namespace: uuid.New().String(),
		},
		LogDirectory: "/var/log/pods/",
	}
	podSandboxResponse, err := criRuntime.RunPodSandbox(
		context.Background(),
		&criapi.RunPodSandboxRequest{
			Config: &podSandboxConfig,
		},
	)
	if err != nil {
		return err
	}
	podSandboxId := podSandboxResponse.PodSandboxId
	containerIds := []string{}
	for _, container := range pod.Containers {
		containerResponse, err := criRuntime.CreateContainer(
			context.Background(),
			&criapi.CreateContainerRequest{
				PodSandboxId: podSandboxId,
				Config: &criapi.ContainerConfig{
					Metadata: &criapi.ContainerMetadata{
						Name: container.Name,
					},
					Image: &criapi.ImageSpec{
						Image: container.Image,
					},
					Args:    container.Command,
					LogPath: fmt.Sprintf("%s-%s-%s.log", pod.Name, podSandboxId, container.Name),
				},
				SandboxConfig: &podSandboxConfig,
			},
		)
		if err != nil {
			return err
		}
		containerIds = append(containerIds, containerResponse.ContainerId)
	}
	for _, containerId := range containerIds {
		_, err = criRuntime.StartContainer(
			context.Background(),
			&criapi.StartContainerRequest{
				ContainerId: containerId,
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}
