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
	"time"

	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
	"github.com/pkg/errors"
)

// Reconcile reconciles all clusters
func Reconcile(maxRetries int, retryWaitTime time.Duration) error {
	klog.V(1).Info("reading input manifests")
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	hypervisors := manifests.RetrieveHypervisors(string(stdin))
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	clusterReconciler := reconciler.NewClusterReconciler(hypervisors, clusters, components)

	var reconcileErrs reconciler.ReconcileErrors
	for i := 0; i < maxRetries; i++ {
		reconcileErrs = clusterReconciler.Reconcile()
		if reconcileErrs == nil {
			break
		}
		klog.V(2).Infof("failed to reconcile some resources: %v, retrying (%d/%d) after %s of wait time", reconcileErrs, i+1, maxRetries, retryWaitTime)
		time.Sleep(retryWaitTime)
	}

	clusterSpecs, err := clusterReconciler.Specs()
	if err != nil {
		return err
	}
	fmt.Print(clusterSpecs)

	if reconcileErrs != nil {
		return errors.Wrap(reconcileErrs, "failed to reconcile some resources")
	}

	return nil
}
