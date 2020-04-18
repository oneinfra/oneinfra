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

package components

import (
	"path/filepath"
)

func globalSecretsPath() string {
	return "/etc/oneinfra/clusters"
}

func clusterSecretsPath(clusterName string) string {
	return filepath.Join(globalSecretsPath(), clusterName)
}

func secretsPath(clusterName, componentName string) string {
	return filepath.Join(clusterSecretsPath(clusterName), componentName)
}

func globalStoragePath() string {
	return "/var/lib/oneinfra/clusters"
}

func clusterStoragePath(clusterName string) string {
	return filepath.Join(globalStoragePath(), clusterName)
}

func componentStoragePath(clusterName, componentName string) string {
	return filepath.Join(clusterStoragePath(clusterName), componentName)
}

func storagePath(clusterName, componentName, subcomponentName string) string {
	return filepath.Join(componentStoragePath(clusterName, componentName), subcomponentName)
}

func secretsPathFile(clusterName, componentName, file string) string {
	return filepath.Join(secretsPath(clusterName, componentName), file)
}
