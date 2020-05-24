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

package components

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"k8s.io/klog"
)

const (
	// WireguardHostPortName represents the wireguard host port
	// allocation name
	WireguardHostPortName = "wireguard"
)

const (
	wireguardSystemdServiceTemplate = `[Unit]
Description=oneinfra {{ .ClusterNamespace }}-{{ .ClusterName }} wireguard configuration
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/bash {{ .WireguardScriptPath }}

[Install]
WantedBy=multi-user.target
`

	wireguardScriptTemplate = `if ! which ip &> /dev/null; then
  echo "ip executable not found, please install iproute2 in this hypervisor"
  exit 1
fi
if ! which wg &> /dev/null; then
  echo "wg executable not found, please install wireguard-tools in this hypervisor"
  exit 1
fi
if ! ip netns pids {{ .NetworkNamespace }} &> /dev/null; then
  ip netns add {{ .NetworkNamespace }}
fi
if ! ip netns exec {{ .NetworkNamespace }} ip a show dev {{ .WireguardInterfaceName }} &> /dev/null; then
  ip link del {{ .WireguardInterfaceName }}
  ip link add dev {{ .WireguardInterfaceName }} type wireguard
  ip addr add {{ .VPNCIDR }} dev {{ .WireguardInterfaceName }}
  wg set {{ .WireguardInterfaceName }} listen-port {{ .ListenPort }} private-key {{ .PrivateKeyPath }}
  ip link set {{ .WireguardInterfaceName }} netns {{ .NetworkNamespace }}
  ip netns exec {{ .NetworkNamespace }} ip link set {{ .WireguardInterfaceName }} up
  ip netns exec {{ .NetworkNamespace }} ip route add default dev {{ .WireguardInterfaceName }}
fi
{{ $data := . }}
{{- range $peer := .Peers }}
ip netns exec {{ $data.NetworkNamespace }} wg set {{ $data.WireguardInterfaceName }} peer {{ $peer.PublicKey }} allowed-ips {{ $peer.AllowedIPs }}
{{- end }}
`
)

func (ingress *ControlPlaneIngress) reconcileWireguard(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	wireguardSystemdServiceContents, err := ingress.wireguardSystemdServiceContents(inquirer)
	if err != nil {
		return err
	}
	wireguardScriptContents, err := ingress.wireguardScriptContents(inquirer)
	if err != nil {
		klog.Fatal(err)
		return err
	}
	privateKeyPath := componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "wg.key")
	err = hypervisor.UploadFiles(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		map[string]string{
			ingress.wireguardSystemdServicePath(inquirer): wireguardSystemdServiceContents,
			ingress.wireguardScriptPath(inquirer):         wireguardScriptContents,
			privateKeyPath:                                cluster.VPN.PrivateKey,
		},
	)
	if err != nil {
		return err
	}
	if err := hypervisor.EnsureImage(infra.ToolingImage); err != nil {
		return err
	}
	enableAndStartArgs := []string{
		fmt.Sprintf("dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.EnableUnitFiles array:string:'%s' boolean:false boolean:true", ingress.wireguardSystemdServiceName(inquirer)),
		fmt.Sprintf("dbus-send --system --print-reply --dest=org.freedesktop.systemd1 /org/freedesktop/systemd1 org.freedesktop.systemd1.Manager.StartUnit string:'%s' string:'replace'", ingress.wireguardSystemdServiceName(inquirer)),
	}
	return hypervisor.RunAndWaitForPod(cluster.Namespace, cluster.Name, component.Name, podapi.Pod{
		Name: fmt.Sprintf("enable-and-start-wireguard-%s-%s", cluster.Namespace, cluster.Name),
		Containers: []podapi.Container{
			{
				Name:    fmt.Sprintf("enable-and-start-wireguard-%s-%s", cluster.Namespace, cluster.Name),
				Image:   infra.ToolingImage,
				Command: []string{"/bin/sh", "-c"},
				Args: []string{
					strings.Join(enableAndStartArgs, "&&"),
				},
				Mounts: map[string]string{
					"/var/run/dbus": "/var/run/dbus",
				},
				Privileges: podapi.PrivilegesPrivileged,
			},
		},
		Privileges: podapi.PrivilegesPrivileged,
	})
}

func (ingress *ControlPlaneIngress) netNSName(inquirer inquirer.ReconcilerInquirer) string {
	cluster := inquirer.Cluster()
	return fmt.Sprintf("oneinfra-wireguard-%s-%s", cluster.Namespace, cluster.Name)
}

func (ingress *ControlPlaneIngress) wireguardInterfaceName(inquirer inquirer.ReconcilerInquirer) string {
	cluster := inquirer.Cluster()
	return fmt.Sprintf("wg-%x", sha1.Sum([]byte(fmt.Sprintf("%s-%s", cluster.Namespace, cluster.Name))))[0:15]
}

func (ingress *ControlPlaneIngress) wireguardSystemdServiceName(inquirer inquirer.ReconcilerInquirer) string {
	cluster := inquirer.Cluster()
	return fmt.Sprintf("oneinfra-wireguard-%s-%s.service", cluster.Namespace, cluster.Name)
}

func (ingress *ControlPlaneIngress) wireguardSystemdServicePath(inquirer inquirer.ReconcilerInquirer) string {
	return filepath.Join("/etc/systemd/system", ingress.wireguardSystemdServiceName(inquirer))
}

func (ingress *ControlPlaneIngress) wireguardScriptPath(inquirer inquirer.ReconcilerInquirer) string {
	cluster := inquirer.Cluster()
	component := inquirer.Component()
	return componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "wg.sh")
}

func (ingress *ControlPlaneIngress) wireguardSystemdServiceContents(inquirer inquirer.ReconcilerInquirer) (string, error) {
	cluster := inquirer.Cluster()
	wireguardServiceData := struct {
		ClusterNamespace    string
		ClusterName         string
		WireguardScriptPath string
	}{
		ClusterNamespace:    cluster.Namespace,
		ClusterName:         cluster.Name,
		WireguardScriptPath: ingress.wireguardScriptPath(inquirer),
	}
	var rendered bytes.Buffer
	template, err := template.New("").Parse(wireguardSystemdServiceTemplate)
	err = template.Execute(&rendered, wireguardServiceData)
	return rendered.String(), err
}

func (ingress *ControlPlaneIngress) wireguardScriptContents(inquirer inquirer.ReconcilerInquirer) (string, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	wireguardHostPort, err := component.RequestPort(hypervisor, WireguardHostPortName)
	if err != nil {
		return "", err
	}
	privateKeyPath := componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "wg.key")
	wireguardScriptData := struct {
		WireguardInterfaceName string
		VPNCIDR                string
		ListenPort             string
		PrivateKeyPath         string
		NetworkNamespace       string
		Peers                  []struct {
			PublicKey  string
			AllowedIPs string
		}
	}{
		WireguardInterfaceName: ingress.wireguardInterfaceName(inquirer),
		VPNCIDR:                cluster.VPN.CIDR.String(),
		ListenPort:             strconv.Itoa(wireguardHostPort),
		PrivateKeyPath:         privateKeyPath,
		NetworkNamespace:       ingress.netNSName(inquirer),
		Peers: []struct {
			PublicKey  string
			AllowedIPs string
		}{},
	}
	for _, vpnPeer := range cluster.VPNPeers {
		ipAddress, _, err := net.ParseCIDR(vpnPeer.Address)
		if err != nil {
			continue
		}
		var ipAddressNet net.IPNet
		if len(ipAddress) == net.IPv6len {
			ipAddressNet = net.IPNet{IP: ipAddress, Mask: net.CIDRMask(128, 128)}
		} else {
			ipAddressNet = net.IPNet{IP: ipAddress, Mask: net.CIDRMask(32, 32)}
		}
		wireguardScriptData.Peers = append(
			wireguardScriptData.Peers,
			struct {
				PublicKey  string
				AllowedIPs string
			}{
				PublicKey:  vpnPeer.PublicKey,
				AllowedIPs: ipAddressNet.String(),
			},
		)
	}
	var rendered bytes.Buffer
	template, err := template.New("").Parse(wireguardScriptTemplate)
	err = template.Execute(&rendered, wireguardScriptData)
	return rendered.String(), err
}
