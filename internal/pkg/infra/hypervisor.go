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
	"k8s.io/klog"

	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
)

const (
	toolingImage           = "oneinfra/tooling:latest"
	podSandboxSHA1SumLabel = "oneinfra/pod-sha1sum"
	clusterNameLabel       = "oneinfra/cluster-name"
	componentNameLabel     = "oneinfra/component-name"
)

// Hypervisor represents an hypervisor
type Hypervisor struct {
	Name               string
	Namespace          string
	ResourceVersion    string
	Labels             map[string]string
	Annotations        map[string]string
	Public             bool
	IPAddress          string
	Files              ClusterFileMap
	Endpoint           hypervisorEndpoint
	criRuntime         criapi.RuntimeServiceClient
	criImage           criapi.ImageServiceClient
	portRangeLow       int
	portRangeHigh      int
	freedPorts         []int
	allocatedPorts     HypervisorPortAllocationList
	loadedContentsHash string
}

// HypervisorMap represents a map of hypervisors
type HypervisorMap map[string]*Hypervisor

// HypervisorList represents a list of hypervisors
type HypervisorList []*Hypervisor

// NewHypervisorFromv1alpha1 returns an hypervisor based on a versioned hypervisor
func NewHypervisorFromv1alpha1(hypervisor *infrav1alpha1.Hypervisor) (*Hypervisor, error) {
	hypervisorFiles := hypervisor.Status.Files
	if hypervisorFiles == nil {
		hypervisorFiles = infrav1alpha1.ClusterFileMap{}
	}
	res := Hypervisor{
		Name:            hypervisor.Name,
		Namespace:       hypervisor.Namespace,
		ResourceVersion: hypervisor.ResourceVersion,
		Labels:          hypervisor.Labels,
		Annotations:     hypervisor.Annotations,
		Public:          hypervisor.Spec.Public,
		IPAddress:       hypervisor.Spec.IPAddress,
		Files:           NewClusterFileMapFromv1alpha1(hypervisorFiles),
		portRangeLow:    hypervisor.Spec.PortRange.Low,
		portRangeHigh:   hypervisor.Spec.PortRange.High,
		freedPorts:      hypervisor.Status.FreedPorts,
		allocatedPorts:  NewHypervisorPortAllocationListFromv1alpha1(hypervisor.Status.AllocatedPorts),
	}
	if err := setHypervisorEndpointFromv1alpha1(hypervisor, &res); err != nil {
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
func (hypervisor *Hypervisor) PodSandboxConfig(cluster *cluster.Cluster, pod podapi.Pod) (criapi.PodSandboxConfig, error) {
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
	podSandboxConfig := criapi.PodSandboxConfig{
		Metadata: &criapi.PodSandboxMetadata{
			Name: pod.Name,
			Uid:  podSum,
		},
		Labels: map[string]string{
			componentNameLabel:     pod.Name,
			podSandboxSHA1SumLabel: podSum,
		},
		PortMappings: portMappings,
		LogDirectory: "/var/log/pods/",
	}
	if cluster != nil {
		podSandboxConfig.Labels[clusterNameLabel] = cluster.Name
		podSandboxConfig.Metadata.Namespace = fmt.Sprintf("%s-%s-%s", cluster.Name, pod.Name, podSandboxConfig.Metadata.Uid)
	} else {
		podSandboxConfig.Metadata.Namespace = fmt.Sprintf("%s-%s", pod.Name, podSandboxConfig.Metadata.Uid)
	}
	if pod.Privileges == podapi.PrivilegesNetworkPrivileged {
		podSandboxConfig.Linux = &criapi.LinuxPodSandboxConfig{
			SecurityContext: &criapi.LinuxSandboxSecurityContext{
				Privileged: true,
				NamespaceOptions: &criapi.NamespaceOption{
					Network: criapi.NamespaceMode_NODE,
				},
			},
		}
	}
	return podSandboxConfig, nil
}

// IsPodRunning returns whether a pod is running on the current hypervisor
func (hypervisor *Hypervisor) IsPodRunning(pod podapi.Pod) (bool, string, error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return false, "", err
	}
	podSum, err := pod.SHA1Sum()
	if err != nil {
		return false, "", err
	}
	klog.V(2).Infof("checking if a pod %q in hypervisor %q is running", pod.Name, hypervisor.Name)
	podSandboxList, err := criRuntime.ListPodSandbox(
		context.TODO(),
		&criapi.ListPodSandboxRequest{
			Filter: &criapi.PodSandboxFilter{
				LabelSelector: map[string]string{
					podSandboxSHA1SumLabel: podSum,
				},
			},
		},
	)
	if err != nil {
		return false, "", err
	}
	podSandboxID := ""
	for _, podSandbox := range podSandboxList.Items {
		if podSandbox.State == criapi.PodSandboxState_SANDBOX_READY {
			podSandboxID = podSandbox.Id
			break
		}
	}
	if len(podSandboxID) == 0 {
		return false, "", nil
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
		return false, podSandboxID, err
	}
	for _, container := range containerList.Containers {
		containerStatus, err := criRuntime.ContainerStatus(
			context.TODO(),
			&criapi.ContainerStatusRequest{
				ContainerId: container.Id,
			},
		)
		if err != nil {
			return false, podSandboxID, err
		}
		if containerStatus.Status.State != criapi.ContainerState_CONTAINER_RUNNING {
			return false, podSandboxID, nil
		}
	}
	return true, podSandboxID, nil
}

// RunPod runs a pod on the current hypervisor
func (hypervisor *Hypervisor) RunPod(cluster *cluster.Cluster, pod podapi.Pod) (string, error) {
	isPodRunning, podSandboxID, err := hypervisor.IsPodRunning(pod)
	if err != nil {
		return "", err
	}
	if isPodRunning {
		klog.V(2).Infof("all containers within pod %q in hypervisor %q are running", pod.Name, hypervisor.Name)
		return podSandboxID, nil
	}
	return hypervisor.runPodInNewSandbox(cluster, pod)
}

func (hypervisor *Hypervisor) runPodInNewSandbox(cluster *cluster.Cluster, pod podapi.Pod) (string, error) {
	klog.V(2).Infof("running pod %q in hypervisor %q", pod.Name, hypervisor.Name)
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return "", err
	}
	podSandboxConfig, err := hypervisor.PodSandboxConfig(cluster, pod)
	if err != nil {
		return "", err
	}
	podSandboxResponse, err := criRuntime.RunPodSandbox(
		context.TODO(),
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
		if container.Privileges == podapi.PrivilegesNetworkPrivileged {
			createContainerRequest.Config.Linux = &criapi.LinuxContainerConfig{
				SecurityContext: &criapi.LinuxContainerSecurityContext{
					Privileged: true,
					NamespaceOptions: &criapi.NamespaceOption{
						Network: criapi.NamespaceMode_NODE,
					},
				},
			}
		}
		containerResponse, err := criRuntime.CreateContainer(
			context.TODO(),
			&createContainerRequest,
		)
		if err != nil {
			return "", err
		}
		containerIds = append(containerIds, containerResponse.ContainerId)
	}
	for _, containerID := range containerIds {
		_, err = criRuntime.StartContainer(
			context.TODO(),
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
// provided cluster and component
func (hypervisor *Hypervisor) ListPods(clusterName, componentName string) ([]string, error) {
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return []string{}, err
	}
	podSandboxList, err := criRuntime.ListPodSandbox(
		context.TODO(),
		&criapi.ListPodSandboxRequest{
			Filter: &criapi.PodSandboxFilter{
				LabelSelector: map[string]string{
					clusterNameLabel:   clusterName,
					componentNameLabel: componentName,
				},
			},
		},
	)
	if err != nil {
		return []string{}, errors.Errorf("could not list pods for cluster %q and component %q", clusterName, componentName)
	}
	res := []string{}
	for _, podSandbox := range podSandboxList.Items {
		res = append(res, podSandbox.Id)
	}
	return res, nil
}

// DeletePod deletes a pod on the current hypervisor
func (hypervisor *Hypervisor) DeletePod(podSandboxID string) error {
	klog.V(2).Infof("deleting pod %q from hypervisor %q", podSandboxID, hypervisor.Name)
	criRuntime, err := hypervisor.CRIRuntime()
	if err != nil {
		return err
	}
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
func (hypervisor *Hypervisor) RunAndWaitForPod(cluster *cluster.Cluster, pod podapi.Pod) error {
	podSandboxID, err := hypervisor.RunPod(cluster, pod)
	if err != nil {
		return err
	}
	if err := hypervisor.WaitForPod(podSandboxID); err != nil {
		return err
	}
	return hypervisor.DeletePod(podSandboxID)
}

// UploadFiles uploads a map of files, with location as keys, and
// contents as values
func (hypervisor *Hypervisor) UploadFiles(clusterName, componentName string, files map[string]string) error {
	if err := hypervisor.EnsureImage(toolingImage); err != nil {
		return err
	}
	for fileLocation, fileContents := range files {
		if err := hypervisor.uploadFile(clusterName, componentName, fileLocation, fileContents); err != nil {
			return err
		}
	}
	return nil
}

// UploadFile uploads a file to the current hypervisor to hostPath
// with given fileContents
func (hypervisor *Hypervisor) UploadFile(clusterName, componentName, hostPath, fileContents string) error {
	if err := hypervisor.EnsureImage(toolingImage); err != nil {
		return err
	}
	return hypervisor.uploadFile(clusterName, componentName, hostPath, fileContents)
}

// FileUpToDate returns whether the given file contents match on the
// host
func (hypervisor *Hypervisor) FileUpToDate(clusterName, componentName, hostPath, fileContents string) bool {
	fileContentsSHA1 := fmt.Sprintf("%x", sha1.Sum([]byte(fileContents)))
	if currentFileContentsSHA1, exists := hypervisor.Files[clusterName][componentName][hostPath]; exists {
		if currentFileContentsSHA1 == fileContentsSHA1 {
			return true
		}
	}
	return false
}

func (hypervisor *Hypervisor) uploadFile(clusterName, componentName, hostPath, fileContents string) error {
	if hypervisor.FileUpToDate(clusterName, componentName, hostPath, fileContents) {
		klog.V(2).Infof("skipping file upload to hypervisor %q at location %q, hash matches", hypervisor.Name, hostPath)
		return nil
	}
	klog.V(2).Infof("uploading file to hypervisor %q at location %q", hypervisor.Name, hostPath)
	hostPathDir := filepath.Dir(hostPath)
	uploadFilePod := podapi.NewSingleContainerPod(
		fmt.Sprintf("upload-file-%x", md5.Sum([]byte(fileContents))),
		toolingImage,
		[]string{"write-base64-file.sh"},
		[]string{
			base64.StdEncoding.EncodeToString([]byte(fileContents)),
			hostPath,
		},
		map[string]string{hostPathDir: hostPathDir},
		map[int]int{},
		podapi.PrivilegesUnprivileged,
	)
	podSandboxID, err := hypervisor.runPodInNewSandbox(nil, uploadFilePod)
	if err != nil {
		return err
	}
	if err := hypervisor.WaitForPod(podSandboxID); err != nil {
		return err
	}
	if hypervisor.Files != nil {
		if hypervisor.Files[clusterName] == nil {
			hypervisor.Files[clusterName] = ComponentFileMap{}
		}
		if hypervisor.Files[clusterName][componentName] == nil {
			hypervisor.Files[clusterName][componentName] = FileMap{}
		}
		hypervisor.Files[clusterName][componentName][hostPath] = fmt.Sprintf("%x", sha1.Sum([]byte(fileContents)))
	}
	return hypervisor.DeletePod(podSandboxID)
}

// HasPort returns whether a port exists for the given clusterName and componentName
func (hypervisor *Hypervisor) HasPort(clusterName, componentName string) (bool, int) {
	for _, allocatedPort := range hypervisor.allocatedPorts {
		if allocatedPort.Cluster == clusterName && allocatedPort.Component == componentName {
			return true, allocatedPort.Port
		}
	}
	return false, 0
}

// RequestPort requests a port on the current hypervisor
func (hypervisor *Hypervisor) RequestPort(clusterName, componentName string) (int, error) {
	if hasPort, existingPort := hypervisor.HasPort(clusterName, componentName); hasPort {
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
		Cluster:   clusterName,
		Component: componentName,
		Port:      newPort,
	})
	klog.V(2).Infof("port requested for hypervisor %q; assigned: %d", hypervisor.Name, newPort)
	return newPort, nil
}

// Export exports the hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	resHypervisor := infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name:            hypervisor.Name,
			Namespace:       hypervisor.Namespace,
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
