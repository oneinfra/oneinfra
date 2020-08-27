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
	"k8s.io/klog/v2"
)

// ContainerImage is a container image name
type ContainerImage string

// ContainerImageWithTags is a type representing a container image
type ContainerImageWithTags struct {
	Image ContainerImage
	Tags  []string
}

// ContainerImageList is a list of container images and a list of tags
// for each image
type ContainerImageList []ContainerImageWithTags

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
func BuildContainerImages(chosenContainerImages ContainerImageList, forceBuild bool) {
	executeForEachContainerImage(chosenContainerImages, shouldBuildImage(forceBuild), buildImage)
}

func shouldBuildImage(forceBuild bool) func(containerImage ContainerImage, containerVersion string) bool {
	if forceBuild {
		return func(_ ContainerImage, _ string) bool {
			return true
		}
	}
	return func(containerImage ContainerImage, containerVersion string) bool {
		return exec.Command("docker", "inspect", fmt.Sprintf("%s/%s:%s", namespace, containerImage, containerVersion)).Run() != nil
	}
}

func buildImage(containerImage ContainerImage, containerVersion string) *exec.Cmd {
	return exec.Command("make", "image")
}

// PublishContainerImages publishes the container images
func PublishContainerImages(chosenContainerImages ContainerImageList, forcePublish bool) {
	executeForEachContainerImage(
		chosenContainerImages,
		func(containerImage ContainerImage, containerVersion string) bool {
			if forcePublish {
				klog.Info("force publish was set, building image")
				return true
			}
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
			if shouldBuildImage(forcePublish)(containerImage, containerVersion) {
				if err := rawExecuteForContainerImage(containerImage, containerVersion, buildImage); err != nil {
					klog.Warningf("could not build image %s", fmt.Sprintf("%s/%s:%s", namespace, containerImage, containerVersion))
				}
			}
			return exec.Command("make", "publish")
		},
	)
}

func executeForEachContainerImage(chosenContainerImages ContainerImageList, shouldDo func(ContainerImage, string) bool, do func(ContainerImage, string) *exec.Cmd) {
	containerImages := containerImagesFromChosen(chosenContainerImages)
	for _, containerImage := range containerImages {
		for _, containerTag := range containerImage.Tags {
			if shouldDo(containerImage.Image, containerTag) {
				if err := executeForContainerImage(containerImage.Image, containerTag, do); err != nil {
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
	res := rawExecuteForContainerImage(containerImage, containerVersion, do)
	if err := os.Chdir(cwd); err != nil {
		return errors.Errorf("could not change directory: %v", err)
	}
	return res
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
			fmt.Sprintf("CRI_TOOLS_VERSION=%s", constants.ContainerdTestVersions[containerVersion].CRIToolsVersion),
			fmt.Sprintf("CNI_PLUGINS_VERSION=%s", constants.ContainerdTestVersions[containerVersion].CNIPluginsVersion),
		}...)
	case hypervisor:
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("KUBERNETES_VERSION=%s", containerVersion),
			fmt.Sprintf("CONTAINERD_VERSION=%s", constants.KubernetesTestVersions[containerVersion].ContainerdVersion),
			fmt.Sprintf("ETCD_VERSION=%s", constants.KubernetesVersions[containerVersion].EtcdVersion),
			fmt.Sprintf("PAUSE_VERSION=%s", constants.KubernetesTestVersions[containerVersion].PauseVersion),
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
