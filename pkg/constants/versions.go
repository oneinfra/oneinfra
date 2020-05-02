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

package constants

import (
	"log"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

// Component represents a versioned component
type Component string

const (
	// CRITools is the CRI tools component
	CRITools Component = "cri-tools"
	// Containerd is the containerd component
	Containerd Component = "containerd"
	// CNIPlugins is the CNI plugins component
	CNIPlugins Component = "cni-plugins"
	// Etcd is the etcd component
	Etcd Component = "etcd"
	// Pause is the pause component
	Pause Component = "pause"
	// CoreDNS is the CoreDNS component
	CoreDNS Component = "coredns"
)

var (
	// KubernetesComponents is the list of all versioned components
	KubernetesComponents = []Component{CRITools, Containerd, CNIPlugins, Etcd, Pause}
)

// ReleaseInfo represents a list of supported component versions
type ReleaseInfo struct {
	Version                  string              `json:"version"`
	DefaultKubernetesVersion string              `json:"defaultKubernetesVersion"`
	ContainerdVersions       []ContainerdVersion `json:"containerdVersions"`
	KubernetesVersions       []KubernetesVersion `json:"kubernetesVersions"`
}

// ContainerdVersion represents a supported containerd version for testing
type ContainerdVersion struct {
	Version           string `json:"version"`
	CRIToolsVersion   string `json:"criToolsVersion"`
	CNIPluginsVersion string `json:"cniPluginsVersion"`
}

// KubernetesVersion represents a supported Kubernetes version
type KubernetesVersion struct {
	Version           string `json:"version"`
	ContainerdVersion string `json:"containerdVersion"`
	EtcdVersion       string `json:"etcdVersion"`
	PauseVersion      string `json:"pauseVersion"`
	CoreDNSVersion    string `json:"coreDNSVersion"`
}

var (
	// ReleaseData includes all release information
	ReleaseData *ReleaseInfo
	// ContainerdVersions has a map of the test containerd versions
	ContainerdVersions map[string]ContainerdVersion
	// KubernetesVersions has a map of the supported Kubernetes versions
	KubernetesVersions map[string]KubernetesVersion
)

func init() {
	var currReleaseData ReleaseInfo
	if err := yaml.Unmarshal([]byte(RawReleaseData), &currReleaseData); err != nil {
		log.Fatalf("could not unmarshal RELEASE file contents: %v", err)
	}
	ReleaseData = &currReleaseData
	ContainerdVersions = map[string]ContainerdVersion{}
	for _, containerdVersion := range ReleaseData.ContainerdVersions {
		ContainerdVersions[containerdVersion.Version] = containerdVersion
	}
	KubernetesVersions = map[string]KubernetesVersion{}
	for _, kubernetesVersion := range ReleaseData.KubernetesVersions {
		KubernetesVersions[kubernetesVersion.Version] = kubernetesVersion
	}
}

// KubernetesVersionBundle returns the KubernetesVersion for the
// provided version
func KubernetesVersionBundle(version string) (*KubernetesVersion, error) {
	if kubernetesVersion, exists := KubernetesVersions[version]; exists {
		return &kubernetesVersion, nil
	}
	return nil, errors.Errorf("could not find Kubernetes version %q in the known versions", version)
}

// KubernetesComponentVersion returns the component version for the
// given Kubernetes version and component
func KubernetesComponentVersion(version string, component Component) (string, error) {
	kubernetesVersionBundle, err := KubernetesVersionBundle(version)
	if err != nil {
		return "", err
	}
	switch component {
	case CRITools:
		return ContainerdVersions[kubernetesVersionBundle.ContainerdVersion].CRIToolsVersion, nil
	case Containerd:
		return kubernetesVersionBundle.ContainerdVersion, nil
	case CNIPlugins:
		return ContainerdVersions[kubernetesVersionBundle.ContainerdVersion].CNIPluginsVersion, nil
	case Etcd:
		return kubernetesVersionBundle.EtcdVersion, nil
	case Pause:
		return kubernetesVersionBundle.PauseVersion, nil
	case CoreDNS:
		return kubernetesVersionBundle.CoreDNSVersion, nil
	}
	return "", errors.Errorf("could not find component %q in version %q", component, version)
}
