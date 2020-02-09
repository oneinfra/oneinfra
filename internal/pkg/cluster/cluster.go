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

package cluster

import (
	"fmt"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
	"oneinfra.ereslibre.es/m/internal/pkg/node"
)

type Cluster struct {
	Name  string
	nodes []*node.Node
}

type ClusterMap map[string]*Cluster
type ClusterList []*Cluster

func NewCluster(clusterName string) *Cluster {
	return &Cluster{
		Name: clusterName,
	}
}

func ClusterWithNodesFromv1alpha1(cluster *clusterv1alpha1.Cluster, nodes node.NodeList) (*Cluster, error) {
	res := Cluster{
		Name:  cluster.ObjectMeta.Name,
		nodes: []*node.Node{},
	}
	for _, node := range nodes {
		if node.ClusterName == res.Name {
			res.nodes = append(res.nodes, node)
		}
	}
	return &res, nil
}

func (cluster *Cluster) Reconcile() error {
	for _, node := range cluster.nodes {
		if err := node.Reconcile(); err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) Export() *clusterv1alpha1.Cluster {
	return &clusterv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster.Name,
		},
		Spec: clusterv1alpha1.ClusterSpec{},
	}
}

func (cluster *Cluster) Specs() (string, error) {
	scheme := runtime.NewScheme()
	if err := clusterv1alpha1.AddToScheme(scheme); err != nil {
		return "", err
	}
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, clusterv1alpha1.GroupVersion)
	clusterObject := cluster.Export()
	if encodedCluster, err := runtime.Encode(encoder, clusterObject); err == nil {
		return string(encodedCluster), nil
	}
	return "", errors.Errorf("could not encode cluster %q", cluster.Name)
}

func (clusterList ClusterList) Specs() (string, error) {
	res := ""
	for _, cluster := range clusterList {
		clusterSpec, err := cluster.Specs()
		if err != nil {
			continue
		}
		res += fmt.Sprintf("---\n%s", clusterSpec)
	}
	return res, nil
}

func (clusterList ClusterList) Reconcile() error {
	for _, cluster := range clusterList {
		if err := cluster.Reconcile(); err != nil {
			return err
		}
	}
	return nil
}
