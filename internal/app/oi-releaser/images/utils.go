/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

func containerImagesFromChosen(chosenContainerImages ContainerImageList) ContainerImageList {
	if len(chosenContainerImages) > 0 {
		return chosenContainerImages
	}
	res := ContainerImageList{}
	containerdImageWithTags := ContainerImageWithTags{
		Image: containerd,
		Tags:  []string{},
	}
	for containerdVersion := range constants.ContainerdTestVersions {
		containerdImageWithTags.Tags = append(
			containerdImageWithTags.Tags,
			containerdVersion,
		)
	}
	if len(containerdImageWithTags.Tags) > 0 {
		res = append(res, containerdImageWithTags)
	}
	hypervisorImageWithTags := ContainerImageWithTags{
		Image: hypervisor,
		Tags:  []string{},
	}
	kubeletInstallerImageWithTags := ContainerImageWithTags{
		Image: kubeletInstaller,
		Tags:  []string{},
	}
	for kubernetesVersion := range constants.KubernetesVersions {
		hypervisorImageWithTags.Tags = append(
			hypervisorImageWithTags.Tags,
			kubernetesVersion,
		)
		kubeletInstallerImageWithTags.Tags = append(
			kubeletInstallerImageWithTags.Tags,
			kubernetesVersion,
		)
	}
	if len(hypervisorImageWithTags.Tags) > 0 {
		res = append(res, hypervisorImageWithTags)
	}
	if len(kubeletInstallerImageWithTags.Tags) > 0 {
		res = append(res, kubeletInstallerImageWithTags)
	}
	res = append(res, ContainerImageWithTags{Image: oi, Tags: []string{constants.BuildVersion}})
	res = append(res, ContainerImageWithTags{Image: oiManager, Tags: []string{constants.BuildVersion}})
	return res
}
