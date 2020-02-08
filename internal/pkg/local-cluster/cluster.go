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

type HypervisorCluster struct {
	Name        string
	Hypervisors []*Hypervisor
}

func NewHypervisorCluster(name string, size int) *HypervisorCluster {
	cluster := HypervisorCluster{
		Name:        name,
		Hypervisors: []*Hypervisor{},
	}
	for i := 0; i < size; i++ {
		cluster.addHypervisor(
			&Hypervisor{
				Name:              fmt.Sprintf("hypervisor-%d", i),
				HypervisorCluster: &cluster,
			},
		)
	}
	return &cluster
}

func LoadCluster(name string) (*HypervisorCluster, error) {
	hypervisorCluster := HypervisorCluster{Name: name}
	hypervisors, err := ioutil.ReadDir(hypervisorCluster.directory())
	if err != nil {
		return nil, err
	}
	return NewHypervisorCluster(name, len(hypervisors)), nil
}

func (hypervisorCluster *HypervisorCluster) addHypervisor(hypervisor *Hypervisor) {
	hypervisorCluster.Hypervisors = append(hypervisorCluster.Hypervisors, hypervisor)
}

func (hypervisorCluster *HypervisorCluster) Create() error {
	if err := hypervisorCluster.createDirectory(); err != nil {
		return err
	}
	for _, hypervisor := range hypervisorCluster.Hypervisors {
		if err := hypervisor.Create(); err != nil {
			return err
		}
	}
	return nil
}

func (hypervisorCluster *HypervisorCluster) Wait() error {
	var wg sync.WaitGroup
	wg.Add(len(hypervisorCluster.Hypervisors))
	for _, hypervisor := range hypervisorCluster.Hypervisors {
		go func(hypervisor *Hypervisor) {
			hypervisor.Wait()
			wg.Done()
		}(hypervisor)
	}
	wg.Wait()
	return nil
}

func (hypervisorCluster *HypervisorCluster) Destroy() error {
	for _, hypervisor := range hypervisorCluster.Hypervisors {
		if err := hypervisor.Destroy(); err != nil {
			log.Printf("could not destroy hypervisor %q; continuing...\n", hypervisor.Name)
		}
	}
	return os.RemoveAll(hypervisorCluster.directory())
}

func (hypervisorCluster *HypervisorCluster) createDirectory() error {
	return os.MkdirAll(hypervisorCluster.directory(), 0755)
}

func (hypervisorCluster *HypervisorCluster) directory() string {
	return filepath.Join(os.TempDir(), "oneinfra-clusters", hypervisorCluster.Name)
}

func (hypervisorCluster *HypervisorCluster) Export() []infrav1alpha1.Hypervisor {
	hypervisors := []infrav1alpha1.Hypervisor{}
	for _, hypervisor := range hypervisorCluster.Hypervisors {
		hypervisors = append(hypervisors, hypervisor.Export())
	}
	return hypervisors
}

func (hypervisorCluster *HypervisorCluster) Specs() string {
	res := "---\n"
	scheme := runtime.NewScheme()
	infrav1alpha1.AddToScheme(scheme)
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, infrav1alpha1.GroupVersion)
	for _, hypervisor := range hypervisorCluster.Hypervisors {
		hypervisorObject := hypervisor.Export()
		if encodedHypervisor, err := runtime.Encode(encoder, &hypervisorObject); err == nil {
			res += fmt.Sprintf("%s---\n", string(encodedHypervisor))
		}
	}
	return res
}
