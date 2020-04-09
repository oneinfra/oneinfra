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

package images

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

type ContainerImage string
type ContainerImageMapWithTags map[ContainerImage][]string

const (
	namespace string = "oneinfra"
)

const (
	containerd       ContainerImage = "containerd"
	hypervisor       ContainerImage = "hypervisor"
	kubeletInstaller ContainerImage = "kubelet-installer"
	oi               ContainerImage = "oi"
	oiManager        ContainerImage = "oi-manager"
)

// BuildContainerImages builds the container images to be published
func BuildContainerImages(chosenContainerImages ContainerImageMapWithTags) {
	executeForEachContainerImage(chosenContainerImages, shouldBuildImage, buildImage)
}

func shouldBuildImage(containerImage ContainerImage, containerVersion string) bool {
	return exec.Command("docker", "inspect", fmt.Sprintf("%s/%s:%s", namespace, containerImage, containerVersion)).Run() != nil
}

func buildImage(containerImage ContainerImage, containerVersion string) *exec.Cmd {
	return exec.Command("make", "image")
}

// PublishContainerImages publishes the container images
func PublishContainerImages(chosenContainerImages ContainerImageMapWithTags) {
	executeForEachContainerImage(
		chosenContainerImages,
		func(containerImage ContainerImage, containerVersion string) bool {
			resp, err := http.Get(fmt.Sprintf("https://index.docker.io/v1/repositories/%s/tags/%s", fmt.Sprintf("%s/%s", namespace, containerImage), containerVersion))
			if err != nil {
				klog.Warningf("could not check if image %s/%s:%s exists in registry: %v", namespace, containerImage, containerVersion, err)
				return false
			}
			if resp.StatusCode == 200 {
				klog.Infof("image %s/%s:%s exists; will not publish image", namespace, containerImage, containerVersion)
				return false
			}
			klog.Infof("got response %d from the registry; will publish image %s/%s:%s", resp.StatusCode, namespace, containerImage, containerVersion)
			return true
		},
		func(containerImage ContainerImage, containerVersion string) *exec.Cmd {
			if shouldBuildImage(containerImage, containerVersion) {
				if err := rawExecuteForContainerImage(containerImage, containerVersion, buildImage); err != nil {
					klog.Warning("could not build image %q", fmt.Sprintf("%s:%s", containerImage, containerVersion))
				}
			}
			return exec.Command("make", "publish")
		},
	)
}

func executeForEachContainerImage(chosenContainerImages ContainerImageMapWithTags, shouldDo func(ContainerImage, string) bool, do func(ContainerImage, string) *exec.Cmd) {
	containerImages := containerImagesFromChosen(chosenContainerImages)
	for containerImage, containerVersions := range containerImages {
		for _, containerVersion := range containerVersions {
			if shouldDo(containerImage, containerVersion) {
				if err := executeForContainerImage(containerImage, containerVersion, do); err != nil {
					log.Printf("failed to execute command for image %q: %v", containerImage, err)
				}
			}
		}
	}
}

func executeForContainerImage(containerImage ContainerImage, containerVersion string, do func(ContainerImage, string) *exec.Cmd) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Errorf("could not read current working directory: %v", err)
	}
	if err := os.Chdir(filepath.Join(cwd, "images", string(containerImage))); err != nil {
		return errors.Errorf("could not change directory: %v", err)
	}
	return rawExecuteForContainerImage(containerImage, containerVersion, do)
}

func rawExecuteForContainerImage(containerImage ContainerImage, containerVersion string, do func(ContainerImage, string) *exec.Cmd) error {
	cmd := do(containerImage, containerVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setCmdEnv(cmd, containerImage, containerVersion)
	return cmd.Run()
}

func setCmdEnv(cmd *exec.Cmd, containerImage ContainerImage, containerVersion string) {
	cmd.Env = os.Environ()
	switch containerImage {
	case containerd:
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("CONTAINERD_VERSION=%s", containerVersion),
			fmt.Sprintf("CRI_TOOLS_VERSION=%s", constants.ReleaseData.ContainerdVersions[containerVersion].CRIToolsVersion),
			fmt.Sprintf("CNI_PLUGINS_VERSION=%s", constants.ReleaseData.ContainerdVersions[containerVersion].CNIPluginsVersion),
		}...)
	case hypervisor:
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("KUBERNETES_VERSION=%s", containerVersion),
			fmt.Sprintf("CONTAINERD_VERSION=%s", constants.ReleaseData.KubernetesVersions[containerVersion].ContainerdVersion),
			fmt.Sprintf("ETCD_VERSION=%s", constants.ReleaseData.KubernetesVersions[containerVersion].EtcdVersion),
			fmt.Sprintf("PAUSE_VERSION=%s", constants.ReleaseData.KubernetesVersions[containerVersion].PauseVersion),
		}...)
	case kubeletInstaller:
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("KUBERNETES_VERSION=%s", containerVersion),
		}...)
	case oi, oiManager:
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("ONEINFRA_VERSION=%s", containerVersion),
		}...)
	}
}
