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
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
)

const (
	kubeAPIServerImage         = "k8s.gcr.io/kube-apiserver:v%s"
	kubeControllerManagerImage = "k8s.gcr.io/kube-controller-manager:v%s"
	kubeSchedulerImage         = "k8s.gcr.io/kube-scheduler:v%s"
)

// ControlPlane represents a complete control plane instance,
// including: etcd, API server, controller-manager and scheduler
type ControlPlane struct{}

// Reconcile reconciles the kube-apiserver
func (controlPlane *ControlPlane) Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	clusterLoadBalancers := inquirer.ClusterComponents(componentapi.ControlPlaneIngressRole)
	if !clusterLoadBalancers.AllWithHypervisorAssigned() {
		return errors.Errorf("could not reconcile component %q, not all load balancers have an hypervisor assigned", component.Name)
	}
	kubernetesVersion := inquirer.Cluster().KubernetesVersion
	versionBundle, err := constants.KubernetesVersionBundle(kubernetesVersion)
	if err != nil {
		return errors.Errorf("could not retrieve version bundle for version %q", kubernetesVersion)
	}
	klog.V(1).Infof("reconciling component %q, present in hypervisor %q, belonging to cluster %q", component.Name, hypervisor.Name, cluster.Name)
	err = hypervisor.EnsureImages(
		fmt.Sprintf(etcdImage, versionBundle.EtcdVersion),
		fmt.Sprintf(kubeAPIServerImage, kubernetesVersion),
		fmt.Sprintf(kubeControllerManagerImage, kubernetesVersion),
		fmt.Sprintf(kubeSchedulerImage, kubernetesVersion),
	)
	if err != nil {
		return err
	}
	etcdAPIServerClientCertificate, err := component.ClientCertificate(
		cluster.CertificateAuthorities.EtcdClient,
		"apiserver-etcd-client",
		fmt.Sprintf("apiserver-etcd-client-%s", component.Name),
		[]string{cluster.Name},
		[]string{},
	)
	if err != nil {
		return err
	}
	kubeAPIServerExtraSANs := cluster.APIServer.ExtraSANs
	// Add all load balancer IP addresses. This is necessary if the
	// ingress is operating at L4
	for _, clusterLoadBalancer := range clusterLoadBalancers {
		loadBalancerHypervisor := inquirer.ComponentHypervisor(clusterLoadBalancer)
		if loadBalancerHypervisor == nil {
			continue
		}
		kubeAPIServerExtraSANs = append(
			kubeAPIServerExtraSANs,
			loadBalancerHypervisor.IPAddress,
		)
	}
	apiServerCertificate, err := component.ServerCertificate(
		cluster.APIServer.CA,
		"kube-apiserver",
		"kube-apiserver",
		[]string{"kube-apiserver"},
		kubeAPIServerExtraSANs,
	)
	if err != nil {
		return err
	}
	controllerManagerKubeConfig, err := component.KubeConfig(cluster, "https://127.0.0.1:6443", "controller-manager")
	if err != nil {
		return err
	}
	schedulerKubeConfig, err := component.KubeConfig(cluster, "https://127.0.0.1:6443", "scheduler")
	if err != nil {
		return err
	}
	err = hypervisor.UploadFiles(
		cluster.Name,
		component.Name,
		map[string]string{
			// etcd secrets
			secretsPathFile(cluster.Name, component.Name, "etcd-ca.crt"):               cluster.EtcdServer.CA.Certificate,
			secretsPathFile(cluster.Name, component.Name, "apiserver-etcd-client.crt"): etcdAPIServerClientCertificate.Certificate,
			secretsPathFile(cluster.Name, component.Name, "apiserver-etcd-client.key"): etcdAPIServerClientCertificate.PrivateKey,
			// API server secrets
			secretsPathFile(cluster.Name, component.Name, "apiserver-client-ca.crt"): cluster.CertificateAuthorities.APIServerClient.Certificate,
			secretsPathFile(cluster.Name, component.Name, "apiserver.crt"):           apiServerCertificate.Certificate,
			secretsPathFile(cluster.Name, component.Name, "apiserver.key"):           apiServerCertificate.PrivateKey,
			secretsPathFile(cluster.Name, component.Name, "service-account-pub.key"): cluster.APIServer.ServiceAccount.PublicKey,
			// controller-manager secrets
			secretsPathFile(cluster.Name, component.Name, "controller-manager.kubeconfig"): controllerManagerKubeConfig,
			secretsPathFile(cluster.Name, component.Name, "service-account.key"):           cluster.APIServer.ServiceAccount.PrivateKey,
			secretsPathFile(cluster.Name, component.Name, "cluster-signing-ca.crt"):        cluster.CertificateAuthorities.CertificateSigner.Certificate,
			secretsPathFile(cluster.Name, component.Name, "cluster-signing-ca.key"):        cluster.CertificateAuthorities.CertificateSigner.PrivateKey,
			// scheduler secrets
			secretsPathFile(cluster.Name, component.Name, "scheduler.kubeconfig"): schedulerKubeConfig,
		},
	)
	if err != nil {
		return err
	}
	apiserverHostPort, err := component.RequestPort(hypervisor, "apiserver")
	if err != nil {
		return err
	}
	if err := controlPlane.runEtcd(inquirer); err != nil {
		return err
	}
	etcdClientHostPort, exists := component.AllocatedHostPorts["etcd-client"]
	if !exists {
		return errors.New("etcd client host port not found")
	}
	etcdServers := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdClientHostPort))}
	_, err = hypervisor.RunPod(
		cluster,
		pod.NewPod(
			fmt.Sprintf("control-plane-%s", cluster.Name),
			[]pod.Container{
				{
					Name:    "kube-apiserver",
					Image:   fmt.Sprintf(kubeAPIServerImage, kubernetesVersion),
					Command: []string{"kube-apiserver"},
					Args: []string{
						// Each API server accesses the local etcd component only, to
						// avoid reconfigurations; this could be improved in the
						// future though, to reconfigure them pointing to all
						// available etcd instances
						"--etcd-servers", etcdServers.String(),
						"--etcd-cafile", secretsPathFile(cluster.Name, component.Name, "etcd-ca.crt"),
						"--etcd-certfile", secretsPathFile(cluster.Name, component.Name, "apiserver-etcd-client.crt"),
						"--etcd-keyfile", secretsPathFile(cluster.Name, component.Name, "apiserver-etcd-client.key"),
						"--anonymous-auth", "false",
						"--authorization-mode", "Node,RBAC",
						"--enable-bootstrap-token-auth",
						"--allow-privileged", "true",
						"--tls-cert-file", secretsPathFile(cluster.Name, component.Name, "apiserver.crt"),
						"--tls-private-key-file", secretsPathFile(cluster.Name, component.Name, "apiserver.key"),
						"--client-ca-file", secretsPathFile(cluster.Name, component.Name, "apiserver-client-ca.crt"),
						"--service-account-key-file", secretsPathFile(cluster.Name, component.Name, "service-account-pub.key"),
						"--kubelet-preferred-address-types", "ExternalIP,ExternalDNS,Hostname,InternalDNS,InternalIP",
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name, component.Name): secretsPath(cluster.Name, component.Name),
					},
				},
				{
					Name:    "kube-controller-manager",
					Image:   fmt.Sprintf(kubeControllerManagerImage, kubernetesVersion),
					Command: []string{"kube-controller-manager"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster.Name, component.Name, "controller-manager.kubeconfig"),
						"--controllers=*,tokencleaner",
						"--service-account-private-key-file", secretsPathFile(cluster.Name, component.Name, "service-account.key"),
						"--cluster-signing-cert-file", secretsPathFile(cluster.Name, component.Name, "cluster-signing-ca.crt"),
						"--cluster-signing-key-file", secretsPathFile(cluster.Name, component.Name, "cluster-signing-ca.key"),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name, component.Name): secretsPath(cluster.Name, component.Name),
					},
				},
				{
					Name:    "kube-scheduler",
					Image:   fmt.Sprintf(kubeSchedulerImage, kubernetesVersion),
					Command: []string{"kube-scheduler"},
					Args: []string{
						"--kubeconfig", secretsPathFile(cluster.Name, component.Name, "scheduler.kubeconfig"),
					},
					Mounts: map[string]string{
						secretsPath(cluster.Name, component.Name): secretsPath(cluster.Name, component.Name),
					},
				},
			},
			map[int]int{
				apiserverHostPort: 6443,
			},
			pod.PrivilegesUnprivileged,
		),
	)
	return err
}
