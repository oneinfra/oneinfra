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

package constants

import (
	"log"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// ReleaseInfo represents a list of supported component versions
type ReleaseInfo struct {
	KubernetesVersions []KubernetesVersion `json:"kubernetesVersions"`
}

// KubernetesVersion represents a supported Kubernetes version
type KubernetesVersion struct {
	KubernetesVersion string `json:"kubernetesVersion"`
	CRIToolsVersion   string `json:"criToolsVersion"`
	ContainerdVersion string `json:"containerdVersion"`
	CNIPluginsVersion string `json:"cniPluginsVersion"`
	EtcdVersion       string `json:"etcdVersion"`
	PauseVersion      string `json:"pauseVersion"`
}

var (
	// ReleaseData includes all release information
	ReleaseData          *ReleaseInfo
	kubernetesVersionMap map[string]KubernetesVersion
	// LatestKubernetesVersion has the latest Kubernetes version
	LatestKubernetesVersion string
)

func init() {
	var currReleaseData ReleaseInfo
	if err := yaml.Unmarshal([]byte(RawReleaseData), &currReleaseData); err != nil {
		log.Fatalf("could not unmarshal RELEASE file contents: %v", err)
	}
	ReleaseData = &currReleaseData
	kubernetesVersionMap = map[string]KubernetesVersion{}
	for _, kubernetesVersion := range ReleaseData.KubernetesVersions {
		kubernetesVersionMap[kubernetesVersion.KubernetesVersion] = kubernetesVersion
		LatestKubernetesVersion = kubernetesVersion.KubernetesVersion
	}
}

// KubernetesVersionBundle returns the KubernetesVersion for the
// provided version
func KubernetesVersionBundle(version string) (*KubernetesVersion, error) {
	if kubernetesVersion, exists := kubernetesVersionMap[version]; exists {
		return &kubernetesVersion, nil
	}
	return nil, errors.Errorf("could not find Kubernetes version %q in the known versions", version)
}
