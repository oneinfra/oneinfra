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
	"net"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// VPN represents the VPN configuration
type VPN struct {
	Enabled bool
	CIDR    *net.IPNet
}

// VPNPeer represents a VPN peer
type VPNPeer struct {
	Name       string
	Address    string
	PrivateKey string
	PublicKey  string
}

// VPNPeerMap represents a map of VPN peers
type VPNPeerMap map[string]*VPNPeer

func newVPNFromv1alpha1(vpn *clusterv1alpha1.VPN) *VPN {
	if vpn == nil || vpn.CIDR == nil {
		return &VPN{
			Enabled: false,
		}
	}
	return &VPN{
		Enabled: vpn.Enabled,
		CIDR:    newVPNCIDRFromv1alpha1(*vpn.CIDR),
	}
}

// Export exports this VPN to a versioned VPN
func (vpn *VPN) Export() *clusterv1alpha1.VPN {
	if vpn.CIDR == nil {
		return &clusterv1alpha1.VPN{
			Enabled: false,
		}
	}
	vpnCIDR := vpn.CIDR.String()
	return &clusterv1alpha1.VPN{
		Enabled: vpn.Enabled,
		CIDR:    &vpnCIDR,
	}
}

func newVPNPeersFromv1alpha1(peers []clusterv1alpha1.VPNPeer) VPNPeerMap {
	res := VPNPeerMap{}
	for _, peer := range peers {
		res[peer.Name] = &VPNPeer{
			Name:       peer.Name,
			Address:    peer.Address,
			PrivateKey: peer.PrivateKey,
			PublicKey:  peer.PublicKey,
		}
	}
	return res
}

func newVPNCIDRFromv1alpha1(vpnCIDR string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(vpnCIDR)
	if err != nil {
		return &net.IPNet{}
	}
	return ipNet
}

func newVPNPeer(name, address, privateKey, publicKey string) *VPNPeer {
	return &VPNPeer{
		Name:       name,
		Address:    address,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}

// Export exports the map of VPN peers to a versioned list of VPN peers
func (vpnPeerMap VPNPeerMap) Export() []clusterv1alpha1.VPNPeer {
	res := []clusterv1alpha1.VPNPeer{}
	for _, peer := range vpnPeerMap {
		res = append(res, clusterv1alpha1.VPNPeer{
			Name:       peer.Name,
			Address:    peer.Address,
			PrivateKey: peer.PrivateKey,
			PublicKey:  peer.PublicKey,
		})
	}
	return res
}

// ReconcileMinimalVPNPeers reconciles a minimal list of VPN peers
func (cluster *Cluster) ReconcileMinimalVPNPeers() error {
	_, err := cluster.GenerateVPNPeer(constants.OneInfraControlPlaneIngressVPNPeerName)
	return err
}
