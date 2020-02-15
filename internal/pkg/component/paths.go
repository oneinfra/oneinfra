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

package component

import (
	"path/filepath"
)

func secretsPath(clusterName string) string {
	return filepath.Join("/etc/oneinfra/clusters", clusterName)
}

func storagePath(clusterName string) string {
	return filepath.Join("/var/lib/oneinfra/clusters", clusterName)
}

func secretsPathFile(clusterName, file string) string {
	return filepath.Join(secretsPath(clusterName), file)
}
