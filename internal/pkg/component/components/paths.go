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

package components

import (
	"path/filepath"
)

func globalSecretsPath() string {
	return "/etc/oneinfra/clusters"
}

func globalStoragePath() string {
	return "/var/lib/oneinfra/clusters"
}

func namespacedClusterSecretsPath(clusterNamespace string) string {
	return filepath.Join(globalSecretsPath(), clusterNamespace)
}

func clusterSecretsPath(clusterNamespace, clusterName string) string {
	return filepath.Join(namespacedClusterSecretsPath(clusterNamespace), clusterName)
}

func componentSecretsPath(clusterNamespace, clusterName, componentName string) string {
	return filepath.Join(clusterSecretsPath(clusterNamespace, clusterName), componentName)
}

func componentSecretsPathFile(clusterNamespace, clusterName, componentName, file string) string {
	return filepath.Join(componentSecretsPath(clusterNamespace, clusterName, componentName), file)
}

func namespacedClusterStoragePath(clusterNamespace string) string {
	return filepath.Join(globalStoragePath(), clusterNamespace)
}

func clusterStoragePath(clusterNamespace, clusterName string) string {
	return filepath.Join(namespacedClusterStoragePath(clusterNamespace), clusterName)
}

func componentStoragePath(clusterNamespace, clusterName, componentName string) string {
	return filepath.Join(clusterStoragePath(clusterNamespace, clusterName), componentName)
}

func subcomponentStoragePath(clusterNamespace, clusterName, componentName, subcomponentName string) string {
	return filepath.Join(componentStoragePath(clusterNamespace, clusterName, componentName), subcomponentName)
}
