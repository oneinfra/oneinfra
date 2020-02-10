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
	"math/rand"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
)

// Hypervisor represents an hypervisor
type Hypervisor struct {
	Name               string
	CRIRuntimeEndpoint string
	CRIImageEndpoint   string
	criRuntime         criapi.RuntimeServiceClient
	criImage           criapi.ImageServiceClient
}

// HypervisorMap represents a map of hypervisors
type HypervisorMap map[string]*Hypervisor

// HypervisorList represents a list of hypervisors
type HypervisorList []*Hypervisor

// NewHypervisorFromv1alpha1 returns an hypervisor based on a versioned hypervisor
func NewHypervisorFromv1alpha1(hypervisor *infrav1alpha1.Hypervisor) (*Hypervisor, error) {
	return &Hypervisor{
		Name:               hypervisor.ObjectMeta.Name,
		CRIRuntimeEndpoint: hypervisor.Spec.CRIRuntimeEndpoint,
		CRIImageEndpoint:   hypervisor.Spec.CRIRuntimeEndpoint,
	}, nil
}

// CRIRuntime returns the runtime service client for the current hypervisor
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

// CRIImage returns the image service client for the current hypervisor
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

// PullImage pulls the requested image on the current hypervisor
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

// PullImages pulls the requested images on the current hypervisor
func (hypervisor *Hypervisor) PullImages(images ...string) error {
	for _, image := range images {
		if err := hypervisor.PullImage(image); err != nil {
			return err
		}
	}
	return nil
}

// RunPod runs a pod on the current hypervisor
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
	podSandboxID := podSandboxResponse.PodSandboxId
	containerIds := []string{}
	for _, container := range pod.Containers {
		containerResponse, err := criRuntime.CreateContainer(
			context.Background(),
			&criapi.CreateContainerRequest{
				PodSandboxId: podSandboxID,
				Config: &criapi.ContainerConfig{
					Metadata: &criapi.ContainerMetadata{
						Name: container.Name,
					},
					Image: &criapi.ImageSpec{
						Image: container.Image,
					},
					Args:    container.Command,
					LogPath: fmt.Sprintf("%s-%s-%s.log", pod.Name, podSandboxID, container.Name),
				},
				SandboxConfig: &podSandboxConfig,
			},
		)
		if err != nil {
			return err
		}
		containerIds = append(containerIds, containerResponse.ContainerId)
	}
	for _, containerID := range containerIds {
		_, err = criRuntime.StartContainer(
			context.Background(),
			&criapi.StartContainerRequest{
				ContainerId: containerID,
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// Export exports the hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	return &infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name: hypervisor.Name,
		},
		Spec: infrav1alpha1.HypervisorSpec{
			CRIRuntimeEndpoint: hypervisor.CRIImageEndpoint,
		},
	}
}

// Specs returns the versioned specs of this hypervisor
func (hypervisor *Hypervisor) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := infrav1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, infrav1alpha1.GroupVersion)
	hypervisorObject := hypervisor.Export()
	if encodedHypervisor, err := runtime.Encode(encoder, hypervisorObject); err == nil {
		return string(encodedHypervisor), nil
	}
	return "", errors.Errorf("could not encode hypervisor %q", hypervisor.Name)
}

// Specs returns the versioned specs of all hypervisors in this map
func (hypervisorMap HypervisorMap) Specs() (string, error) {
	res := ""
	for _, hypervisor := range hypervisorMap {
		hypervisorSpec, err := hypervisor.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", hypervisorSpec)
	}
	return res, nil
}

// List returns a list of hypervisors from this map
func (hypervisorMap HypervisorMap) List() HypervisorList {
	hypervisorList := HypervisorList{}
	for _, hypervisor := range hypervisorMap {
		hypervisorList = append(hypervisorList, hypervisor)
	}
	return hypervisorList
}

// Sample returns a random hypervisor from the curent list
func (hypervisorList HypervisorList) Sample() *Hypervisor {
	return hypervisorList[rand.Intn(len(hypervisorList))]
}
