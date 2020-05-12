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
	"context"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	releasecomponents "github.com/oneinfra/oneinfra/internal/pkg/release-components"
	constantsapi "github.com/oneinfra/oneinfra/pkg/constants"
	"github.com/oneinfra/oneinfra/pkg/versions"
	"github.com/pkg/errors"
)

// KubernetesVersionBundle returns the KubernetesVersion for the
// provided version
func KubernetesVersionBundle(version string) (*versions.KubernetesVersion, error) {
	if kubernetesVersion, exists := KubernetesVersions[version]; exists {
		return &kubernetesVersion, nil
	}
	return nil, errors.Errorf("could not find Kubernetes version %q in the known versions", version)
}

// KubernetesComponentVersion returns the component version for the
// given Kubernetes version and component
func KubernetesComponentVersion(version string, component releasecomponents.KubernetesComponent) (string, error) {
	kubernetesVersionBundle, err := KubernetesVersionBundle(version)
	if err != nil {
		return "", err
	}
	switch component {
	case releasecomponents.Etcd:
		return kubernetesVersionBundle.EtcdVersion, nil
	case releasecomponents.CoreDNS:
		return kubernetesVersionBundle.CoreDNSVersion, nil
	}
	return "", errors.Errorf("could not find component %q in version %q", component, version)
}

// UpdateOneInfraVersionsConfigMap creates or updates the oneinfra
// version ConfigMap
func UpdateOneInfraVersionsConfigMap(ctx context.Context, client client.Client) error {
	versionConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constantsapi.OneInfraVersionsConfigMap,
			Namespace: constantsapi.OneInfraNamespace,
		},
		Data: map[string]string{
			constantsapi.OneInfraVersionsKeyName: RawReleaseData,
		},
	}
	err := client.Create(ctx, versionConfigMap)
	if err == nil || !apierrors.IsAlreadyExists(err) {
		return err
	}
	return client.Update(ctx, versionConfigMap)
}
