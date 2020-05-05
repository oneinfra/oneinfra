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
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

const (
	// APIServerHostPortName represents the apiserver host port
	// allocation name
	APIServerHostPortName = "apiserver"
)

const (
	kubeAPIServerImage         = "k8s.gcr.io/kube-apiserver:v%s"
	kubeControllerManagerImage = "k8s.gcr.io/kube-controller-manager:v%s"
	kubeSchedulerImage         = "k8s.gcr.io/kube-scheduler:v%s"
)

// ControlPlane represents a complete control plane instance,
// including: etcd, API server, controller-manager and scheduler
type ControlPlane struct{}

// PreReconcile pre-reconciles the control plane component
func (controlPlane *ControlPlane) PreReconcile(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	if component.HypervisorName == "" {
		return errors.Errorf("could not pre-reconcile component %q; no hypervisor assigned yet", component.Name)
	}
	hypervisor := inquirer.Hypervisor()
	if _, err := component.RequestPort(hypervisor, APIServerHostPortName); err != nil {
		return err
	}
	if _, err := component.RequestPort(hypervisor, EtcdPeerHostPortName); err != nil {
		return err
	}
	if _, err := component.RequestPort(hypervisor, EtcdClientHostPortName); err != nil {
		return err
	}
	return nil
}

func (controlPlane *ControlPlane) reconcileInputAndOutputEndpoints(inquirer inquirer.ReconcilerInquirer) error {
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
	component.InputEndpoints = map[string]string{}
	return nil
}

// Reconcile reconciles the control plane component
func (controlPlane *ControlPlane) Reconcile(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	kubernetesVersion := inquirer.Cluster().KubernetesVersion
	versionBundle, err := constants.KubernetesVersionBundle(kubernetesVersion)
	if err != nil {
		return errors.Errorf("could not retrieve version bundle for version %q", kubernetesVersion)
	}
	klog.V(1).Infof("reconciling component %q, present in hypervisor %q, belonging to cluster %q", component.Name, hypervisor.Name, cluster.Name)
	if err := controlPlane.reconcileInputAndOutputEndpoints(inquirer); err != nil {
		return err
	}
	err = hypervisor.EnsureImages(
		fmt.Sprintf(etcdImage, versionBundle.EtcdVersion),
		fmt.Sprintf(kubeAPIServerImage, kubernetesVersion),
		fmt.Sprintf(kubeControllerManagerImage, kubernetesVersion),
		fmt.Sprintf(kubeSchedulerImage, kubernetesVersion),
	)
	if err != nil {
		return err
	}
	advertiseAddressHost, advertiseAddressPort, err := controlPlane.kubeAPIServerAdvertiseAddressAndPort(inquirer)
	if err != nil {
		return err
	}
	if err := controlPlane.uploadFiles(inquirer); err != nil {
		return err
	}
	apiserverHostPort, err := component.RequestPort(hypervisor, APIServerHostPortName)
	if err != nil {
		return err
	}
	if err := controlPlane.runEtcd(inquirer); err != nil {
		return err
	}
	kubeControllerManagerArguments := []string{
		"--kubeconfig", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "controller-manager.kubeconfig"),
		"--controllers=*,tokencleaner",
		"--service-account-private-key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "service-account.key"),
		"--cluster-signing-cert-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "cluster-signing-ca.crt"),
		"--cluster-signing-key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "cluster-signing-ca.key"),
		"--cluster-cidr", cluster.ClusterCIDR,
		"--service-cluster-ip-range", cluster.ServiceCIDR,
		"--allocate-node-cidrs", "true",
	}
	if cluster.NodeCIDRMaskSize > 0 {
		kubeControllerManagerArguments = append(
			kubeControllerManagerArguments,
			"--node-cidr-mask-size", strconv.Itoa(cluster.NodeCIDRMaskSize),
		)
	}
	if cluster.NodeCIDRMaskSizeIPv4 > 0 {
		kubeControllerManagerArguments = append(
			kubeControllerManagerArguments,
			"--node-cidr-mask-size-ipv4", strconv.Itoa(cluster.NodeCIDRMaskSizeIPv4),
		)
	}
	if cluster.NodeCIDRMaskSizeIPv6 > 0 {
		kubeControllerManagerArguments = append(
			kubeControllerManagerArguments,
			"--node-cidr-mask-size-ipv6", strconv.Itoa(cluster.NodeCIDRMaskSizeIPv6),
		)
	}
	_, err = hypervisor.EnsurePod(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		pod.NewPod(
			controlPlane.controlPlanePodName(inquirer),
			[]pod.Container{
				{
					Name:    "kube-apiserver",
					Image:   fmt.Sprintf(kubeAPIServerImage, kubernetesVersion),
					Command: []string{"kube-apiserver"},
					Args: []string{
						"--insecure-port", "0",
						"--advertise-address", advertiseAddressHost,
						"--secure-port", strconv.Itoa(advertiseAddressPort),
						"--etcd-servers", strings.Join(controlPlane.etcdClientEndpoints(inquirer), ","),
						"--etcd-cafile", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-ca.crt"),
						"--etcd-certfile", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.crt"),
						"--etcd-keyfile", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.key"),
						"--anonymous-auth", "false",
						"--authorization-mode", "Node,RBAC",
						"--enable-bootstrap-token-auth",
						"--allow-privileged", "true",
						"--tls-cert-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver.crt"),
						"--tls-private-key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver.key"),
						"--client-ca-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-client-ca.crt"),
						"--service-account-key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "service-account-pub.key"),
						"--kubelet-certificate-authority", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "kubelet-ca.crt"),
						"--kubelet-client-certificate", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-kubelet-client.crt"),
						"--kubelet-client-key", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-kubelet-client.key"),
						"--kubelet-preferred-address-types", "ExternalIP,ExternalDNS,InternalIP,InternalDNS,Hostname",
						"--service-cluster-ip-range", cluster.ServiceCIDR,
					},
					Mounts: map[string]string{
						componentSecretsPath(cluster.Namespace, cluster.Name, component.Name): componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
					},
				},
				{
					Name:    "kube-controller-manager",
					Image:   fmt.Sprintf(kubeControllerManagerImage, kubernetesVersion),
					Command: []string{"kube-controller-manager"},
					Args:    kubeControllerManagerArguments,
					Mounts: map[string]string{
						componentSecretsPath(cluster.Namespace, cluster.Name, component.Name): componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
					},
				},
				{
					Name:    "kube-scheduler",
					Image:   fmt.Sprintf(kubeSchedulerImage, kubernetesVersion),
					Command: []string{"kube-scheduler"},
					Args: []string{
						"--kubeconfig", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "scheduler.kubeconfig"),
					},
					Mounts: map[string]string{
						componentSecretsPath(cluster.Namespace, cluster.Name, component.Name): componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
					},
				},
			},
			map[int]int{
				apiserverHostPort: advertiseAddressPort,
			},
			pod.PrivilegesUnprivileged,
		),
	)
	return err
}

func (controlPlane *ControlPlane) kubeAPIServerSANs(inquirer inquirer.ReconcilerInquirer) ([]string, error) {
	cluster := inquirer.Cluster()
	clusterLoadBalancers := inquirer.ClusterComponents(componentapi.ControlPlaneIngressRole)
	if !clusterLoadBalancers.AllWithHypervisorAssigned() {
		return []string{}, errors.New("not all load balancers have an hypervisor assigned")
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
	kubernetesServiceIP, err := cluster.KubernetesServiceIP()
	if err != nil {
		return []string{}, err
	}
	kubeAPIServerExtraSANs = append(
		kubeAPIServerExtraSANs,
		kubernetesServiceIP,
	)
	return kubeAPIServerExtraSANs, nil
}

func (controlPlane *ControlPlane) uploadFiles(inquirer inquirer.ReconcilerInquirer) error {
	cluster := inquirer.Cluster()
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
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
	kubeAPIServerExtraSANs, err := controlPlane.kubeAPIServerSANs(inquirer)
	if err != nil {
		return err
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
	kubeletClientCertificate, err := component.ClientCertificate(
		cluster.CertificateAuthorities.KubeletClient,
		"kube-apiserver-kubelet-client",
		"kube-apiserver-kubelet-client",
		[]string{constants.OneInfraKubeletProxierExtraGroups},
		[]string{},
	)
	if err != nil {
		return err
	}
	_, advertiseAddressPort, err := controlPlane.kubeAPIServerAdvertiseAddressAndPort(inquirer)
	if err != nil {
		return err
	}
	apiserverURL := url.URL{Scheme: "https", Host: net.JoinHostPort("127.0.0.1", strconv.Itoa(advertiseAddressPort))}
	controllerManagerKubeConfig, err := component.KubeConfig(cluster, apiserverURL.String(), "controller-manager")
	if err != nil {
		return err
	}
	schedulerKubeConfig, err := component.KubeConfig(cluster, apiserverURL.String(), "scheduler")
	if err != nil {
		return err
	}
	return hypervisor.UploadFiles(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		map[string]string{
			// etcd secrets
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-ca.crt"):               cluster.EtcdServer.CA.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.crt"): etcdAPIServerClientCertificate.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.key"): etcdAPIServerClientCertificate.PrivateKey,
			// API server secrets
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-client-ca.crt"):      cluster.CertificateAuthorities.APIServerClient.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver.crt"):                apiServerCertificate.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver.key"):                apiServerCertificate.PrivateKey,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "service-account-pub.key"):      cluster.APIServer.ServiceAccount.PublicKey,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-kubelet-client.crt"): kubeletClientCertificate.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-kubelet-client.key"): kubeletClientCertificate.PrivateKey,
			// controller-manager secrets
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "controller-manager.kubeconfig"): controllerManagerKubeConfig,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "service-account.key"):           cluster.APIServer.ServiceAccount.PrivateKey,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "cluster-signing-ca.crt"):        cluster.CertificateAuthorities.CertificateSigner.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "cluster-signing-ca.key"):        cluster.CertificateAuthorities.CertificateSigner.PrivateKey,
			// scheduler secrets
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "scheduler.kubeconfig"): schedulerKubeConfig,
			// kubelet secrets
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "kubelet-ca.crt"): cluster.CertificateAuthorities.Kubelet.Certificate,
		},
	)
}

func (controlPlane *ControlPlane) kubeAPIServerAdvertiseAddressAndPort(inquirer inquirer.ReconcilerInquirer) (string, int, error) {
	advertiseAddressHost := ""
	advertiseAddressPort := 0
	clusterLoadBalancers := inquirer.ClusterComponents(componentapi.ControlPlaneIngressRole)
	if !clusterLoadBalancers.AllWithHypervisorAssigned() {
		return "", 0, errors.New("not all load balancers have an hypervisor assigned")
	}
	for _, clusterLoadBalancer := range clusterLoadBalancers {
		loadBalancerHypervisor := inquirer.ComponentHypervisor(clusterLoadBalancer)
		if loadBalancerHypervisor == nil {
			continue
		}
		advertiseAddressHost = loadBalancerHypervisor.IPAddress
		var err error
		advertiseAddressPort, err = clusterLoadBalancer.RequestPort(loadBalancerHypervisor, APIServerHostPortName)
		if err != nil {
			continue
		}
	}
	return advertiseAddressHost, advertiseAddressPort, nil
}

func (controlPlane *ControlPlane) controlPlanePodName(inquirer inquirer.ReconcilerInquirer) string {
	return fmt.Sprintf("control-plane-%s", inquirer.Cluster().Name)
}

func (controlPlane *ControlPlane) stopControlPlane(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	err := hypervisor.DeletePod(
		inquirer.Cluster().Namespace,
		inquirer.Cluster().Name,
		component.Name,
		controlPlane.controlPlanePodName(inquirer),
	)
	if err == nil {
		if err := component.FreePort(hypervisor, APIServerHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", APIServerHostPortName, hypervisor.Name)
		}
	}
	return err
}

// ReconcileDeletion reconciles the kube-apiserver deletion
func (controlPlane *ControlPlane) ReconcileDeletion(inquirer inquirer.ReconcilerInquirer) error {
	hypervisor := inquirer.Hypervisor()
	if hypervisor == nil {
		return nil
	}
	if err := controlPlane.stopControlPlane(inquirer); err != nil {
		return err
	}
	if inquirer.Cluster().DeletionTimestamp == nil {
		if err := controlPlane.removeEtcdMember(inquirer); err != nil {
			return err
		}
	}
	if err := controlPlane.stopEtcd(inquirer); err != nil {
		return err
	}
	return controlPlane.hostCleanup(inquirer)
}

func (controlPlane *ControlPlane) hostCleanup(inquirer inquirer.ReconcilerInquirer) error {
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
					Name:    "etcd-cleanup",
					Image:   infra.ToolingImage,
					Command: []string{"/bin/sh"},
					Args: []string{
						"-c",
						fmt.Sprintf(
							"rm -rf %s && ((rmdir %s && rmdir %s && rmdir %s) || true)",
							subcomponentStoragePath(cluster.Namespace, cluster.Name, component.Name, "etcd"),
							componentStoragePath(cluster.Namespace, cluster.Name, component.Name),
							clusterStoragePath(cluster.Namespace, cluster.Name),
							namespacedClusterStoragePath(cluster.Namespace),
						),
					},
					Mounts: map[string]string{
						globalStoragePath(): globalStoragePath(),
					},
				},
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
