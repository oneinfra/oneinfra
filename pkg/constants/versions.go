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
	ConsoleVersion           string              `json:"consoleVersion"`
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
