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

// NewClusterFileMapFromv1alpha1 creates a cluster file map based on a
// versioned cluster file map
func NewClusterFileMapFromv1alpha1(clusterFileMap infrav1alpha1.ClusterFileMap) ClusterFileMap {
	res := ClusterFileMap{}
	for clusterName, componentFileMap := range clusterFileMap {
		if res[clusterName] == nil {
			res[clusterName] = ComponentFileMap{}
		}
		for componentName, fileMap := range componentFileMap {
			if res[clusterName][componentName] == nil {
				res[clusterName][componentName] = FileMap{}
			}
			for fileName, fileSum := range fileMap {
				res[clusterName][componentName][fileName] = fileSum
			}
		}
	}
	return res
}

// Export exports this cluster file map to a versioned cluster file map
func (clusterFileMap ClusterFileMap) Export() infrav1alpha1.ClusterFileMap {
	res := infrav1alpha1.ClusterFileMap{}
	for clusterName, componentFileMap := range clusterFileMap {
		if res[clusterName] == nil {
			res[clusterName] = infrav1alpha1.ComponentFileMap{}
		}
		for componentName, fileMap := range componentFileMap {
			if res[clusterName][componentName] == nil {
				res[clusterName][componentName] = infrav1alpha1.FileMap{}
			}
			for fileName, fileSum := range fileMap {
				res[clusterName][componentName][fileName] = fileSum
			}
		}
	}
	return res
}
