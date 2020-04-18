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

	clusterreconciler "github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	componentreconciler "github.com/oneinfra/oneinfra/internal/pkg/component/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
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

	componentReconciler := componentreconciler.NewComponentReconciler(hypervisors, clusters, components)
	clusterReconciler := clusterreconciler.NewClusterReconciler(hypervisors, clusters, components)
	var componentReconcileErrs, clusterReconcileErrs reconciler.ReconcileErrors
	for i := 0; i < maxRetries; i++ {
		componentReconcileErrs = componentReconciler.Reconcile()
		clusterReconcileErrs = clusterReconciler.Reconcile()
		if componentReconcileErrs == nil && clusterReconcileErrs == nil {
			break
		}
		time.Sleep(retryWaitTime)
	}

	if componentReconcileErrs != nil {
		klog.V(2).Infof("failed to reconcile some components: %v", componentReconcileErrs)
	}

	if clusterReconcileErrs != nil {
		klog.V(2).Infof("failed to reconcile some clusters: %v", clusterReconcileErrs)
	}

	clusterSpecs, err := clusterReconciler.Specs()
	if err != nil {
		return err
	}
	fmt.Print(clusterSpecs)

	if componentReconcileErrs != nil || clusterReconcileErrs != nil {
		return errors.New("failed to reconcile some resources")
	}

	klog.Info("reconciliation finished successfully")

	return nil
}
