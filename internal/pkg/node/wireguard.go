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

package node

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"text/template"

	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
)

const (
	wireguardSystemdServiceTemplate = `[Unit]
Description=oneinfra wireguard configuration
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/bash {{ .WireguardScriptPath }}

[Install]
WantedBy=multi-user.target
`

	wireguardScriptTemplate = `if ! which ip &> /dev/null; then
  echo "ip executable not found, please install iproute2 in this node"
  exit 1
fi
if ! which wg &> /dev/null; then
  echo "wg executable not found, please install wireguard-tools in this node"
  exit 1
fi
ip link del oi-wg
ip link add dev oi-wg type wireguard
ip addr add {{ .Address }} dev oi-wg
wg set oi-wg private-key {{ .PeerPrivateKeyPath }} peer {{ .EndpointPublicKey }} endpoint {{ .Endpoint }} allowed-ips {{ .CIDR }} persistent-keepalive 20
ip link set oi-wg up
`
)

func wireguardScriptContents(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey crypto.SymmetricKey) (string, error) {
	cidr, err := decrypt(symmetricKey, nodeJoinRequest.VPN.CIDR)
	if err != nil {
		return "", err
	}
	peerAddress, err := decrypt(symmetricKey, nodeJoinRequest.VPN.Address)
	if err != nil {
		return "", err
	}
	endpointAddress, err := decrypt(symmetricKey, nodeJoinRequest.VPN.Endpoint)
	if err != nil {
		return "", err
	}
	template, err := template.New("").Parse(wireguardScriptTemplate)
	if err != nil {
		return "", err
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, struct {
		CIDR               string
		Address            string
		PeerPrivateKeyPath string
		Endpoint           string
		EndpointPublicKey  string
	}{
		CIDR:               cidr,
		Address:            peerAddress,
		PeerPrivateKeyPath: peerPrivateKeyPath,
		Endpoint:           endpointAddress,
		EndpointPublicKey:  nodeJoinRequest.VPN.EndpointPublicKey,
	})
	return rendered.String(), err
}

func wireguardSystemdServiceContents() (string, error) {
	template, err := template.New("").Parse(wireguardSystemdServiceTemplate)
	if err != nil {
		return "", err
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, struct {
		WireguardScriptPath string
	}{
		WireguardScriptPath: wireguardScriptPath,
	})
	return rendered.String(), err
}

func setupWireguard(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey crypto.SymmetricKey) error {
	wireguardSystemdServiceContents, err := wireguardSystemdServiceContents()
	if err != nil {
		return err
	}
	wireguardScriptContents, err := wireguardScriptContents(nodeJoinRequest, symmetricKey)
	if err != nil {
		return err
	}
	peerPrivateKey, err := decrypt(symmetricKey, nodeJoinRequest.VPN.PeerPrivateKey)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(peerPrivateKeyPath, []byte(peerPrivateKey), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(wireguardScriptPath, []byte(wireguardScriptContents), 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(wireguardSystemdServicePath, []byte(wireguardSystemdServiceContents), 0600); err != nil {
		return err
	}
	return exec.Command("systemctl", "enable", "--now", "oi-wg").Run()
}
