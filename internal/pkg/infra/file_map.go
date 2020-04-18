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

package infra

import (
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
)

// FileMap is a map of file paths as keys and their sum as values
type FileMap map[string]string

// ComponentFileMap is a map of filemaps, with component as keys, and
// filemaps as values
type ComponentFileMap map[string]FileMap

// ClusterFileMap is a map of component filemaps, with clusters as
// keys, and component filemaps as values
type ClusterFileMap map[string]ComponentFileMap

// NamespacedClusterFileMap is a map of cluster filemaps, with
// namespaces as keys, and cluster filemaps as values
type NamespacedClusterFileMap map[string]ClusterFileMap

// NewNamespacedClusterFileMapFromv1alpha1 creates a namespaced
// cluster file map based on a versioned cluster file map
func NewNamespacedClusterFileMapFromv1alpha1(namespacedClusterFileMap infrav1alpha1.NamespacedClusterFileMap) NamespacedClusterFileMap {
	res := NamespacedClusterFileMap{}
	for namespaceName, clusterFileMap := range namespacedClusterFileMap {
		if res[namespaceName] == nil {
			res[namespaceName] = ClusterFileMap{}
		}
		for clusterName, componentFileMap := range clusterFileMap {
			if res[namespaceName][clusterName] == nil {
				res[namespaceName][clusterName] = ComponentFileMap{}
			}
			for componentName, fileMap := range componentFileMap {
				if res[namespaceName][clusterName][componentName] == nil {
					res[namespaceName][clusterName][componentName] = FileMap{}
				}
				for fileName, fileSum := range fileMap {
					res[namespaceName][clusterName][componentName][fileName] = fileSum
				}
			}
		}
	}
	return res
}

// Export exports this cluster file map to a versioned cluster file map
func (namespacedClusterFileMap NamespacedClusterFileMap) Export() infrav1alpha1.NamespacedClusterFileMap {
	res := infrav1alpha1.NamespacedClusterFileMap{}
	for namespaceName, clusterFileMap := range namespacedClusterFileMap {
		if res[namespaceName] == nil {
			res[namespaceName] = infrav1alpha1.ClusterFileMap{}
		}
		for clusterName, componentFileMap := range clusterFileMap {
			if res[namespaceName][clusterName] == nil {
				res[namespaceName][clusterName] = infrav1alpha1.ComponentFileMap{}
			}
			for componentName, fileMap := range componentFileMap {
				if res[namespaceName][clusterName][componentName] == nil {
					res[namespaceName][clusterName][componentName] = infrav1alpha1.FileMap{}
				}
				for fileName, fileSum := range fileMap {
					res[namespaceName][clusterName][componentName][fileName] = fileSum
				}
			}
		}
	}
	return res
}
