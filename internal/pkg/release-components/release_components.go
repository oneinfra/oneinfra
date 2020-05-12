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

package releasecomponents

// KubernetesComponent represents a versioned Kubernetes component
type KubernetesComponent string

// KubernetesTestComponent represents a versioned test component
type KubernetesTestComponent string

const (
	// Etcd is the etcd component
	Etcd KubernetesComponent = "etcd"
	// CoreDNS is the CoreDNS component
	CoreDNS KubernetesComponent = "coredns"
)

const (
	// CRITools is the CRI tools component
	CRITools KubernetesTestComponent = "cri-tools"
	// Containerd is the containerd component
	Containerd KubernetesTestComponent = "containerd"
	// CNIPlugins is the CNI plugins component
	CNIPlugins KubernetesTestComponent = "cni-plugins"
	// Pause is the pause component
	Pause KubernetesTestComponent = "pause"
)

var (
	// KubernetesComponents is the list of all versioned components
	KubernetesComponents = []KubernetesComponent{Etcd, CoreDNS}
)

var (
	// KubernetesTestComponents is the list of all versioned test
	// components
	KubernetesTestComponents = []KubernetesTestComponent{CRITools, Containerd, CNIPlugins, Pause}
)
