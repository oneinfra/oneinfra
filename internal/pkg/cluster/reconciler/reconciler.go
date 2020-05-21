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
	"net"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog"

	clusterapi "github.com/oneinfra/oneinfra/internal/pkg/cluster"
	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/component/components"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/reconciler"
	"github.com/oneinfra/oneinfra/internal/pkg/utils"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// OptionalReconcile represents what optional reconciliations should
// or should not take place
type OptionalReconcile struct {
	ReconcileNodeJoinRequests bool
}

// ClusterReconciler represents a cluster reconciler
type ClusterReconciler struct {
	hypervisorMap infra.HypervisorMap
	clusterMap    clusterapi.Map
	componentList componentapi.List
}

// NewClusterReconciler creates a cluster reconciler with the provided hypervisors, clusters and components
func NewClusterReconciler(hypervisorMap infra.HypervisorMap, clusterMap clusterapi.Map, componentList componentapi.List) *ClusterReconciler {
	return &ClusterReconciler{
		hypervisorMap: hypervisorMap,
		clusterMap:    clusterMap,
		componentList: componentList,
	}
}

// IsComponentScheduled returns whether this component is scheduled to
// an existing hypervisor
func (clusterReconciler *ClusterReconciler) IsComponentScheduled(component *componentapi.Component) bool {
	if component.HypervisorName == "" {
		return false
	}
	_, exists := clusterReconciler.hypervisorMap[component.HypervisorName]
	return exists
}

// IsClusterFullyScheduled returns whether all components assigned to
// this cluster are scheduled
func (clusterReconciler *ClusterReconciler) IsClusterFullyScheduled(clusterNamespace, clusterName string) bool {
	hasComponents := false
	for _, component := range clusterReconciler.componentList.WithCluster(clusterNamespace, clusterName) {
		hasComponents = true
		if !clusterReconciler.IsComponentScheduled(component) {
			return false
		}
	}
	return hasComponents
}

// Reconcile reconciles the provided clusters
func (clusterReconciler *ClusterReconciler) Reconcile(optionalReconcile OptionalReconcile, clustersToReconcile ...*clusterapi.Cluster) reconciler.ReconcileErrors {
	if len(clustersToReconcile) == 0 {
		clustersToReconcile = []*clusterapi.Cluster{}
		for _, cluster := range clusterReconciler.clusterMap {
			clustersToReconcile = append(
				clustersToReconcile,
				cluster,
			)
		}
	}
	reconcileErrors := reconciler.ReconcileErrors{}
	for _, cluster := range clustersToReconcile {
		klog.V(1).Infof("reconciling cluster %q", cluster.Name)
		if !clusterReconciler.IsClusterFullyScheduled(cluster.Namespace, cluster.Name) {
			klog.Infof("cluster %q is not fully scheduled; skipping", cluster.Name)
			reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("cluster is not fully scheduled"))
			continue
		}
		clusterReconciler.reconcileAPIServerEndpoint(cluster, &reconcileErrors)
		if !reconcileErrors.IsClusterErrorFree(cluster.Namespace, cluster.Name) {
			continue
		}
		cluster.Conditions.SetCondition(
			clusterapi.ReconcileStarted,
			conditions.ConditionTrue,
		)
		if cluster.VPN.Enabled {
			clusterReconciler.reconcileMinimalVPNPeers(cluster, &reconcileErrors)
			clusterReconciler.reconcileVPNServerEndpoint(cluster, &reconcileErrors)
		}
		clusterReconciler.reconcileCustomResourceDefinitions(cluster, &reconcileErrors)
		clusterReconciler.reconcileNamespaces(cluster, &reconcileErrors)
		clusterReconciler.reconcilePermissions(cluster, &reconcileErrors)
		clusterReconciler.reconcileJoinTokens(cluster, &reconcileErrors)
		clusterReconciler.reconcileStorageEndpoints(cluster, &reconcileErrors)
		clusterReconciler.reconcileKubeProxy(cluster, &reconcileErrors)
		clusterReconciler.reconcileCoreDNS(cluster, &reconcileErrors)
		if optionalReconcile.ReconcileNodeJoinRequests {
			clusterReconciler.reconcileNodeJoinRequests(cluster, &reconcileErrors)
		}
		clusterReconciler.reconcileJoinPublicKeyConfigMap(cluster, &reconcileErrors)
		if reconcileErrors.IsClusterErrorFree(cluster.Namespace, cluster.Name) {
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

func (clusterReconciler *ClusterReconciler) reconcileAPIServerEndpoint(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	controlPlaneIngressList := clusterReconciler.componentList.WithRole(componentapi.ControlPlaneIngressRole)
	if len(controlPlaneIngressList) == 0 {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("could not find any control plane ingress component"))
		return
	}
	if len(controlPlaneIngressList) > 1 {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("more than one control plane ingress component was found"))
		return
	}
	controlPlaneIngress := controlPlaneIngressList[0]
	hypervisor, exists := clusterReconciler.hypervisorMap[controlPlaneIngress.HypervisorName]
	if !exists {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("found one control plane ingress component, but it is assigned to an unknown hypervisor"))
		return
	}
	apiserverHostPort, exists := controlPlaneIngress.AllocatedHostPorts[components.APIServerHostPortName]
	if !exists {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Errorf("found one control plane ingress component, but it does not have a named host port %q yet", components.APIServerHostPortName))
		return
	}
	url := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort))}
	cluster.APIServerEndpoint = url.String()
}

func (clusterReconciler *ClusterReconciler) reconcileVPNServerEndpoint(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	controlPlaneIngressList := clusterReconciler.componentList.WithRole(componentapi.ControlPlaneIngressRole)
	if len(controlPlaneIngressList) == 0 {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("could not find any control plane ingress component"))
		return
	}
	if len(controlPlaneIngressList) > 1 {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("more than one control plane ingress component was found"))
		return
	}
	controlPlaneIngress := controlPlaneIngressList[0]
	hypervisor, exists := clusterReconciler.hypervisorMap[controlPlaneIngress.HypervisorName]
	if !exists {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("found one control plane ingress component, but it is assigned to an unknown hypervisor"))
		return
	}
	wireguardHostPort, exists := controlPlaneIngress.AllocatedHostPorts[components.WireguardHostPortName]
	if !exists {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Errorf("found one control plane ingress component, but it does not have a named host port %q yet", components.WireguardHostPortName))
		return
	}
	cluster.VPNServerEndpoint = net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(wireguardHostPort))
}

func (clusterReconciler *ClusterReconciler) reconcileMinimalVPNPeers(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileMinimalVPNPeers(); err != nil {
		klog.Errorf("failed to reconcile minimal VPN peers for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile minimal VPN peers"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileCustomResourceDefinitions(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileCustomResourceDefinitions(); err != nil {
		klog.Errorf("failed to reconcile custom resource definitions for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile custom resource definitions"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileNamespaces(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileNamespaces(); err != nil {
		klog.Errorf("failed to reconcile namespaces for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile namespaces"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcilePermissions(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcilePermissions(); err != nil {
		klog.Errorf("failed to reconcile permissions for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile permissions"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileJoinTokens(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileJoinTokens(); err != nil {
		klog.Errorf("failed to reconcile join tokens for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile join tokens"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileStorageEndpoints(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	controlPlaneList := clusterReconciler.componentList.WithRole(componentapi.ControlPlaneRole)
	if cluster.ControlPlaneReplicas != len(controlPlaneList) {
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.New("the number of control plane components does not match the number of desired replicas"))
		return
	}
	storagePeerEndpoints := map[string]string{}
	storageClientEndpoints := map[string]string{}
	for _, controlPlane := range controlPlaneList {
		if controlPlane.DeletionTimestamp != nil {
			continue
		}
		hypervisor, exists := clusterReconciler.hypervisorMap[controlPlane.HypervisorName]
		if !exists {
			reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Errorf("could not find hypervisor for component %s/%s", controlPlane.Namespace, controlPlane.Name))
			return
		}
		etcdPeerHostPort, exists := controlPlane.AllocatedHostPorts[components.EtcdPeerHostPortName]
		if !exists {
			reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Errorf("could not find etcd peer host port for component %s/%s", controlPlane.Namespace, controlPlane.Name))
			return
		}
		storagePeerURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
		storagePeerEndpoints[controlPlane.Name] = storagePeerURL.String()
		etcdClientHostPort, exists := controlPlane.AllocatedHostPorts[components.EtcdClientHostPortName]
		if !exists {
			reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Errorf("could not find etcd client host port for component %s/%s", controlPlane.Namespace, controlPlane.Name))
			return
		}
		storageClientURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdClientHostPort))}
		storageClientEndpoints[controlPlane.Name] = storageClientURL.String()
	}
	cluster.StoragePeerEndpoints = storagePeerEndpoints
	cluster.StorageClientEndpoints = storageClientEndpoints
}

func (clusterReconciler *ClusterReconciler) reconcileKubeProxy(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileKubeProxy(); err != nil {
		klog.Errorf("failed to reconcile kube-proxy for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile kube-proxy"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileCoreDNS(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileCoreDNS(); err != nil {
		klog.Errorf("failed to reconcile CoreDNS for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile CoreDNS"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileNodeJoinRequests(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileNodeJoinRequests(); err != nil {
		klog.Errorf("failed to reconcile node join requests for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile node join requests"))
	}
}

func (clusterReconciler *ClusterReconciler) reconcileJoinPublicKeyConfigMap(cluster *clusterapi.Cluster, reconcileErrors *reconciler.ReconcileErrors) {
	if err := cluster.ReconcileJoinPublicKeyConfigMap(); err != nil {
		klog.Errorf("failed to reconcile join public key ConfigMap for cluster %q: %v", cluster.Name, err)
		reconcileErrors.AddClusterError(cluster.Namespace, cluster.Name, errors.Wrap(err, "failed to reconcile join public key ConfigMap"))
	}
}

// ReconcileDeletion reconciles the deletion of the provided clusters
func (clusterReconciler *ClusterReconciler) ReconcileDeletion(clustersToDelete ...*clusterapi.Cluster) reconciler.ReconcileErrors {
	reconcileErrors := reconciler.ReconcileErrors{}
	for _, clusterToDelete := range clustersToDelete {
		if len(clusterReconciler.ComponentList().WithCluster(clusterToDelete.Namespace, clusterToDelete.Name)) > 0 {
			reconcileErrors.AddClusterError(clusterToDelete.Namespace, clusterToDelete.Name, errors.New("not all components are deleted yet"))
			continue
		}
		clusterToDelete.Finalizers = utils.RemoveElementsFromList(
			clusterToDelete.Finalizers,
			constants.OneInfraCleanupFinalizer,
		)
	}
	if len(reconcileErrors) == 0 {
		return nil
	}
	return reconcileErrors
}

// Specs returns the versioned specs for all resources
func (clusterReconciler *ClusterReconciler) Specs() (string, error) {
	res := ""
	hypervisors, err := clusterReconciler.hypervisorMap.Specs()
	if err != nil {
		return "", nil
	}
	res += hypervisors
	clusters, err := clusterReconciler.clusterMap.Specs()
	if err != nil {
		return "", nil
	}
	res += clusters
	components, err := clusterReconciler.componentList.Specs()
	if err != nil {
		return "", nil
	}
	res += components
	return res, nil
}
