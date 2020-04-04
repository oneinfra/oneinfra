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

package reconciler

import (
	"k8s.io/klog"

	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/component"
	componentreconciler "github.com/oneinfra/oneinfra/internal/pkg/component/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/pkg/errors"
)

// ClusterReconciler represents a cluster reconciler
type ClusterReconciler struct {
	HypervisorMap infra.HypervisorMap
	ClusterMap    clusterapi.Map
	ComponentList component.List
}

// NewClusterReconciler creates a cluster reconciler with the provided hypervisors, clusters and components
func NewClusterReconciler(hypervisorMap infra.HypervisorMap, clusterMap clusterapi.Map, componentList component.List) *ClusterReconciler {
	return &ClusterReconciler{
		HypervisorMap: hypervisorMap,
		ClusterMap:    clusterMap,
		ComponentList: componentList,
	}
}

// IsComponentScheduled returns whether this component is scheduled to
// an existing hypervisor
func (clusterReconciler *ClusterReconciler) IsComponentScheduled(component *component.Component) bool {
	if component.HypervisorName == "" {
		return false
	}
	_, exists := clusterReconciler.HypervisorMap[component.HypervisorName]
	return exists
}

// IsClusterFullyScheduled returns whether all components assigned to
// this cluster are scheduled
func (clusterReconciler *ClusterReconciler) IsClusterFullyScheduled(clusterName string) bool {
	hasComponents := false
	for _, component := range clusterReconciler.ComponentList.WithCluster(clusterName) {
		hasComponents = true
		if !clusterReconciler.IsComponentScheduled(component) {
			return false
		}
	}
	return hasComponents
}

// Reconcile reconciles all components known to this cluster
// reconciler. If nothing failed, it will return nil, otherwise a
// ReconcileErrors object will be returned specifying what
// reconciliations failed
func (clusterReconciler *ClusterReconciler) Reconcile() ReconcileErrors {
	klog.V(1).Info("starting reconciliation process")
	reconcileErrors := ReconcileErrors{}
	for clusterName, cluster := range clusterReconciler.ClusterMap {
		if !clusterReconciler.IsClusterFullyScheduled(clusterName) {
			klog.Infof("cluster %q is not fully scheduled; skipping", clusterName)
			continue
		}
		cluster.Conditions.SetCondition(
			clusterapi.ReconcileStarted,
			conditions.ConditionTrue,
		)

		clusterReconciler.reconcileMinimalVPNPeers(cluster, &reconcileErrors)
		clusterReconciler.reconcileControlPlaneComponents(clusterName, &reconcileErrors)
		clusterReconciler.reconcileControlPlaneIngressComponents(clusterName, &reconcileErrors)
		clusterReconciler.reconcileCustomResourceDefinitions(cluster, &reconcileErrors)
		clusterReconciler.reconcileNamespaces(cluster, &reconcileErrors)
		clusterReconciler.reconcilePermissions(cluster, &reconcileErrors)
		clusterReconciler.reconcileJoinTokens(cluster, &reconcileErrors)
		clusterReconciler.reconcileNodeJoinRequests(cluster, &reconcileErrors)

		if reconcileErrors.IsClusterErrorFree(clusterName) {
			cluster.Conditions.SetCondition(
				clusterapi.ReconcileSucceeded,
				conditions.ConditionTrue,
			)
		} else {
			cluster.Conditions.SetCondition(
				clusterapi.ReconcileSucceeded,
				conditions.ConditionFalse,
			)
		}
	}
	if len(reconcileErrors) == 0 {
		return nil
	}
	return reconcileErrors
}

func (clusterReconciler *ClusterReconciler) reconcileMinimalVPNPeers(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcileMinimalVPNPeers(); err != nil {
		klog.Errorf("failed to reconcile minimal VPN peers for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile minimal VPN peers"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileControlPlaneComponents(clusterName string, reconcileErrors *ReconcileErrors) {
	for _, componentObj := range clusterReconciler.ComponentList.WithCluster(clusterName).WithRole(component.ControlPlaneRole) {
		err := componentreconciler.Reconcile(
			&ClusterReconcilerInquirer{
				component:         componentObj,
				clusterReconciler: clusterReconciler,
			},
		)
		if err != nil {
			klog.Errorf("failed to reconcile component %q: %v", componentObj.Name, err)
			reconcileErrors.addComponentError(componentObj.ClusterName, componentObj.Name, err)
		}
	}
}

func (clusterReconciler *ClusterReconciler) reconcileControlPlaneIngressComponents(clusterName string, reconcileErrors *ReconcileErrors) {
	for _, componentObj := range clusterReconciler.ComponentList.WithCluster(clusterName).WithRole(component.ControlPlaneIngressRole) {
		err := componentreconciler.Reconcile(
			&ClusterReconcilerInquirer{
				component:         componentObj,
				clusterReconciler: clusterReconciler,
			},
		)
		if err != nil {
			klog.Errorf("failed to reconcile component %q: %v", componentObj.Name, err)
			reconcileErrors.addComponentError(componentObj.ClusterName, componentObj.Name, err)
		}
	}
}

func (clusterReconciler *ClusterReconciler) reconcileCustomResourceDefinitions(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcileCustomResourceDefinitions(); err != nil {
		klog.Errorf("failed to reconcile custom resource definitions for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile custom resource definitions"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileNamespaces(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcileNamespaces(); err != nil {
		klog.Errorf("failed to reconcile namespaces for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile namespaces"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcilePermissions(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcilePermissions(); err != nil {
		klog.Errorf("failed to reconcile permissions for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile permissions"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileJoinTokens(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcileJoinTokens(); err != nil {
		klog.Errorf("failed to reconcile join tokens for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile join tokens"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileNodeJoinRequests(cluster *clusterapi.Cluster, reconcileErrors *ReconcileErrors) {
	if err := cluster.ReconcileNodeJoinRequests(); err != nil {
		klog.Errorf("failed to reconcile node join requests for cluster %q: %v", cluster.Name, err)
		reconcileErrors.addClusterError(cluster.Name, errors.Wrap(err, "failed to reconcile node join requests"))
	}
}

// Specs returns the versioned specs for all resources
func (clusterReconciler *ClusterReconciler) Specs() (string, error) {
	res := ""
	hypervisors, err := clusterReconciler.HypervisorMap.Specs()
	if err != nil {
		return "", nil
	}
	res += hypervisors
	clusters, err := clusterReconciler.ClusterMap.Specs()
	if err != nil {
		return "", nil
	}
	res += clusters
	components, err := clusterReconciler.ComponentList.Specs()
	if err != nil {
		return "", nil
	}
	res += components
	return res, nil
}
