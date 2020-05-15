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

package manifests

import (
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	yamlutils "github.com/oneinfra/oneinfra/internal/pkg/yaml"
)

// ResourceActuator represents a function that takes an hypervisor
// map, a cluster map and a component list and performs some kind of
// logic with them
type ResourceActuator func(infra.HypervisorMap, cluster.Map, component.List) (component.List, error)

// WithStdinResources is a function that will take hypervisors,
// clusters and components from stdin and call to a resource actuator
// with them, printing all resulting resources at the end on stdout
func WithStdinResources(resourceActuator ResourceActuator) error {
	hypervisors, clusters, components, err := retrieveResourcesFromStdin()
	if err != nil {
		return err
	}
	components, err = resourceActuator(hypervisors, clusters, components)
	if err != nil {
		return err
	}
	res := ""
	if hypervisorsSpecs, err := hypervisors.Specs(); err == nil {
		res += hypervisorsSpecs
	}
	if clustersSpecs, err := clusters.Specs(); err == nil {
		res += clustersSpecs
	}
	if componentsSpecs, err := components.Specs(); err == nil {
		res += componentsSpecs
	}
	fmt.Print(res)
	return nil
}

// WithStdinResourcesSilent is similar to WithStdinResources, only
// that it won't print the resulting resources after calling to the
// actuator
func WithStdinResourcesSilent(resourceActuator ResourceActuator) error {
	hypervisors, clusters, components, err := retrieveResourcesFromStdin()
	if err != nil {
		return err
	}
	_, err = resourceActuator(hypervisors, clusters, components)
	return err
}

func retrieveResourcesFromStdin() (infra.HypervisorMap, cluster.Map, component.List, error) {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return infra.HypervisorMap{}, cluster.Map{}, component.List{}, err
	}
	hypervisors := RetrieveHypervisors(string(stdin))
	clusters := RetrieveClusters(string(stdin))
	components := RetrieveComponents(string(stdin))
	return hypervisors, clusters, components, nil
}

// RetrieveHypervisors returns an hypervisor map from the given manifests
func RetrieveHypervisors(manifests string) infra.HypervisorMap {
	klog.V(1).Info("retrieving hypervisors from manifests")
	hypervisors := infra.HypervisorMap{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := infrav1alpha1.AddToScheme(scheme); err != nil {
		return infra.HypervisorMap{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		hypervisor := infrav1alpha1.Hypervisor{}
		if _, _, err := serializer.Decode([]byte(document), nil, &hypervisor); err != nil || hypervisor.TypeMeta.Kind != "Hypervisor" {
			continue
		}
		internalHypervisor, err := infra.NewHypervisorFromv1alpha1(&hypervisor, nil)
		if err != nil {
			continue
		}
		hypervisors[internalHypervisor.Name] = internalHypervisor
	}
	return hypervisors
}

// RetrieveClusters returns a cluster list from the given manifests
func RetrieveClusters(manifests string) cluster.Map {
	klog.V(1).Info("retrieving clusters from manifests")
	clusters := cluster.Map{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return cluster.Map{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		clusterObj := clusterv1alpha1.Cluster{}
		if _, _, err := serializer.Decode([]byte(document), nil, &clusterObj); err != nil || clusterObj.TypeMeta.Kind != "Cluster" {
			continue
		}
		internalCluster, err := cluster.NewClusterFromv1alpha1(&clusterObj)
		if err != nil {
			continue
		}
		clusters[internalCluster.Name] = internalCluster
	}
	return clusters
}

// RetrieveComponents returns a component list from the given manifests
func RetrieveComponents(manifests string) component.List {
	klog.V(1).Info("retrieving components from manifests")
	components := component.List{}
	documents := yamlutils.SplitDocuments(manifests)
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return component.List{}
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{Yaml: true})
	for _, document := range documents {
		componentObj := clusterv1alpha1.Component{}
		if _, _, err := serializer.Decode([]byte(document), nil, &componentObj); err != nil || componentObj.TypeMeta.Kind != "Component" {
			continue
		}
		internalComponent, err := component.NewComponentFromv1alpha1(&componentObj)
		if err != nil {
			continue
		}
		components = append(components, internalComponent)
	}
	return components
}
