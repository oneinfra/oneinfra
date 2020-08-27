/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package infra

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"

	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
)

const (
	// ToolingImage is the tooling image repository and tag
	ToolingImage = "oneinfra/tooling:latest"
)

const (
	podSandboxSHA1SumLabel = "oneinfra/pod-sha1sum"
	clusterNamespaceLabel  = "oneinfra/cluster-namespace"
	clusterNameLabel       = "oneinfra/cluster-name"
	componentNameLabel     = "oneinfra/component-name"
	podNameLabel           = "oneinfra/pod-name"
)

// Hypervisor represents an hypervisor
type Hypervisor struct {
	Name               string
	ResourceVersion    string
	Labels             map[string]string
	Annotations        map[string]string
	Public             bool
	IPAddress          string
	Files              NamespacedClusterFileMap
	Endpoint           hypervisorEndpoint
	criRuntime         criapi.RuntimeServiceClient
	criImage           criapi.ImageServiceClient
	portRangeLow       int
	portRangeHigh      int
	freedPorts         []int
	allocatedPorts     HypervisorPortAllocationList
	loadedContentsHash string
	connectionPool     *HypervisorConnectionPool
}

// HypervisorMap represents a map of hypervisors
type HypervisorMap map[string]*Hypervisor

// HypervisorList represents a list of hypervisors
type HypervisorList []*Hypervisor

// NewHypervisorFromv1alpha1 returns an hypervisor based on a versioned hypervisor
func NewHypervisorFromv1alpha1(hypervisor *infrav1alpha1.Hypervisor, connectionPool *HypervisorConnectionPool) (*Hypervisor, error) {
	hypervisorFiles := hypervisor.Status.Files
	if hypervisorFiles == nil {
		hypervisorFiles = infrav1alpha1.NamespacedClusterFileMap{}
	}
	if connectionPool == nil {
		connectionPool = &HypervisorConnectionPool{}
	}
	res := Hypervisor{
		Name:            hypervisor.Name,
		ResourceVersion: hypervisor.ResourceVersion,
		Labels:          hypervisor.Labels,
		Annotations:     hypervisor.Annotations,
		Public:          hypervisor.Spec.Public,
		IPAddress:       hypervisor.Spec.IPAddress,
		Files:           NewNamespacedClusterFileMapFromv1alpha1(hypervisorFiles),
		portRangeLow:    hypervisor.Spec.PortRange.Low,
		portRangeHigh:   hypervisor.Spec.PortRange.High,
		freedPorts:      hypervisor.Status.FreedPorts,
		allocatedPorts:  NewHypervisorPortAllocationListFromv1alpha1(hypervisor.Status.AllocatedPorts),
		connectionPool:  connectionPool,
	}
	if err := setHypervisorEndpointFromv1alpha1(hypervisor, connectionPool, &res); err != nil {
		return nil, err
	}
	if err := res.RefreshCachedSpecs(); err != nil {
		return nil, err
	}
	return &res, nil
}

// NewLocalHypervisor creates a local hypervisor
func NewLocalHypervisor(name, criEndpoint string) *Hypervisor {
	return &Hypervisor{
		Name: name,
		Endpoint: &localHypervisorEndpoint{
			CRIEndpoint: criEndpoint,
		},
	}
}

// CRIRuntime returns the runtime service client for the current hypervisor
func (hypervisor *Hypervisor) CRIRuntime() (criapi.RuntimeServiceClient, error) {
	if hypervisor.criRuntime != nil {
		return hypervisor.criRuntime, nil
	}
	conn, err := hypervisor.Endpoint.Connection()
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
	conn, err := hypervisor.Endpoint.Connection()
	if err != nil {
		return nil, err
	}
	hypervisor.criImage = criapi.NewImageServiceClient(conn)
	return hypervisor.criImage, nil
}

// EnsureImage ensures that the requested image is present on the current hypervisor
func (hypervisor *Hypervisor) EnsureImage(image string) error {
	klog.V(2).Infof("ensuring that image %q exists in hypervisor %q", image, hypervisor.Name)
	criImage, err := hypervisor.CRIImage()
	if err != nil {
		return err
	}
	imageStatus, err := criImage.ImageStatus(context.TODO(), &criapi.ImageStatusRequest{
		Image: &criapi.ImageSpec{
			Image: image,
		},
	})
	if err != nil {
		return err
	}
	if imageStatus.Image == nil {
		_, err = criImage.PullImage(context.TODO(), &criapi.PullImageRequest{
			Image: &criapi.ImageSpec{
				Image: image,
			},
		})
		return err
	}
	return nil
}

// EnsureImages ensures that the requested images are present on the current hypervisor
func (hypervisor *Hypervisor) EnsureImages(images ...string) error {
	for _, image := range images {
		if err := hypervisor.EnsureImage(image); err != nil {
			return err
		}
	}
	return nil
}

// PodSandboxConfig returns a pod sandbox config for the given pod and cluster
func (hypervisor *Hypervisor) PodSandboxConfig(clusterNamespace, clusterName, componentName string, pod podapi.Pod) (criapi.PodSandboxConfig, error) {
	portMappings := []*criapi.PortMapping{}
	for hostPort, podPort := range pod.Ports {
		portMappings = append(portMappings, &criapi.PortMapping{
			HostPort:      int32(hostPort),
			ContainerPort: int32(podPort),
		})
	}
	podSum, err := pod.SHA1Sum()
	if err != nil {
		return criapi.PodSandboxConfig{}, err
	}
	clusterAndComponentName := ""
	if len(clusterNamespace) > 0 {
		clusterAndComponentName += fmt.Sprintf("%s-", clusterNamespace)
	}
	if len(clusterName) > 0 {
		clusterAndComponentName += fmt.Sprintf("%s-", clusterName)
	}
	if len(componentName) > 0 {
		clusterAndComponentName += fmt.Sprintf("%s-", componentName)
	}
	podSandboxConfig := criapi.PodSandboxConfig{
		Metadata: &criapi.PodSandboxMetadata{
			Name:      pod.Name,
			Namespace: fmt.Sprintf("%s%s-%s", clusterAndComponentName, pod.Name, podSum),
			Uid:       podSum,
		},
		Labels: map[string]string{
			clusterNamespaceLabel:  clusterNamespace,
			clusterNameLabel:       clusterName,
			componentNameLabel:     componentName,
			podNameLabel:           pod.Name,
			podSandboxSHA1SumLabel: podSum,
		},
		PortMappings: portMappings,
		LogDirectory: "/var/log/pods/",
	}
	if pod.Privileges&podapi.PrivilegesPrivileged != 0 {
		podSandboxConfig.Linux = &criapi.LinuxPodSandboxConfig{
			SecurityContext: &criapi.LinuxSandboxSecurityContext{
				Privileged: true,
			},
		}
	}
	if pod.Privileges == podapi.PrivilegesNetworkPrivileged {
		podSandboxConfig.Linux.SecurityContext.NamespaceOptions = &criapi.NamespaceOption{
			Network: criapi.NamespaceMode_NODE,
		}
	}
	return podSandboxConfig, nil
}

// IsPodRunning returns whether a pod is running on the current
// hypervisor, along with extra information about the pod sandbox ID,
// what containers are currently running, and which are not
func (hypervisor *Hypervisor) IsPodRunning(clusterNamespace, clusterName, componentName string, pod podapi.Pod) (isPodRunning bool, podSandboxID string, allContainersRunning bool, runningContainers, notRunningContainers map[string]*criapi.Container, err error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return false, "", false, nil, nil, err
	}
	podSum, err := pod.SHA1Sum()
	if err != nil {
		return false, "", false, nil, nil, err
	}
	klog.V(2).Infof("checking if a pod %q in hypervisor %q is running", pod.Name, hypervisor.Name)
	podSandboxList, err := criRuntime.ListPodSandbox(
		context.TODO(),
		&criapi.ListPodSandboxRequest{
			Filter: &criapi.PodSandboxFilter{
				LabelSelector: map[string]string{
					clusterNamespaceLabel:  clusterNamespace,
					clusterNameLabel:       clusterName,
					componentNameLabel:     componentName,
					podNameLabel:           pod.Name,
					podSandboxSHA1SumLabel: podSum,
				},
			},
		},
	)
	if err != nil {
		return false, "", false, nil, nil, err
	}
	podSandboxID = ""
	for _, podSandbox := range podSandboxList.Items {
		if podSandbox.State == criapi.PodSandboxState_SANDBOX_READY {
			podSandboxID = podSandbox.Id
			break
		}
	}
	if podSandboxID == "" {
		return false, "", false, nil, nil, nil
	}
	klog.V(2).Infof("checking if all containers within pod %q in hypervisor %q are running", pod.Name, hypervisor.Name)
	containerList, err := criRuntime.ListContainers(
		context.TODO(),
		&criapi.ListContainersRequest{
			Filter: &criapi.ContainerFilter{
				PodSandboxId: podSandboxID,
			},
		},
	)
	if err != nil {
		return false, podSandboxID, false, nil, nil, err
	}
	podRunningContainers := map[string]*criapi.Container{}
	podNotRunningContainers := map[string]*criapi.Container{}
	for _, container := range containerList.Containers {
		containerStatus, err := criRuntime.ContainerStatus(
			context.TODO(),
			&criapi.ContainerStatusRequest{
				ContainerId: container.Id,
			},
		)
		if err != nil {
			continue
		}
		if containerStatus.Status.State == criapi.ContainerState_CONTAINER_RUNNING {
			podRunningContainers[container.Metadata.Name] = container
		} else {
			podNotRunningContainers[container.Metadata.Name] = container
		}
	}
	return true,
		podSandboxID,
		len(podRunningContainers) == len(pod.Containers),
		podRunningContainers,
		podNotRunningContainers,
		nil
}

// EnsurePod runs a pod on the current hypervisor
func (hypervisor *Hypervisor) EnsurePod(clusterNamespace, clusterName, componentName string, pod podapi.Pod) (string, error) {
	isPodRunning, podSandboxID, allContainersRunning, podRunningContainers, podNotRunningContainers, err := hypervisor.IsPodRunning(clusterNamespace, clusterName, componentName, pod)
	if err != nil {
		return "", err
	}
	if isPodRunning && allContainersRunning {
		klog.V(2).Infof("pod %q and all its containers in hypervisor %q are running", pod.Name, hypervisor.Name)
		return podSandboxID, nil
	}
	if err := hypervisor.DeletePod(clusterNamespace, clusterName, componentName, pod.Name); err != nil {
		klog.V(2).Infof("could not delete pods named %q: %v", pod.Name, err)
	}
	return hypervisor.ensurePod(
		clusterNamespace,
		clusterName,
		componentName,
		podSandboxID,
		podRunningContainers,
		podNotRunningContainers,
		pod,
	)
}

func (hypervisor *Hypervisor) ensurePod(clusterNamespace, clusterName, componentName string, podSandboxID string, podRunningContainers, podNotRunningContainers map[string]*criapi.Container, pod podapi.Pod) (string, error) {
	klog.V(2).Infof("running pod %q in hypervisor %q", pod.Name, hypervisor.Name)
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return "", err
	}
	podSandboxConfig, err := hypervisor.PodSandboxConfig(clusterNamespace, clusterName, componentName, pod)
	if err != nil {
		return "", err
	}
	if podSandboxID == "" {
		podSandboxResponse, err := criRuntime.RunPodSandbox(
			context.TODO(),
			&criapi.RunPodSandboxRequest{
				Config: &podSandboxConfig,
			},
		)
		if err != nil {
			return "", err
		}
		podSandboxID = podSandboxResponse.PodSandboxId
	}
	err = hypervisor.ensureContainers(
		criRuntime,
		podSandboxID,
		podSandboxConfig,
		podRunningContainers,
		podNotRunningContainers,
		pod,
	)
	return podSandboxID, err
}

func (hypervisor *Hypervisor) ensureContainers(criRuntime criapi.RuntimeServiceClient, podSandboxID string, podSandboxConfig criapi.PodSandboxConfig, podRunningContainers, podNotRunningContainers map[string]*criapi.Container, pod podapi.Pod) error {
	containerIds := []string{}
	for _, container := range pod.Containers {
		if _, exists := podRunningContainers[container.Name]; exists {
			continue
		}
		if notRunningContainer, exists := podNotRunningContainers[container.Name]; exists {
			_, err := criRuntime.RemoveContainer(
				context.TODO(),
				&criapi.RemoveContainerRequest{
					ContainerId: notRunningContainer.Id,
				})
			if err != nil {
				klog.Warningf("failed to remove container %q: %v", notRunningContainer.Id, err)
			}
		}
		containerMounts := []*criapi.Mount{}
		for hostPath, containerPath := range container.Mounts {
			containerMounts = append(containerMounts, &criapi.Mount{
				HostPath:      hostPath,
				ContainerPath: containerPath,
			})
		}
		createContainerRequest := criapi.CreateContainerRequest{
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
		}
		if len(container.Env) > 0 {
			createContainerRequest.Config.Envs = []*criapi.KeyValue{}
			for envVar, envValue := range container.Env {
				createContainerRequest.Config.Envs = append(
					createContainerRequest.Config.Envs,
					&criapi.KeyValue{
						Key:   envVar,
						Value: envValue,
					},
				)
			}
		}
		if container.Privileges&podapi.PrivilegesPrivileged != 0 {
			createContainerRequest.Config.Linux = &criapi.LinuxContainerConfig{
				SecurityContext: &criapi.LinuxContainerSecurityContext{
					Privileged: true,
				},
			}
		}
		if container.Privileges == podapi.PrivilegesNetworkPrivileged {
			createContainerRequest.Config.Linux.SecurityContext.NamespaceOptions = &criapi.NamespaceOption{
				Network: criapi.NamespaceMode_NODE,
			}
		}
		containerResponse, err := criRuntime.CreateContainer(
			context.TODO(),
			&createContainerRequest,
		)
		if err != nil {
			return err
		}
		containerIds = append(containerIds, containerResponse.ContainerId)
	}
	for _, containerID := range containerIds {
		_, err := criRuntime.StartContainer(
			context.TODO(),
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

// WaitForPod waits for all containers in a pod to have exited
func (hypervisor *Hypervisor) WaitForPod(podSandboxID string) error {
	klog.V(2).Infof("waiting for pod %q to have completed on hypervisor %q", podSandboxID, hypervisor.Name)
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
	for {
		containerList, err := criRuntime.ListContainers(
			context.TODO(),
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

// ListPods returns a list of pod sandbox ID's that belong to the
// provided cluster, component and pod name
func (hypervisor *Hypervisor) ListPods(clusterNamespace, clusterName, componentName, podName string) ([]string, error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return []string{}, err
	}
	podSandboxList, err := criRuntime.ListPodSandbox(
		context.TODO(),
		&criapi.ListPodSandboxRequest{
			Filter: &criapi.PodSandboxFilter{
				LabelSelector: map[string]string{
					clusterNamespaceLabel: clusterNamespace,
					clusterNameLabel:      clusterName,
					componentNameLabel:    componentName,
					podNameLabel:          podName,
				},
			},
		},
	)
	if err != nil {
		return []string{}, errors.Errorf("could not list pods for cluster %q/%q and component %q", clusterNamespace, clusterName, componentName)
	}
	res := []string{}
	for _, podSandbox := range podSandboxList.Items {
		res = append(res, podSandbox.Id)
	}
	return res, nil
}

// ListAllPods returns a list of pod sandbox ID's that belong to the
// provided cluster and component
func (hypervisor *Hypervisor) ListAllPods(clusterNamespace, clusterName, componentName string) ([]string, error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return []string{}, err
	}
	podSandboxList, err := criRuntime.ListPodSandbox(
		context.TODO(),
		&criapi.ListPodSandboxRequest{
			Filter: &criapi.PodSandboxFilter{
				LabelSelector: map[string]string{
					clusterNamespaceLabel: clusterNamespace,
					clusterNameLabel:      clusterName,
					componentNameLabel:    componentName,
				},
			},
		},
	)
	if err != nil {
		return []string{}, errors.Errorf("could not list pods for cluster %q/%q and component %q", clusterNamespace, clusterName, componentName)
	}
	res := []string{}
	for _, podSandbox := range podSandboxList.Items {
		res = append(res, podSandbox.Id)
	}
	return res, nil
}

// DeletePods deletes all pods matching the given cluster and component
func (hypervisor *Hypervisor) DeletePods(clusterNamespace, clusterName, componentName string) error {
	klog.V(2).Infof("deleting pods for cluster %q and component %q from hypervisor %q", clusterName, componentName, hypervisor.Name)
	podList, err := hypervisor.ListAllPods(clusterNamespace, clusterName, componentName)
	if err != nil {
		return err
	}
	for _, pod := range podList {
		if err := hypervisor.DeletePodWithID(pod); err != nil {
			return err
		}
	}
	return nil
}

// DeletePod deletes all pods matching the given cluster, component and pod name
func (hypervisor *Hypervisor) DeletePod(clusterNamespace, clusterName, componentName, podName string) error {
	podList, err := hypervisor.ListPods(clusterNamespace, clusterName, componentName, podName)
	if err != nil {
		return err
	}
	for _, pod := range podList {
		if err := hypervisor.DeletePodWithID(pod); err != nil {
			return err
		}
	}
	return nil
}

// DeletePodWithID deletes a pod on the current hypervisor
func (hypervisor *Hypervisor) DeletePodWithID(podSandboxID string) error {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
	klog.V(2).Infof("deleting pod %q from hypervisor %q", podSandboxID, hypervisor.Name)
	_, err = criRuntime.StopPodSandbox(
		context.TODO(),
		&criapi.StopPodSandboxRequest{
			PodSandboxId: podSandboxID,
		},
	)
	if err != nil {
		return err
	}
	_, err = criRuntime.RemovePodSandbox(
		context.TODO(),
		&criapi.RemovePodSandboxRequest{
			PodSandboxId: podSandboxID,
		},
	)
	return err
}

// RunAndWaitForPod runs and waits for all containers within a pod to be finished
func (hypervisor *Hypervisor) RunAndWaitForPod(clusterNamespace, clusterName, componentName string, pod podapi.Pod) error {
	podSandboxID, err := hypervisor.EnsurePod(clusterNamespace, clusterName, componentName, pod)
	if err != nil {
		return err
	}
	if err := hypervisor.WaitForPod(podSandboxID); err != nil {
		return err
	}
	return hypervisor.DeletePodWithID(podSandboxID)
}

// UploadFiles uploads a map of files, with location as keys, and
// contents as values
func (hypervisor *Hypervisor) UploadFiles(clusterNamespace, clusterName, componentName string, files map[string]string) error {
	filesToUpload := []podapi.Container{}
	for fileLocation, fileContents := range files {
		if hypervisor.FileUpToDate(clusterNamespace, clusterName, componentName, fileLocation, fileContents) {
			klog.V(2).Infof("skipping file upload to hypervisor %q at location %q, hash matches", hypervisor.Name, fileLocation)
			continue
		}
		klog.V(2).Infof("preparing file upload to hypervisor %q at location %q", hypervisor.Name, fileLocation)
		fileLocationDir := filepath.Dir(fileLocation)
		filesToUpload = append(
			filesToUpload,
			podapi.Container{
				Name:    fmt.Sprintf("upload-file-%x", md5.Sum([]byte(fileContents))),
				Image:   ToolingImage,
				Command: []string{"write-base64-file.sh"},
				Args: []string{
					base64.StdEncoding.EncodeToString([]byte(fileContents)),
					fileLocation,
				},
				Mounts:     map[string]string{fileLocationDir: fileLocationDir},
				Privileges: podapi.PrivilegesUnprivileged,
			},
		)
	}
	if len(filesToUpload) == 0 {
		return nil
	}
	if err := hypervisor.EnsureImage(ToolingImage); err != nil {
		return err
	}
	err := hypervisor.RunAndWaitForPod(
		clusterNamespace,
		clusterName,
		componentName,
		podapi.Pod{
			Name:       "upload-files",
			Containers: filesToUpload,
			Ports:      map[int]int{},
			Privileges: podapi.PrivilegesUnprivileged,
		},
	)
	if err == nil && hypervisor.Files != nil {
		if hypervisor.Files[clusterNamespace] == nil {
			hypervisor.Files[clusterNamespace] = ClusterFileMap{}
		}
		if hypervisor.Files[clusterNamespace][clusterName] == nil {
			hypervisor.Files[clusterNamespace][clusterName] = ComponentFileMap{}
		}
		if hypervisor.Files[clusterNamespace][clusterName][componentName] == nil {
			hypervisor.Files[clusterNamespace][clusterName][componentName] = FileMap{}
		}
		for fileLocation, fileContents := range files {
			hypervisor.Files[clusterNamespace][clusterName][componentName][fileLocation] = fmt.Sprintf("%x", sha1.Sum([]byte(fileContents)))
		}
	}
	return err
}

// UploadFile uploads a file to the current hypervisor to hostPath
// with given fileContents
func (hypervisor *Hypervisor) UploadFile(clusterNamespace, clusterName, componentName, hostPath, fileContents string) error {
	return hypervisor.UploadFiles(
		clusterNamespace,
		clusterName,
		componentName,
		map[string]string{
			hostPath: fileContents,
		},
	)
}

// FileUpToDate returns whether the given file contents match on the
// host
func (hypervisor *Hypervisor) FileUpToDate(clusterNamespace, clusterName, componentName, hostPath, fileContents string) bool {
	fileContentsSHA1 := fmt.Sprintf("%x", sha1.Sum([]byte(fileContents)))
	if currentFileContentsSHA1, exists := hypervisor.Files[clusterNamespace][clusterName][componentName][hostPath]; exists {
		return currentFileContentsSHA1 == fileContentsSHA1
	}
	return false
}

// HasPort returns whether a port exists for the given clusterName and componentName
func (hypervisor *Hypervisor) HasPort(clusterNamespace, clusterName, componentName string) (bool, int) {
	for _, allocatedPort := range hypervisor.allocatedPorts {
		if allocatedPort.ClusterNamespace == clusterNamespace && allocatedPort.Cluster == clusterName && allocatedPort.Component == componentName {
			return true, allocatedPort.Port
		}
	}
	return false, 0
}

// RequestPort requests a port on the current hypervisor
func (hypervisor *Hypervisor) RequestPort(clusterNamespace, clusterName, componentName string) (int, error) {
	if hasPort, existingPort := hypervisor.HasPort(clusterNamespace, clusterName, componentName); hasPort {
		return existingPort, nil
	}
	var newPort int
	if len(hypervisor.freedPorts) > 0 {
		newPort, hypervisor.freedPorts = hypervisor.freedPorts[0], hypervisor.freedPorts[1:]
	} else {
		newPort = hypervisor.portRangeLow + len(hypervisor.allocatedPorts)
		if newPort > hypervisor.portRangeHigh {
			return 0, errors.Errorf("no available ports on hypervisor %q", hypervisor.Name)
		}
	}
	hypervisor.allocatedPorts = append(hypervisor.allocatedPorts, HypervisorPortAllocation{
		ClusterNamespace: clusterNamespace,
		Cluster:          clusterName,
		Component:        componentName,
		Port:             newPort,
	})
	klog.V(2).Infof("port requested for hypervisor %q; assigned: %d", hypervisor.Name, newPort)
	return newPort, nil
}

// FreePort frees a port on the given hypervisor
func (hypervisor *Hypervisor) FreePort(clusterNamespace, clusterName, componentName string) error {
	newAllocatedPorts := HypervisorPortAllocationList{}
	var hypervisorPortAllocation *HypervisorPortAllocation
	for _, portAllocation := range hypervisor.allocatedPorts {
		if portAllocation.ClusterNamespace == clusterNamespace && portAllocation.Cluster == clusterName && portAllocation.Component == componentName {
			portAllocation := portAllocation
			hypervisorPortAllocation = &portAllocation
		} else {
			newAllocatedPorts = append(newAllocatedPorts, portAllocation)
		}
	}
	if hypervisorPortAllocation == nil {
		return nil
	}
	hypervisor.allocatedPorts = newAllocatedPorts
	if hypervisor.freedPorts == nil {
		hypervisor.freedPorts = []int{}
	}
	hypervisor.freedPorts = append(
		hypervisor.freedPorts,
		hypervisorPortAllocation.Port,
	)
	return nil
}

// Export exports the hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	resHypervisor := infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            hypervisor.Name,
			ResourceVersion: hypervisor.ResourceVersion,
			Labels:          hypervisor.Labels,
			Annotations:     hypervisor.Annotations,
		},
		Spec: infrav1alpha1.HypervisorSpec{
			Public:    hypervisor.Public,
			IPAddress: hypervisor.IPAddress,
			PortRange: infrav1alpha1.HypervisorPortRange{
				Low:  hypervisor.portRangeLow,
				High: hypervisor.portRangeHigh,
			},
		},
		Status: infrav1alpha1.HypervisorStatus{
			AllocatedPorts: hypervisor.allocatedPorts.Export(),
			FreedPorts:     hypervisor.freedPorts,
			Files:          hypervisor.Files.Export(),
		},
	}
	resHypervisor.Spec.LocalCRIEndpoint, resHypervisor.Spec.RemoteCRIEndpoint = hypervisor.Endpoint.Export()
	return &resHypervisor
}

// RefreshCachedSpecs refreshes the cached spec
func (hypervisor *Hypervisor) RefreshCachedSpecs() error {
	specs, err := hypervisor.Specs()
	if err != nil {
		return err
	}
	hypervisor.loadedContentsHash = fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return nil
}

// IsDirty returns whether this cluster is dirty compared to when it
// was loaded
func (hypervisor *Hypervisor) IsDirty() (bool, error) {
	specs, err := hypervisor.Specs()
	if err != nil {
		return false, err
	}
	currentContentsHash := fmt.Sprintf("%x", sha1.Sum([]byte(specs)))
	return hypervisor.loadedContentsHash != currentContentsHash, nil
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

// List converts the hypervisor map to an hypervisor list
func (hypervisorMap HypervisorMap) List() HypervisorList {
	hypervisorList := HypervisorList{}
	for _, hypervisor := range hypervisorMap {
		hypervisorList = append(hypervisorList, hypervisor)
	}
	return hypervisorList
}

// PublicList returns a list of public hypervisors from this map
func (hypervisorMap HypervisorMap) PublicList() HypervisorList {
	hypervisorList := HypervisorList{}
	for _, hypervisor := range hypervisorMap {
		if hypervisor.Public {
			hypervisorList = append(hypervisorList, hypervisor)
		}
	}
	return hypervisorList
}

// PrivateList returns a list of private hypervisors from this map
func (hypervisorMap HypervisorMap) PrivateList() HypervisorList {
	hypervisorList := HypervisorList{}
	for _, hypervisor := range hypervisorMap {
		if hypervisor.Public {
			continue
		}
		hypervisorList = append(hypervisorList, hypervisor)
	}
	return hypervisorList
}

// IPAddresses returns the list of IP addresses
func (hypervisorList HypervisorList) IPAddresses() []string {
	ipAddresses := []string{}
	for _, hypervisor := range hypervisorList {
		ipAddresses = append(ipAddresses, hypervisor.IPAddress)
	}
	return ipAddresses
}

// Sample returns a random hypervisor from the current list
func (hypervisorList HypervisorList) Sample() (*Hypervisor, error) {
	if len(hypervisorList) == 0 {
		return nil, errors.New("no hypervisors available")
	}
	return hypervisorList[rand.Intn(len(hypervisorList))], nil
}
