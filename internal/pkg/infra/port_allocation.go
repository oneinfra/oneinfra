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
	infrav1alpha1 "github.com/oneinfra/oneinfra/m/apis/infra/v1alpha1"
)

// HypervisorPortAllocation represents a port allocation in an hypervisor
type HypervisorPortAllocation struct {
	Cluster   string
	Component string
	Port      int
}

// HypervisorPortAllocationList represents a list of port allocations in an hypervisor
type HypervisorPortAllocationList []HypervisorPortAllocation

// NewHypervisorPortAllocationListFromv1alpha1 creates an hypervisor port allocation list
func NewHypervisorPortAllocationListFromv1alpha1(hypervisorPortAllocationList []infrav1alpha1.HypervisorPortAllocation) HypervisorPortAllocationList {
	res := HypervisorPortAllocationList{}
	for _, hypervisorPortAllocation := range hypervisorPortAllocationList {
		res = append(res, HypervisorPortAllocation{
			Cluster:   hypervisorPortAllocation.Cluster,
			Component: hypervisorPortAllocation.Component,
			Port:      hypervisorPortAllocation.Port,
		})
	}
	return res
}

// Export exports the hypervisor port allocation list to a versioned object
func (hypervisorPortAllocationList HypervisorPortAllocationList) Export() []infrav1alpha1.HypervisorPortAllocation {
	res := []infrav1alpha1.HypervisorPortAllocation{}
	for _, hypervisorPortAllocation := range hypervisorPortAllocationList {
		res = append(res, infrav1alpha1.HypervisorPortAllocation{
			Cluster:   hypervisorPortAllocation.Cluster,
			Component: hypervisorPortAllocation.Component,
			Port:      hypervisorPortAllocation.Port,
		})
	}
	return res
}
