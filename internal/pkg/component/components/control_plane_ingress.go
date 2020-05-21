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
	"net/url"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/klog"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
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

// PreReconcile pre-reconciles the control plane ingress component
func (ingress *ControlPlaneIngress) PreReconcile(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	if component.HypervisorName == "" {
		return errors.Errorf("could not pre-reconcile component %q; no hypervisor assigned yet", component.Name)
	}
	hypervisor := inquirer.Hypervisor()
	if _, err := component.RequestPort(hypervisor, APIServerHostPortName); err != nil {
		return err
	}
	cluster := inquirer.Cluster()
	if cluster.VPN.Enabled {
		if _, err := component.RequestPort(hypervisor, WireguardHostPortName); err != nil {
			return err
		}
	}
	return nil
}

func (ingress *ControlPlaneIngress) reconcileInputAndOutputEndpoints(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	apiserverHostPort, err := component.RequestPort(hypervisor, APIServerHostPortName)
	if err != nil {
		return err
	}
	outputEndpointURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort))}
	component.OutputEndpoints = map[string]string{
		component.Name: outputEndpointURL.String(),
	}
	clusterComponents := inquirer.ClusterComponents(componentapi.ControlPlaneRole)
	component.InputEndpoints = map[string]string{}
	for _, clusterComponent := range clusterComponents {
		if clusterComponent.DeletionTimestamp != nil {
			continue
		}
		hypervisor := inquirer.ComponentHypervisor(clusterComponent)
		if hypervisor == nil {
			continue
		}
		apiserverHostPort, err := clusterComponent.RequestPort(hypervisor, APIServerHostPortName)
		if err != nil {
			return err
		}
		inputEndpointURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(apiserverHostPort))}
		component.InputEndpoints[clusterComponent.Name] = inputEndpointURL.String()
	}
	return nil
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
	if len(clusterComponents) != cluster.ControlPlaneReplicas {
		return errors.Errorf("could not reconcile component %q, control plane replicas do not match with current listed number of components", component.Name)
	}
	klog.V(1).Infof("reconciling control plane ingress in component %q, present in hypervisor %q, belonging to cluster %q", component.Name, hypervisor.Name, cluster.Name)
	if err := ingress.reconcileInputAndOutputEndpoints(inquirer); err != nil {
		return err
	}
	if err := hypervisor.EnsureImage(haProxyImage); err != nil {
		return err
	}
	apiserverHostPort, err := component.RequestPort(hypervisor, APIServerHostPortName)
	if err != nil {
		return err
	}
	haProxyConfig, err := ingress.haProxyConfiguration(inquirer, clusterComponents)
	if err != nil {
		return err
	}
	err = hypervisor.UploadFile(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "haproxy.cfg"),
		haProxyConfig,
	)
	if err != nil {
		return err
	}
	_, err = hypervisor.EnsurePod(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		pod.NewPod(
			ingress.ingressPodName(inquirer),
			[]pod.Container{
				{
					Name:  "haproxy",
					Image: haProxyImage,
					Mounts: map[string]string{
						componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "haproxy.cfg"): "/etc/haproxy/haproxy.cfg",
					},
					Annotations: map[string]string{
						"oneinfra/haproxy-config-sha1sum": fmt.Sprintf("%x", sha1.Sum([]byte(haProxyConfig))),
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
	if cluster.VPN.Enabled {
		if err := ingress.reconcileWireguard(inquirer); err != nil {
			return err
		}
	}
	return nil
}

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
		apiserverHostPort, exists := component.AllocatedHostPorts[APIServerHostPortName]
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

func (ingress *ControlPlaneIngress) ingressPodName(inquirer inquirer.ReconcilerInquirer) string {
	return fmt.Sprintf("control-plane-ingress-%s", inquirer.Cluster().Name)
}

func (ingress *ControlPlaneIngress) stopIngress(inquirer inquirer.ReconcilerInquirer) error {
	err := inquirer.Hypervisor().DeletePod(
		inquirer.Cluster().Namespace,
		inquirer.Cluster().Name,
		inquirer.Component().Name,
		ingress.ingressPodName(inquirer),
	)
	if err == nil {
		component := inquirer.Component()
		cluster := inquirer.Cluster()
		hypervisor := inquirer.Hypervisor()
		if err := component.FreePort(hypervisor, APIServerHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", APIServerHostPortName, hypervisor.Name)
		}
		if cluster.VPN.Enabled {
			if err := component.FreePort(hypervisor, WireguardHostPortName); err != nil {
				return errors.Wrapf(err, "could not free port %q for hypervisor %q", WireguardHostPortName, hypervisor.Name)
			}
		}
	}
	return err
}

// ReconcileDeletion reconciles the control plane ingress deletion
func (ingress *ControlPlaneIngress) ReconcileDeletion(inquirer inquirer.ReconcilerInquirer) error {
	hypervisor := inquirer.Hypervisor()
	if hypervisor == nil {
		return nil
	}
	if err := ingress.stopIngress(inquirer); err != nil {
		return err
	}
	return ingress.hostCleanup(inquirer)
}

func (ingress *ControlPlaneIngress) hostCleanup(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	res := hypervisor.RunAndWaitForPod(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		pod.NewPod(
			fmt.Sprintf("%s-%s-%s-cleanup", cluster.Namespace, cluster.Name, component.Name),
			[]pod.Container{
				{
					Name:    "secrets-cleanup",
					Image:   infra.ToolingImage,
					Command: []string{"/bin/sh"},
					Args: []string{
						"-c",
						fmt.Sprintf(
							"rm -rf %s && ((rmdir %s && rmdir %s) || true)",
							componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
							clusterSecretsPath(cluster.Namespace, cluster.Name),
							namespacedClusterSecretsPath(cluster.Namespace),
						),
					},
					Mounts: map[string]string{
						globalSecretsPath(): globalSecretsPath(),
					},
				},
			},
			map[int]int{},
			pod.PrivilegesUnprivileged,
		),
	)
	if res == nil {
		cleanupHypervisorFileMap(hypervisor, cluster.Namespace, cluster.Name, component.Name)
	}
	return res
}
