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
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

func containerImagesFromChosen(chosenContainerImages ContainerImageMapWithTags) ContainerImageMapWithTags {
	if len(chosenContainerImages) > 0 {
		return chosenContainerImages
	}
	res := ContainerImageMapWithTags{}
	for containerdVersion := range constants.ContainerdTestVersions {
		if res[containerd] == nil {
			res[containerd] = []string{}
		}
		res[containerd] = append(res[containerd], containerdVersion)
	}
	for kubernetesVersion := range constants.KubernetesVersions {
		if res[hypervisor] == nil {
			res[hypervisor] = []string{}
			res[kubeletInstaller] = []string{}
		}
		res[hypervisor] = append(res[hypervisor], kubernetesVersion)
		res[kubeletInstaller] = append(res[kubeletInstaller], kubernetesVersion)
	}
	res[oi] = []string{constants.ReleaseData.Version}
	res[oiManager] = []string{constants.ReleaseData.Version}
	return res
}
