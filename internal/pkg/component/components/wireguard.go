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

package components

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"

	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"k8s.io/klog"
)

const (
	wireguardImage = "oneinfra/wireguard:latest"
)

const (
	wireguardTemplate = `[Interface]
Address = {{ .Address }}
ListenPort = {{ .ListenPort }}
PrivateKey = {{ .PrivateKey }}

{{- range $peer := .Peers }}
[Peer]
PublicKey = {{ $peer.PublicKey }}
AllowedIPs = {{ $peer.AllowedIPs }}
{{- end }}
`
)

func (ingress *ControlPlaneIngress) wireguardConfiguration(inquirer inquirer.ReconcilerInquirer) (string, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	wireguardHostPort, err := component.RequestPort(hypervisor, "wireguard")
	if err != nil {
		return "", err
	}
	vpnPeer, err := cluster.VPNPeer("control-plane-ingress")
	if err != nil {
		return "", err
	}
	template, err := template.New("").Parse(wireguardTemplate)
	if err != nil {
		return "", err
	}
	wireguardConfigData := struct {
		Address    string
		ListenPort string
		PrivateKey string
		Peers      []struct {
			PublicKey  string
			AllowedIPs string
		}
	}{
		Address:    vpnPeer.Address,
		ListenPort: strconv.Itoa(wireguardHostPort),
		PrivateKey: vpnPeer.PrivateKey,
		Peers: []struct {
			PublicKey  string
			AllowedIPs string
		}{},
	}
	for _, vpnPeer := range cluster.VPNPeers {
		if vpnPeer.Name == "control-plane-ingress" {
			continue
		}
		wireguardConfigData.Peers = append(wireguardConfigData.Peers, struct {
			PublicKey  string
			AllowedIPs string
		}{
			PublicKey:  vpnPeer.PublicKey,
			AllowedIPs: vpnPeer.Address,
		})
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, wireguardConfigData)
	return rendered.String(), err
}

func (ingress *ControlPlaneIngress) reconcileWireguard(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	if err := hypervisor.EnsureImage(wireguardImage); err != nil {
		return err
	}
	wireguardConfig, err := ingress.wireguardConfiguration(inquirer)
	if err != nil {
		return err
	}
	wireguardConfigPath := secretsPathFile(cluster.Name, component.Name, fmt.Sprintf("wg-%s.conf", cluster.Name))
	if hypervisor.FileUpToDate(cluster.Name, component.Name, wireguardConfigPath, wireguardConfig) {
		klog.V(2).Info("skipping wireguard reconfiguration, since configuration is up to date")
		return nil
	}
	err = hypervisor.UploadFile(
		cluster.Name,
		component.Name,
		wireguardConfigPath,
		wireguardConfig,
	)
	if err != nil {
		return err
	}
	return hypervisor.RunAndWaitForPod(
		cluster.Name,
		component.Name,
		pod.NewPod(
			fmt.Sprintf("wireguard-%s", cluster.Name),
			[]pod.Container{
				{
					Name:    "wireguard",
					Image:   wireguardImage,
					Command: []string{"wg-quick"},
					Args: []string{
						"up",
						secretsPathFile(cluster.Name, component.Name, fmt.Sprintf("wg-%s.conf", cluster.Name)),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name, component.Name): secretsPath(cluster.Name, component.Name),
					},
					Privileges: pod.PrivilegesNetworkPrivileged,
				},
			},
			map[int]int{},
			pod.PrivilegesNetworkPrivileged,
		),
	)
}
