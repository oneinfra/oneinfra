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
	"net"
	"testing"
)

func TestRequestVPNIP(t *testing.T) {
	var tests = []struct {
		cidr        string
		peers       VPNPeerMap
		expectedIP  string
		expectedErr bool
	}{
		{
			cidr:        "10.0.0.0/8",
			peers:       vpnPeers(0),
			expectedIP:  "10.0.0.1",
			expectedErr: false,
		},
		{
			cidr:        "10.0.0.0/8",
			peers:       vpnPeers(1),
			expectedIP:  "10.0.0.2",
			expectedErr: false,
		},
		{
			cidr:        "10.0.0.0/31",
			peers:       vpnPeers(1),
			expectedErr: true,
		},
		{
			cidr:        "10.0.0.0/24",
			peers:       vpnPeers(255),
			expectedErr: true,
		},
		{
			cidr:        "fd00::/8",
			peers:       vpnPeers(0),
			expectedIP:  "fd00::1",
			expectedErr: false,
		},
		{
			cidr:        "fd00::/8",
			peers:       vpnPeers(1),
			expectedIP:  "fd00::2",
			expectedErr: false,
		},
		{
			cidr:        "fd00::/127",
			peers:       vpnPeers(1),
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s (peers: %d)", tt.cidr, len(tt.peers)), func(t *testing.T) {
			_, cidrNetwork, err := net.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("could not parse CIDR %q", tt.cidr)
			}
			cluster := Cluster{VPNCIDR: cidrNetwork, VPNPeers: tt.peers}
			if ip, err := cluster.requestVPNIP(); (err != nil) != tt.expectedErr {
				t.Errorf("got %v error, was expecting: %v", err, tt.expectedErr)
			} else if ip != tt.expectedIP {
				t.Errorf("got %q, was expecting %q", ip, tt.expectedIP)
			}
		})
	}
}

func vpnPeers(peerNumber int) VPNPeerMap {
	res := VPNPeerMap{}
	for i := 0; i < peerNumber; i++ {
		peerName := fmt.Sprintf("peer-%d", i)
		res[peerName] = &VPNPeer{
			Name: peerName,
		}
	}
	return res
}
