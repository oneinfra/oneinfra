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

	"sigs.k8s.io/yaml"

	versionsapi "github.com/oneinfra/oneinfra/pkg/versions"
)

var (
	// ReleaseData includes all release versioning information
	ReleaseData *versionsapi.ReleaseInfo
	// TestData includes all test versioning information
	TestData *TestInfo
	// KubernetesVersions has a map of the supported Kubernetes versions
	KubernetesVersions map[string]versionsapi.KubernetesVersion
	// ContainerdTestVersions has a map of the testing containerd versions
	ContainerdTestVersions map[string]ContainerdVersionDependencies
	// KubernetesTestVersions has a map of the testing kubernetes
	// component versions
	KubernetesTestVersions map[string]KubernetesVersionDependencies
)

func init() {
	initReleaseData()
	initTestData()
}

func initReleaseData() {
	var currReleaseData versionsapi.ReleaseInfo
	if err := yaml.Unmarshal([]byte(RawReleaseData), &currReleaseData); err != nil {
		log.Fatalf("could not unmarshal RELEASE file contents: %v", err)
	}
	ReleaseData = &currReleaseData
	KubernetesVersions = map[string]versionsapi.KubernetesVersion{}
	for _, kubernetesVersion := range ReleaseData.KubernetesVersions {
		KubernetesVersions[kubernetesVersion.Version] = kubernetesVersion
	}
}

func initTestData() {
	var currTestData TestInfo
	if err := yaml.Unmarshal([]byte(RawTestData), &currTestData); err != nil {
		log.Fatalf("could not unmarshal RELEASE_TEST file contents: %v", err)
	}
	TestData = &currTestData
	ContainerdTestVersions = map[string]ContainerdVersionDependencies{}
	for _, containerdVersion := range TestData.ContainerdVersions {
		ContainerdTestVersions[containerdVersion.Version] = containerdVersion
	}
	KubernetesTestVersions = map[string]KubernetesVersionDependencies{}
	for kubernetesVersion, kubernetesData := range TestData.KubernetesVersions {
		KubernetesTestVersions[kubernetesVersion] = kubernetesData
	}
}

//go:generate sh -c "CONST_NAME=RawReleaseData RELEASE_FILE=RELEASE RELEASE_DATA_FILE=internal/pkg/constants/zz_generated.release_data.constants.go ../../../scripts/release-gen.sh"
//go:generate sh -c "CONST_NAME=RawTestData RELEASE_FILE=RELEASE_TEST RELEASE_DATA_FILE=internal/pkg/constants/zz_generated.test_data.go ../../../scripts/release-gen.sh"
