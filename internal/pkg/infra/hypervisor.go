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
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
)

const (
	toolingImage = "oneinfra/tooling:latest"
)

// Hypervisor represents an hypervisor
type Hypervisor struct {
	Name               string
	CRIRuntimeEndpoint string
	CRIImageEndpoint   string
	criRuntime         criapi.RuntimeServiceClient
	criImage           criapi.ImageServiceClient
	portRangeLow       int
	portRangeHigh      int
	allocatedPorts     HypervisorPortAllocationList
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
		portRangeLow:       hypervisor.Spec.PortRange.Low,
		portRangeHigh:      hypervisor.Spec.PortRange.High,
		allocatedPorts:     NewHypervisorPortAllocationListFromv1alpha1(hypervisor.Status.AllocatedPorts),
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
func (hypervisor *Hypervisor) RunPod(cluster *cluster.Cluster, pod Pod) (string, error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return "", err
	}
	portMappings := []*criapi.PortMapping{}
	for hostPort, podPort := range pod.Ports {
		portMappings = append(portMappings, &criapi.PortMapping{
			HostPort:      int32(hostPort),
			ContainerPort: int32(podPort),
		})
	}
	podSandboxConfig := criapi.PodSandboxConfig{
		Metadata: &criapi.PodSandboxMetadata{
			Name:      pod.Name,
			Uid:       uuid.New().String(),
			Namespace: uuid.New().String(),
		},
		Labels: map[string]string{
			"component": pod.Name,
		},
		PortMappings: portMappings,
		LogDirectory: "/var/log/pods/",
	}
	if cluster != nil {
		podSandboxConfig.Labels["cluster"] = cluster.Name
	}
	podSandboxResponse, err := criRuntime.RunPodSandbox(
		context.Background(),
		&criapi.RunPodSandboxRequest{
			Config: &podSandboxConfig,
		},
	)
	if err != nil {
		return "", err
	}
	podSandboxID := podSandboxResponse.PodSandboxId
	containerIds := []string{}
	for _, container := range pod.Containers {
		containerMounts := []*criapi.Mount{}
		for hostPath, containerPath := range container.Mounts {
			containerMounts = append(containerMounts, &criapi.Mount{
				HostPath:      hostPath,
				ContainerPath: containerPath,
			})
		}
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
					Command: container.Command,
					Args:    container.Args,
					Mounts:  containerMounts,
					LogPath: fmt.Sprintf("%s-%s-%s.log", pod.Name, podSandboxID, container.Name),
				},
				SandboxConfig: &podSandboxConfig,
			},
		)
		if err != nil {
			return "", err
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
			return "", err
		}
	}
	return podSandboxID, nil
}

// WaitForPod waits for all containers in a pod to have exited
func (hypervisor *Hypervisor) WaitForPod(podSandboxID string) error {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
	for {
		containerList, err := criRuntime.ListContainers(
			context.Background(),
			&criapi.ListContainersRequest{
				Filter: &criapi.ContainerFilter{
					PodSandboxId: podSandboxID,
				},
			},
		)
		if err != nil {
			return err
		}
		allContainersExited := true
		for _, container := range containerList.Containers {
			if container.State != criapi.ContainerState_CONTAINER_EXITED {
				allContainersExited = false
				break
			}
		}
		if allContainersExited {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// DeletePod deletes a pod on the current hypervisor
func (hypervisor *Hypervisor) DeletePod(podSandboxID string) error {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
	_, err = criRuntime.StopPodSandbox(
		context.Background(),
		&criapi.StopPodSandboxRequest{
			PodSandboxId: podSandboxID,
		},
	)
	if err != nil {
		return err
	}
	_, err = criRuntime.RemovePodSandbox(
		context.Background(),
		&criapi.RemovePodSandboxRequest{
			PodSandboxId: podSandboxID,
		},
	)
	return err
}

// UploadFiles uploads a map of files, with location as keys, and
// contents as values
func (hypervisor *Hypervisor) UploadFiles(files map[string]string) error {
	for fileLocation, fileContents := range files {
		if err := hypervisor.UploadFile(fileContents, fileLocation); err != nil {
			return err
		}
	}
	return nil
}

// UploadFile uploads a file to the current hypervisor to hostPath
// with given fileContents
func (hypervisor *Hypervisor) UploadFile(fileContents, hostPath string) error {
	if err := hypervisor.PullImage(toolingImage); err != nil {
		return err
	}
	hostPathDir := filepath.Dir(hostPath)
	uploadFilePod := NewSingleContainerPod(
		fmt.Sprintf("upload-file-%x", md5.Sum([]byte(fileContents))),
		toolingImage,
		[]string{"write-base64-file.sh"},
		[]string{
			base64.StdEncoding.EncodeToString([]byte(fileContents)),
			hostPath,
		},
		map[string]string{hostPathDir: hostPathDir},
		map[int]int{},
	)
	podSandboxID, err := hypervisor.RunPod(nil, uploadFilePod)
	if err != nil {
		return err
	}
	if err := hypervisor.WaitForPod(podSandboxID); err != nil {
		return err
	}
	return hypervisor.DeletePod(podSandboxID)
}

// RequestPort requests a port on the current hypervisor
func (hypervisor *Hypervisor) RequestPort(clusterName, nodeName string) (int, error) {
	newPort := hypervisor.portRangeLow + len(hypervisor.allocatedPorts)
	if newPort > hypervisor.portRangeHigh {
		return 0, errors.Errorf("no available ports on hypervisor %q", hypervisor.Name)
	}
	hypervisor.allocatedPorts = append(hypervisor.allocatedPorts, HypervisorPortAllocation{
		Cluster: clusterName,
		Node:    nodeName,
		Port:    newPort,
	})
	return newPort, nil
}

// Export exports the hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	return &infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name: hypervisor.Name,
		},
		Spec: infrav1alpha1.HypervisorSpec{
			CRIRuntimeEndpoint: hypervisor.CRIImageEndpoint,
			PortRange: infrav1alpha1.HypervisorPortRange{
				Low:  hypervisor.portRangeLow,
				High: hypervisor.portRangeHigh,
			},
		},
		Status: infrav1alpha1.HypervisorStatus{
			AllocatedPorts: hypervisor.allocatedPorts.Export(),
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
func (hypervisorList HypervisorList) Sample() (*Hypervisor, error) {
	if len(hypervisorList) == 0 {
		return nil, errors.New("no hypervisors available")
	}
	return hypervisorList[rand.Intn(len(hypervisorList))], nil
}
