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

package component

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"k8s.io/klog"

	"oneinfra.ereslibre.es/m/internal/pkg/infra/pod"
	"oneinfra.ereslibre.es/m/internal/pkg/inquirer"
)

const (
	kubeAPIServerImage         = "k8s.gcr.io/kube-apiserver:v1.17.0"
	kubeControllerManagerImage = "k8s.gcr.io/kube-controller-manager:v1.17.0"
	kubeSchedulerImage         = "k8s.gcr.io/kube-scheduler:v1.17.0"
)

// ControlPlane represents a complete control plane instance,
// including: etcd, API server, controller-manager and scheduler
type ControlPlane struct{}

// Reconcile reconciles the kube-apiserver
func (controlPlane *ControlPlane) Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	node := inquirer.Node()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	klog.V(1).Infof("reconciling control plane in node %q, present in hypervisor %q, belonging to cluster %q", node.Name, hypervisor.Name, cluster.Name)
	if err := hypervisor.EnsureImages(etcdImage, kubeAPIServerImage, kubeControllerManagerImage, kubeSchedulerImage); err != nil {
		return err
	}
	controllerManagerKubeConfig, err := cluster.KubeConfig("https://127.0.0.1:6443")
	if err != nil {
		return err
	}
	schedulerKubeConfig, err := cluster.KubeConfig("https://127.0.0.1:6443")
	if err != nil {
		return err
	}
	err = hypervisor.UploadFiles(
		map[string]string{
			// API server secrets
			secretsPathFile(cluster.Name, "apiserver-client-ca.crt"): cluster.CertificateAuthorities.APIServerClient.Certificate,
			secretsPathFile(cluster.Name, "apiserver.crt"):           cluster.APIServer.TLSCert,
			secretsPathFile(cluster.Name, "apiserver.key"):           cluster.APIServer.TLSPrivateKey,
			secretsPathFile(cluster.Name, "service-account-pub.key"): cluster.APIServer.ServiceAccountPublicKey,
			// controller-manager secrets
			secretsPathFile(cluster.Name, "controller-manager.kubeconfig"): controllerManagerKubeConfig,
			secretsPathFile(cluster.Name, "service-account.key"):           cluster.APIServer.ServiceAccountPrivateKey,
			// scheduler secrets
			secretsPathFile(cluster.Name, "scheduler.kubeconfig"): schedulerKubeConfig,
		},
	)
	if err != nil {
		return err
	}
	if err := controlPlane.runEtcd(inquirer); err != nil {
		return err
	}
	apiserverHostPort, ok := node.AllocatedHostPorts["apiserver"]
	if !ok {
		return errors.New("apiserver host port not found")
	}
	etcdClientHostPort, ok := node.AllocatedHostPorts["etcd-client"]
	if !ok {
		return errors.New("etcd client host port not found")
	}
	etcdServers := url.URL{Scheme: "http", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdClientHostPort))}
	_, err = hypervisor.RunPod(
		cluster,
		pod.NewPod(
			fmt.Sprintf("control-plane-%s", cluster.Name),
			[]pod.Container{
				{
					Name:    "kube-apiserver",
					Image:   kubeAPIServerImage,
					Command: []string{"kube-apiserver"},
					Args: []string{
						// Each API server accesses the local etcd node only, to
						// avoid reconfigurations; this could be improved in the
						// future though, to reconfigure them pointing to all
						// available etcd instances
						"--etcd-servers", etcdServers.String(),
						"--anonymous-auth", "false",
						"--authorization-mode", "Node,RBAC",
						"--allow-privileged", "true",
						"--tls-cert-file", secretsPathFile(cluster.Name, "apiserver.crt"),
						"--tls-private-key-file", secretsPathFile(cluster.Name, "apiserver.key"),
						"--client-ca-file", secretsPathFile(cluster.Name, "apiserver-client-ca.crt"),
						"--service-account-key-file", secretsPathFile(cluster.Name, "service-account-pub.key"),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name): secretsPath(cluster.Name),
					},
				},
				{
					Name:    "kube-controller-manager",
					Image:   kubeControllerManagerImage,
					Command: []string{"kube-controller-manager"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster.Name, "controller-manager.kubeconfig"),
						"--service-account-private-key-file", secretsPathFile(cluster.Name, "service-account.key"),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name): secretsPath(cluster.Name),
					},
				},
				{
					Name:    "kube-scheduler",
					Image:   kubeSchedulerImage,
					Command: []string{"kube-scheduler"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster.Name, "scheduler.kubeconfig"),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name): secretsPath(cluster.Name),
					},
				},
			},
			map[int]int{
				apiserverHostPort: 6443,
			},
		),
	)
	return err
}
