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

package localhypervisorset

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"

	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
)

// HypervisorSet represents a set of local hypervisors
type HypervisorSet struct {
	Name        string
	NodeImage   string
	Remote      bool
	NetworkName string
	Hypervisors []*Hypervisor
}

// NewHypervisorSet creates a new set of local hypervisors
func NewHypervisorSet(name, kubernetesVersion string, privateHypervisorSetSize, publicHypervisorSetSize int, remote bool, networkName string) *HypervisorSet {
	networkName, err := NetworkName(networkName)
	if err != nil {
		klog.Fatal(err)
	}
	set := HypervisorSet{
		Name:        name,
		NodeImage:   fmt.Sprintf("oneinfra/hypervisor:%s", kubernetesVersion),
		Remote:      remote,
		NetworkName: networkName,
		Hypervisors: []*Hypervisor{},
	}
	for i := 0; i < privateHypervisorSetSize; i++ {
		set.addHypervisor(
			&Hypervisor{
				Name:                 fmt.Sprintf("private-hypervisor-%d", i),
				Public:               false,
				HypervisorSet:        &set,
				ExposedPortRangeLow:  30000,
				ExposedPortRangeHigh: 60000,
			},
		)
	}
	for i := 0; i < publicHypervisorSetSize; i++ {
		set.addHypervisor(
			&Hypervisor{
				Name:                 fmt.Sprintf("public-hypervisor-%d", i),
				Public:               true,
				HypervisorSet:        &set,
				ExposedPortRangeLow:  30000,
				ExposedPortRangeHigh: 60000,
			},
		)
	}
	return &set
}

// LoadHypervisorSet loads an hypervisor set with name from disk
func LoadHypervisorSet(name string) (*HypervisorSet, error) {
	hypervisorSet := HypervisorSet{Name: name}
	privateHypervisors, err := filepath.Glob(filepath.Join(hypervisorSet.directory(), "private-hypervisor-*"))
	if err != nil {
		return nil, err
	}
	publicHypervisors, err := filepath.Glob(filepath.Join(hypervisorSet.directory(), "public-hypervisor-*"))
	if err != nil {
		return nil, err
	}
	return NewHypervisorSet(name, "", len(privateHypervisors), len(publicHypervisors), false, ""), nil
}

func (hypervisorSet *HypervisorSet) addHypervisor(hypervisor *Hypervisor) {
	hypervisorSet.Hypervisors = append(hypervisorSet.Hypervisors, hypervisor)
}

// Create creates the local hypervisor set
func (hypervisorSet *HypervisorSet) Create() error {
	if err := hypervisorSet.createDirectory(); err != nil {
		return err
	}
	for _, hypervisor := range hypervisorSet.Hypervisors {
		if err := hypervisor.Create(); err != nil {
			return err
		}
	}
	return nil
}

// Wait waits for the local hypervisor set to be created
func (hypervisorSet *HypervisorSet) Wait() error {
	var wg sync.WaitGroup
	wg.Add(len(hypervisorSet.Hypervisors))
	for _, hypervisor := range hypervisorSet.Hypervisors {
		go func(hypervisor *Hypervisor) {
			hypervisor.Wait()
			wg.Done()
		}(hypervisor)
	}
	wg.Wait()
	return nil
}

// StartRemoteCRIEndpoints initializes the CRI endpoint on all hypervisors
func (hypervisorSet *HypervisorSet) StartRemoteCRIEndpoints() error {
	for _, hypervisor := range hypervisorSet.Hypervisors {
		if err := hypervisor.StartRemoteCRIEndpoint(); err != nil {
			return err
		}
	}
	return nil
}

// Destroy destroys the local hypervisor set
func (hypervisorSet *HypervisorSet) Destroy() error {
	for _, hypervisor := range hypervisorSet.Hypervisors {
		if err := hypervisor.Destroy(); err != nil {
			log.Printf("could not destroy hypervisor %q; continuing...\n", hypervisor.Name)
		}
	}
	return os.RemoveAll(hypervisorSet.directory())
}

func (hypervisorSet *HypervisorSet) createDirectory() error {
	return os.MkdirAll(hypervisorSet.directory(), 0755)
}

func (hypervisorSet *HypervisorSet) directory() string {
	return filepath.Join(os.TempDir(), "oneinfra-hypervisor-sets", hypervisorSet.Name)
}

// Export exports the local hypervisor set to a list of versioned hypervisors
func (hypervisorSet *HypervisorSet) Export() []*infrav1alpha1.Hypervisor {
	hypervisors := []*infrav1alpha1.Hypervisor{}
	for _, hypervisor := range hypervisorSet.Hypervisors {
		hypervisors = append(hypervisors, hypervisor.Export())
	}
	return hypervisors
}

// Specs returns the versioned specs for the local hypervisor set
func (hypervisorSet *HypervisorSet) Specs() string {
	res := ""
	scheme := runtime.NewScheme()
	infrav1alpha1.AddToScheme(scheme)
	info, _ := runtime.SerializerInfoForMediaType(serializer.NewCodecFactory(scheme).SupportedMediaTypes(), runtime.ContentTypeYAML)
	encoder := serializer.NewCodecFactory(scheme).EncoderForVersion(info.Serializer, infrav1alpha1.GroupVersion)
	for _, hypervisor := range hypervisorSet.Hypervisors {
		hypervisorObject := hypervisor.Export()
		if encodedHypervisor, err := runtime.Encode(encoder, hypervisorObject); err == nil {
			res += fmt.Sprintf("---\n%s\n", string(encodedHypervisor))
		}
	}
	return res
}
