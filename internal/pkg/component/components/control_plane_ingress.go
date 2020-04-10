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
	"net"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/klog"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
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

func (ingress *ControlPlaneIngress) haProxyConfiguration(inquirer inquirer.ReconcilerInquirer, clusterComponents componentapi.List) (string, error) {
	template, err := template.New("").Parse(haProxyTemplate)
	if err != nil {
		return "", err
	}
	haProxyConfigData := struct {
		APIServers map[string]string
	}{
		APIServers: map[string]string{},
	}
	for _, component := range clusterComponents {
		apiserverHostPort, exists := component.AllocatedHostPorts["apiserver"]
		if !exists {
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
	clusterComponents := inquirer.ClusterComponents(componentapi.ControlPlaneRole)
	if !clusterComponents.AllWithHypervisorAssigned() {
		return errors.Errorf("could not reconcile component %q, not all cluster components have an hypervisor assigned", component.Name)
	}
	klog.V(1).Infof("reconciling control plane ingress in component %q, present in hypervisor %q, belonging to cluster %q", component.Name, hypervisor.Name, cluster.Name)
	if err := hypervisor.EnsureImage(haProxyImage); err != nil {
		return err
	}
	apiserverHostPort, err := component.RequestPort(hypervisor, "apiserver")
	if err != nil {
		return err
	}
	haProxyConfig, err := ingress.haProxyConfiguration(inquirer, clusterComponents)
	if err != nil {
		return err
	}
	err = hypervisor.UploadFile(
		cluster.Name,
		component.Name,
		secretsPathFile(cluster.Name, component.Name, "haproxy.cfg"),
		haProxyConfig,
	)
	if err != nil {
		return err
	}
	_, err = hypervisor.RunPod(
		cluster.Name,
		component.Name,
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
	cluster.APIServerEndpoint = fmt.Sprintf("https://%s", net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort)))
	return ingress.reconcileWireguard(inquirer)
}
