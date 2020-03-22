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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
)

// ReconcileNodeJoinRequests reconciles this cluster node join requests
func (cluster *Cluster) ReconcileNodeJoinRequests() error {
	scheme := runtime.NewScheme()
	if err := nodev1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	client, err := cluster.RESTClient(&nodev1alpha1.GroupVersion, scheme)
	if err != nil {
		return err
	}
	nodeJoinRequestList := nodev1alpha1.NodeJoinRequestList{}
	err = client.
		Get().
		Namespace(constants.OneInfraNamespace).
		Resource("nodejoinrequests").
		Do().
		Into(&nodeJoinRequestList)
	if err != nil {
		return err
	}
	for _, versionedNodeJoinRequest := range nodeJoinRequestList.Items {
		nodeJoinRequest, err := nodejoinrequests.NewNodeJoinRequestFromv1alpha1(&versionedNodeJoinRequest, cluster.JoinKey)
		if err != nil {
			klog.Errorf("cannot parse node join request %q public key: %v", versionedNodeJoinRequest.ObjectMeta.Name, err)
			continue
		}
		if nodeJoinRequest.HasCondition(nodejoinrequests.Issued) {
			continue
		}
		nodeJoinRequest.KubernetesVersion = cluster.KubernetesVersion
		vpnPeer, err := cluster.GenerateVPNPeer(fmt.Sprintf("worker-%s", nodeJoinRequest.Name))
		if err != nil {
			klog.Errorf("cannot request a VPN peer for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		vpnAddress, err := nodeJoinRequest.Encrypt(vpnPeer.Address)
		if err != nil {
			klog.Errorf("cannot encrypt VPN peer address for cluster %q", cluster.Name)
			continue
		}
		nodeJoinRequest.VPNAddress = vpnAddress
		ingressVPNPeerRaw, exists := cluster.VPNPeers[constants.OneInfraControlPlaneIngressVPNPeerName]
		if !exists {
			klog.Errorf("cannot find ingress VPN peer name for cluster %q", cluster.Name)
			continue
		}
		ingressVPNPeer, err := nodeJoinRequest.Encrypt(ingressVPNPeerRaw.Address)
		if err != nil {
			klog.Errorf("cannot encrypt ingress VPN peer address for cluster %q", cluster.Name)
			continue
		}
		nodeJoinRequest.VPNPeer = ingressVPNPeer
		if err := cluster.fillNodeJoinRequestKubeConfig(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubeconfig for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestKubeletConfig(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubelet config for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		nodeJoinRequest.Conditions = append(nodeJoinRequest.Conditions, nodejoinrequests.Issued)
		versionedNodeJoinRequest, err := nodeJoinRequest.Export()
		if err != nil {
			klog.Errorf("could not convert the internal node join request to a versioned node join request: %v", err)
		}
		err = client.
			Put().
			Namespace(constants.OneInfraNamespace).
			Resource("nodejoinrequests").
			Name(nodeJoinRequest.Name).
			SubResource("status").
			Body(versionedNodeJoinRequest).
			Do().
			Error()
		if err != nil {
			klog.Errorf("cannot update node join request status %q: %v", nodeJoinRequest.Name, err)
		}
	}
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestKubeConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	kubeConfig, err := cluster.KubeConfigWithEndpoint(nodeJoinRequest.APIServerEndpoint, fmt.Sprintf("system:node:%s", nodeJoinRequest.Name), []string{"system:nodes"})
	if err != nil {
		return err
	}
	kubeConfig, err = nodeJoinRequest.Encrypt(kubeConfig)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubeConfig = kubeConfig
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestKubeletConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	kubeletConfig, err := cluster.KubeletConfig()
	if err != nil {
		return err
	}
	kubeletConfig, err = nodeJoinRequest.Encrypt(kubeletConfig)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubeletConfig = kubeletConfig
	return nil
}
