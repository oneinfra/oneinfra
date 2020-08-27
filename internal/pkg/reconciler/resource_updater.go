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

package reconciler

import (
	"context"

	"k8s.io/klog/v2"
	clientapi "sigs.k8s.io/controller-runtime/pkg/client"
)

// UpdateResources updates all resources known to this cluster
// reconciler if they are dirty
func UpdateResources(ctx context.Context, reconciler Reconciler, client clientapi.Client) error {
	if err := updateHypervisors(ctx, reconciler, client); err != nil {
		return err
	}
	if err := updateClusters(ctx, reconciler, client); err != nil {
		return err
	}
	if err := updateComponents(ctx, reconciler, client); err != nil {
		return err
	}
	return nil
}

func updateHypervisors(ctx context.Context, reconciler Reconciler, client clientapi.Client) error {
	for _, hypervisor := range reconciler.HypervisorMap() {
		isDirty, err := hypervisor.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if hypervisor %q is dirty", hypervisor.Name)
			continue
		}
		if isDirty {
			if err := client.Status().Update(ctx, hypervisor.Export()); err != nil {
				klog.Errorf("could not update hypervisor %q status: %v", hypervisor.Name, err)
				return err
			}
		}
	}
	return nil
}

func updateClusters(ctx context.Context, reconciler Reconciler, client clientapi.Client) error {
	for _, cluster := range reconciler.ClusterMap() {
		isDirty, err := cluster.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if cluster %q is dirty", cluster.Name)
			continue
		}
		if isDirty {
			if err := client.Status().Update(ctx, cluster.Export()); err != nil {
				klog.Errorf("could not update cluster %q status: %v", cluster.Name, err)
				return err
			}
		}
	}
	return nil
}

func updateComponents(ctx context.Context, reconciler Reconciler, client clientapi.Client) error {
	for _, component := range reconciler.ComponentList() {
		isDirty, err := component.IsDirty()
		if err != nil {
			klog.Errorf("could not determine if component %q is dirty", component.Name)
			continue
		}
		if isDirty {
			if err := client.Status().Update(ctx, component.Export()); err != nil {
				klog.Errorf("could not update component %q status: %v", component.Name, err)
				return err
			}
		}
	}
	return nil
}
