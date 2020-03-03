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
	"io/ioutil"
	"os"

	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// Reconcile reconciles all clusters
func Reconcile() error {
	klog.V(1).Info("reading input manifests")
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	clusterReconciler := reconciler.NewClusterReconciler(hypervisors, clusters, components)
	if err := clusterReconciler.Reconcile(); err != nil {
		return err
	}
	clusterSpecs, err := clusterReconciler.Specs()
	if err != nil {
		return err
	}

	fmt.Print(clusterSpecs)

	return nil
}
