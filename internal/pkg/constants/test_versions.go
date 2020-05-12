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
	releasecomponents "github.com/oneinfra/oneinfra/internal/pkg/release-components"
	"github.com/pkg/errors"
)

// TestInfo represents a list of supported component versions for
// testing
type TestInfo struct {
	ContainerdVersions []ContainerdVersionDependencies          `json:"containerdVersions"`
	KubernetesVersions map[string]KubernetesVersionDependencies `json:"kubernetesVersions"`
}

// KubernetesVersionDependencies represents a supported Kubernetes
// version for testing
type KubernetesVersionDependencies struct {
	ContainerdVersion string `json:"containerdVersion"`
	PauseVersion      string `json:"pauseVersion"`
}

// ContainerdVersionDependencies represents a supported containerd
// version for testing
type ContainerdVersionDependencies struct {
	Version           string `json:"version"`
	CRIToolsVersion   string `json:"criToolsVersion"`
	CNIPluginsVersion string `json:"cniPluginsVersion"`
}

// KubernetesVersionTestDependencyBundle returns the
// KubernetesVersionDependencies for the provided test version
func KubernetesVersionTestDependencyBundle(version string) (*KubernetesVersionDependencies, error) {
	if kubernetesVersion, exists := KubernetesTestVersions[version]; exists {
		return &kubernetesVersion, nil
	}
	return nil, errors.Errorf("could not find Kubernetes test version %q in the known versions", version)
}

// KubernetesTestComponentVersion returns the test component version
// for the given Kubernetes version and component
func KubernetesTestComponentVersion(version string, component releasecomponents.KubernetesTestComponent) (string, error) {
	kubernetesVersionBundle, err := KubernetesVersionTestDependencyBundle(version)
	if err != nil {
		return "", err
	}
	switch component {
	case releasecomponents.CRITools:
		return ContainerdTestVersions[kubernetesVersionBundle.ContainerdVersion].CRIToolsVersion, nil
	case releasecomponents.Containerd:
		return kubernetesVersionBundle.ContainerdVersion, nil
	case releasecomponents.CNIPlugins:
		return ContainerdTestVersions[kubernetesVersionBundle.ContainerdVersion].CNIPluginsVersion, nil
	case releasecomponents.Pause:
		return kubernetesVersionBundle.PauseVersion, nil
	}
	return "", errors.Errorf("could not find component %q in version %q", component, version)
}
