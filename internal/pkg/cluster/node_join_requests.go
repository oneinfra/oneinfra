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

package cluster

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/conditions"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
	"github.com/oneinfra/oneinfra/pkg/constants"
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
		Resource("nodejoinrequests").
		Do(context.TODO()).
		Into(&nodeJoinRequestList)
	if err != nil {
		return err
	}
	for _, versionedNodeJoinRequest := range nodeJoinRequestList.Items {
		nodeJoinRequest, err := nodejoinrequests.NewNodeJoinRequestFromv1alpha1(&versionedNodeJoinRequest, cluster.JoinKey)
		if err != nil {
			klog.Errorf("cannot parse node join request %q public key: %v", versionedNodeJoinRequest.Name, err)
			continue
		}
		if nodeJoinRequest.Conditions.IsCondition(nodejoinrequests.Issued, conditions.ConditionTrue) {
			continue
		}
		if err := cluster.fillNodeJoinRequestKubernetesVersion(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill Kubernetes version for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestVPNAddressAndPeers(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill VPN address and peers for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestKubeConfig(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubeconfig for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestKubeletConfig(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubelet config for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestKubeletServerCertificate(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubelet server certificate for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		if err := cluster.fillNodeJoinRequestKubeletClientCACertificate(nodeJoinRequest); err != nil {
			klog.Errorf("cannot fill kubelet client CA certificate for node join request %q: %v", nodeJoinRequest.Name, err)
			continue
		}
		nodeJoinRequest.Conditions.SetCondition(nodejoinrequests.Issued, conditions.ConditionTrue)
		versionedNodeJoinRequest, err := nodeJoinRequest.Export()
		if err != nil {
			klog.Errorf("could not convert the internal node join request to a versioned node join request: %v", err)
		}
		err = client.
			Put().
			Resource("nodejoinrequests").
			Name(nodeJoinRequest.Name).
			SubResource("status").
			Body(versionedNodeJoinRequest).
			Do(context.TODO()).
			Error()
		if err != nil {
			klog.Errorf("cannot update node join request status %q: %v", nodeJoinRequest.Name, err)
		}
	}
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestKubernetesVersion(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	kubernetesVersion, err := nodeJoinRequest.Encrypt(cluster.KubernetesVersion)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubernetesVersion = kubernetesVersion
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestVPNAddressAndPeers(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	if cluster.VPN == nil {
		return nil
	}
	nodeJoinRequest.VPN = &nodejoinrequests.VPN{}
	vpnCIDR, err := nodeJoinRequest.Encrypt(cluster.VPN.CIDR.String())
	if err != nil {
		return err
	}
	nodeJoinRequest.VPN.CIDR = vpnCIDR
	vpnPeer, err := cluster.GenerateVPNPeer(fmt.Sprintf("worker-%s", nodeJoinRequest.Name))
	if err != nil {
		return err
	}
	vpnAddress, err := nodeJoinRequest.Encrypt(vpnPeer.Address)
	if err != nil {
		return err
	}
	nodeJoinRequest.VPN.Address = vpnAddress
	vpnPeerPrivateKey, err := nodeJoinRequest.Encrypt(vpnPeer.PrivateKey)
	if err != nil {
		return err
	}
	nodeJoinRequest.VPN.PeerPrivateKey = vpnPeerPrivateKey
	ingressVPNEndpoint, err := nodeJoinRequest.Encrypt(cluster.VPNServerEndpoint)
	if err != nil {
		return err
	}
	nodeJoinRequest.VPN.Endpoint = ingressVPNEndpoint
	ingressVPNEndpointRaw, exists := cluster.VPNPeers[constants.OneInfraControlPlaneIngressVPNPeerName]
	if !exists {
		return err
	}
	endpointPublicKey, err := nodeJoinRequest.Encrypt(ingressVPNEndpointRaw.PublicKey)
	if err != nil {
		return err
	}
	nodeJoinRequest.VPN.EndpointPublicKey = endpointPublicKey
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestKubeConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	apiServerEndpoint := nodeJoinRequest.APIServerEndpoint
	if apiServerEndpoint == "" {
		apiServerEndpoint = cluster.APIServerEndpoint
	}
	kubeConfig, err := cluster.KubeConfigWithEndpoint(
		apiServerEndpoint,
		fmt.Sprintf("system:node:%s", nodeJoinRequest.Name),
		[]string{"system:nodes"},
	)
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

func (cluster *Cluster) fillNodeJoinRequestKubeletServerCertificate(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	extraSANs := nodeJoinRequest.ExtraSANs
	extraSANs = append(
		extraSANs,
		nodeJoinRequest.Name,
	)
	certificate, privateKey, err := cluster.CertificateAuthorities.Kubelet.CreateCertificate(
		nodeJoinRequest.Name,
		[]string{cluster.Name},
		extraSANs,
	)
	if err != nil {
		return err
	}
	kubeletServerCertificate, err := nodeJoinRequest.Encrypt(certificate)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubeletServerCertificate = kubeletServerCertificate
	kubeletServerPrivateKey, err := nodeJoinRequest.Encrypt(privateKey)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubeletServerPrivateKey = kubeletServerPrivateKey
	return nil
}

func (cluster *Cluster) fillNodeJoinRequestKubeletClientCACertificate(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	kubeletClientCACertificate, err := nodeJoinRequest.Encrypt(cluster.CertificateAuthorities.KubeletClient.Certificate)
	if err != nil {
		return err
	}
	nodeJoinRequest.KubeletClientCACertificate = kubeletClientCACertificate
	return nil
}
