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

package cluster

import (
	"time"

	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	clusterreconciler "github.com/oneinfra/oneinfra/internal/pkg/cluster/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentreconciler "github.com/oneinfra/oneinfra/internal/pkg/component/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
	"github.com/pkg/errors"
)

// Reconcile reconciles all clusters
func Reconcile(maxRetries int, retryWaitTime time.Duration) error {
	return manifests.WithStdinResources(
		func(hypervisors infra.HypervisorMap, clusters cluster.Map, components component.List) (component.List, error) {
			componentReconciler := componentreconciler.NewComponentReconciler(hypervisors, clusters, components)
			clusterReconciler := clusterreconciler.NewClusterReconciler(hypervisors, clusters, components)
			var componentReconcileErrs, clusterReconcileErrs reconciler.ReconcileErrors
			for i := 0; i < maxRetries; i++ {
				componentReconcileErrs = componentReconciler.Reconcile()
				clusterReconcileErrs = clusterReconciler.Reconcile(clusterreconciler.OptionalReconcile{
					ReconcileNodeJoinRequests: true,
				})
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
			if componentReconcileErrs != nil || clusterReconcileErrs != nil {
				return component.List{}, errors.New("failed to reconcile some resources")
			}
			klog.Info("reconciliation finished successfully")
			return components, nil
		},
	)
}
