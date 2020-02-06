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

package localcluster

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	infrav1alpha1 "oneinfra.ereslibre.es/m/apis/infra/v1alpha1"
)

type Cluster struct {
	Name  string
	Nodes []*Node
}

func NewCluster(name string, size int) Cluster {
	cluster := Cluster{
		Name:  name,
		Nodes: []*Node{},
	}
	for i := 0; i < size; i++ {
		cluster.addNode(
			&Node{
				Name:    fmt.Sprintf("node-%d", i),
				Cluster: &cluster,
			},
		)
	}
	return cluster
}

func LoadCluster(name string) (Cluster, error) {
	cluster := Cluster{Name: name}
	nodes, err := ioutil.ReadDir(cluster.directory())
	if err != nil {
		return Cluster{}, err
	}
	return NewCluster(name, len(nodes)), nil
}

func (cluster *Cluster) addNode(node *Node) {
	cluster.Nodes = append(cluster.Nodes, node)
}

func (cluster *Cluster) Create() error {
	if err := cluster.createDirectory(); err != nil {
		return err
	}
	for _, node := range cluster.Nodes {
		if err := node.Create(); err != nil {
			return err
		}
	}
	return nil
}

func (cluster *Cluster) Wait() error {
	var wg sync.WaitGroup
	wg.Add(len(cluster.Nodes))
	for _, node := range cluster.Nodes {
		go func(node *Node) {
			for {
				if _, err := node.Version(context.Background()); err == nil {
					wg.Done()
					return
				}
			}
		}(node)
	}
	wg.Wait()
	return nil
}

func (cluster *Cluster) Destroy() error {
	for _, node := range cluster.Nodes {
		if err := node.Destroy(); err != nil {
			log.Printf("could not destroy node %q; continuing...\n", node.Name)
		}
	}
	return os.RemoveAll(cluster.directory())
}

func (cluster *Cluster) createDirectory() error {
	return os.MkdirAll(cluster.directory(), 0755)
}

func (cluster *Cluster) directory() string {
	return filepath.Join(os.TempDir(), "oneinfra-clusters", cluster.Name)
}

func (cluster *Cluster) Export() []infrav1alpha1.Hypervisor {
	hypervisors := []infrav1alpha1.Hypervisor{}
	for _, node := range cluster.Nodes {
		hypervisors = append(hypervisors, node.Export())
	}
	return hypervisors
}

func (cluster *Cluster) Specs() string {
	res := "---\n"
	scheme := runtime.NewScheme()
	infrav1alpha1.AddToScheme(scheme)
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, infrav1alpha1.GroupVersion)
	for _, node := range cluster.Nodes {
		nodeObject := node.Export()
		if encodedNode, err := runtime.Encode(encoder, &nodeObject); err == nil {
			res += fmt.Sprintf("%s---\n", string(encodedNode))
		}
	}
	return res
}
