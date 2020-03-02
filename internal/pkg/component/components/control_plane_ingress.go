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
	"errors"
	"fmt"
	"net"
	"strconv"
	"text/template"

	"k8s.io/klog"

	"oneinfra.ereslibre.es/m/internal/pkg/component"
	"oneinfra.ereslibre.es/m/internal/pkg/infra/pod"
	"oneinfra.ereslibre.es/m/internal/pkg/inquirer"
)

const (
	haProxyImage = "oneinfra/haproxy:latest"
)

const (
	haProxyTemplate = `global
  log /dev/log local0
  log /dev/log local1 notice
  daemon
defaults
  log global
  mode tcp
  option dontlognull
  timeout connect 10s
  timeout client  60s
  timeout server  60s
frontend control-plane
  bind *:6443
  default_backend apiservers
backend apiservers
  option httpchk GET /healthz
  {{ range $server, $address := .APIServers }}
  server {{ $server }} {{ $address }} check check-ssl verify none
  {{- end }}
`
)

// ControlPlaneIngress represents an endpoint to a set of control plane instances
type ControlPlaneIngress struct{}

func (ingress *ControlPlaneIngress) haProxyConfiguration(inquirer inquirer.ReconcilerInquirer) (string, error) {
	template, err := template.New("").Parse(haProxyTemplate)
	if err != nil {
		return "", err
	}
	haProxyConfigData := struct {
		APIServers map[string]string
	}{
		APIServers: map[string]string{},
	}
	clusterComponents := inquirer.ClusterComponents(component.ControlPlaneRole)
	for _, component := range clusterComponents {
		apiserverHostPort, ok := component.AllocatedHostPorts["apiserver"]
		if !ok {
			return "", errors.New("apiserver host port not found")
		}
		haProxyConfigData.APIServers[component.Name] = net.JoinHostPort(
			inquirer.ComponentHypervisor(component).IPAddress,
			strconv.Itoa(apiserverHostPort),
		)
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, haProxyConfigData)
	return rendered.String(), err
}

// Reconcile reconciles the control plane ingress
func (ingress *ControlPlaneIngress) Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	klog.V(1).Infof("reconciling control plane ingress in component %q, present in hypervisor %q, belonging to cluster %q", component.Name, hypervisor.Name, cluster.Name)
	if err := hypervisor.EnsureImage(haProxyImage); err != nil {
		return err
	}
	apiserverHostPort, err := component.RequestPort(hypervisor, "apiserver")
	if err != nil {
		return err
	}
	haProxyConfig, err := ingress.haProxyConfiguration(inquirer)
	if err != nil {
		return err
	}
	if err := hypervisor.UploadFile(haProxyConfig, secretsPathFile(cluster.Name, component.Name, "haproxy.cfg")); err != nil {
		return err
	}
	_, err = hypervisor.RunPod(
		cluster,
		pod.NewPod(
			fmt.Sprintf("control-plane-ingress-%s", cluster.Name),
			[]pod.Container{
				{
					Name:  "haproxy",
					Image: haProxyImage,
					Mounts: map[string]string{
						secretsPathFile(cluster.Name, component.Name, "haproxy.cfg"): "/etc/haproxy/haproxy.cfg",
					},
				},
			},
			map[int]int{
				apiserverHostPort: 6443,
			},
			pod.PrivilegesUnprivileged,
		),
	)
	if err != nil {
		return err
	}
	return ingress.reconcileWireguard(inquirer)
}
