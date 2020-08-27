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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	goerrors "errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/component"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
)

const (
	// EtcdPeerHostPortName represents the etcd peer host port
	// allocation name
	EtcdPeerHostPortName = "etcd-peer"

	// EtcdClientHostPortName represents the etcd client host port
	// allocation name
	EtcdClientHostPortName = "etcd-client"
)

const (
	etcdDialTimeout = 5 * time.Second
	etcdImage       = "oneinfra/etcd:%s"
	etcdDataDir     = "/var/lib/etcd"
)

var (
	logOutput = []string{"/dev/null"}
)

func (controlPlane *ControlPlane) etcdClient(inquirer inquirer.ReconcilerInquirer) (*clientv3.Client, error) {
	return controlPlane.etcdClientWithEndpoints(
		inquirer,
		controlPlane.etcdClientEndpoints(inquirer),
	)
}

func (controlPlane *ControlPlane) etcdClientWithEndpoints(inquirer inquirer.ReconcilerInquirer, endpoints []string) (*clientv3.Client, error) {
	component := inquirer.Component()
	cluster := inquirer.Cluster()
	if cluster.HasUninitializedCertificates() {
		return nil, errors.Errorf("cluster has some uninitialized certificates")
	}
	etcdServerCABlock, _ := pem.Decode([]byte(cluster.EtcdServer.CA.Certificate))
	if etcdServerCABlock == nil {
		return nil, errors.Errorf("cannot decode etcd server CA certificate")
	}
	etcdServerCA, err := x509.ParseCertificate(etcdServerCABlock.Bytes)
	if err != nil {
		return nil, err
	}
	etcdClientCert, err := component.ClientCertificate(
		cluster.CertificateAuthorities.EtcdClient,
		"oneinfra-client",
		"oneinfra-client",
		[]string{cluster.Name},
		[]string{},
	)
	if err != nil {
		return nil, err
	}
	etcdClient, err := tls.X509KeyPair([]byte(etcdClientCert.Certificate), []byte(etcdClientCert.PrivateKey))
	if err != nil {
		return nil, err
	}
	etcdServerCAPool := x509.NewCertPool()
	etcdServerCAPool.AddCert(etcdServerCA)
	logConfig := zap.NewProductionConfig()
	logConfig.OutputPaths = logOutput
	logConfig.ErrorOutputPaths = logOutput
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: etcdDialTimeout,
		TLS: &tls.Config{
			RootCAs:      etcdServerCAPool,
			Certificates: []tls.Certificate{etcdClient},
		},
		LogConfig: &logConfig,
	})
}

func (controlPlane *ControlPlane) etcdPeerEndpoints(inquirer inquirer.ReconcilerInquirer) []string {
	endpoints := []string{}
	controlPlaneComponents := inquirer.ClusterComponents(component.ControlPlaneRole)
	for _, controlPlaneComponent := range controlPlaneComponents {
		componentHypervisor := inquirer.ComponentHypervisor(controlPlaneComponent)
		if componentHypervisor == nil {
			continue
		}
		etcdPeerHostPort, err := controlPlaneComponent.RequestPort(componentHypervisor, EtcdPeerHostPortName)
		if err != nil {
			continue
		}
		url := url.URL{Scheme: "https", Host: net.JoinHostPort(componentHypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
		endpoints = append(endpoints, url.String())
	}
	return endpoints
}

func (controlPlane *ControlPlane) etcdClientEndpoints(inquirer inquirer.ReconcilerInquirer) []string {
	endpoints := []string{}
	controlPlaneComponents := inquirer.ClusterComponents(component.ControlPlaneRole)
	for _, controlPlaneComponent := range controlPlaneComponents {
		if controlPlaneComponent.DeletionTimestamp != nil {
			continue
		}
		componentHypervisor := inquirer.ComponentHypervisor(controlPlaneComponent)
		if componentHypervisor == nil {
			continue
		}
		etcdClientHostPort, err := controlPlaneComponent.RequestPort(componentHypervisor, EtcdClientHostPortName)
		if err != nil {
			continue
		}
		url := url.URL{Scheme: "https", Host: net.JoinHostPort(componentHypervisor.IPAddress, strconv.Itoa(etcdClientHostPort))}
		endpoints = append(endpoints, url.String())
	}
	return endpoints
}

func (controlPlane *ControlPlane) etcdPeerEndpoint(inquirer inquirer.ReconcilerInquirer) string {
	hypervisor := inquirer.Hypervisor()
	etcdPeerHostPort, err := controlPlane.etcdPeerHostPort(inquirer)
	if err != nil {
		return ""
	}
	return net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))
}

func (controlPlane *ControlPlane) setupEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return err
	}
	defer etcdClient.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	for i := 0; i < 60; i++ {
		klog.V(2).Infof("adding etcd learner %q", component.Name)
		peerURL, err := controlPlane.etcdPeerURL(inquirer)
		if err != nil {
			klog.V(2).Infof("failed to retrieve etcd peer URL: %v", err)
			return err
		}
		_, err = etcdClient.MemberAddAsLearner(
			ctx,
			[]string{peerURL},
		)
		if err == nil {
			klog.V(2).Infof("etcd learner %q added", component.Name)
			return nil
		}
		if goerrors.Is(err, context.DeadlineExceeded) {
			return errors.Errorf("failed to add etcd learner %q: %v", component.Name, err)
		}
		klog.V(2).Infof("failed to add etcd learner: %v", err)
		time.Sleep(time.Second)
	}
	return errors.Errorf("failed to add etcd learner %q", component.Name)
}

func (controlPlane *ControlPlane) hasEtcdLearner(inquirer inquirer.ReconcilerInquirer) (bool, error) {
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return false, err
	}
	defer etcdClient.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	etcdMembers, err := etcdClient.MemberList(ctx)
	if err != nil {
		return false, err
	}
	peerURL, err := controlPlane.etcdPeerURL(inquirer)
	if err != nil {
		return false, err
	}
	for _, etcdMember := range etcdMembers.Members {
		if reflect.DeepEqual([]string{peerURL}, etcdMember.PeerURLs) {
			return etcdMember.IsLearner, nil
		}
	}
	return false, nil
}

func (controlPlane *ControlPlane) promoteEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	for i := 0; i < 300; i++ {
		klog.V(2).Infof("trying to promote etcd learner %q", component.Name)
		etcdClient, err := controlPlane.etcdClient(inquirer)
		if err != nil {
			return err
		}
		defer etcdClient.Close()
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		etcdMembers, err := etcdClient.MemberList(ctx)
		if err != nil {
			continue
		}
		peerURL, err := controlPlane.etcdPeerURL(inquirer)
		if err != nil {
			return err
		}
		var newMemberID uint64
		memberFound := false
		endpoints := []string{}
		for _, etcdMember := range etcdMembers.Members {
			if reflect.DeepEqual([]string{peerURL}, etcdMember.PeerURLs) {
				if !etcdMember.IsLearner {
					return nil
				}
				memberFound = true
				newMemberID = etcdMember.ID
			} else if len(etcdMember.ClientURLs) > 0 {
				endpoints = append(endpoints, etcdMember.ClientURLs...)
			}
		}
		if !memberFound {
			klog.V(2).Infof("learner member %q not found", component.Name)
			continue
		}
		etcdClient, err = controlPlane.etcdClientWithEndpoints(inquirer, endpoints)
		if err != nil {
			return err
		}
		defer etcdClient.Close()
		if _, err = etcdClient.MemberPromote(ctx, newMemberID); err == nil {
			klog.V(2).Infof("learner member %q successfully promoted", component.Name)
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.Errorf("failed to promote etcd learner member %q", component.Name)
}

func (controlPlane *ControlPlane) etcdPod(inquirer inquirer.ReconcilerInquirer) (pod.Pod, error) {
	component := inquirer.Component()
	etcdPeerHostPort, err := controlPlane.etcdPeerHostPort(inquirer)
	if err != nil {
		return pod.Pod{}, errors.Wrapf(err, "could not allocate etcd peer host port for component %q", component.Name)
	}
	etcdClientHostPort, err := controlPlane.etcdClientHostPort(inquirer)
	if err != nil {
		return pod.Pod{}, errors.Wrapf(err, "could not allocate etcd client host port for component %q", component.Name)
	}
	etcdContainer, err := controlPlane.etcdContainer(inquirer, etcdClientHostPort, etcdPeerHostPort)
	if err != nil {
		return pod.Pod{}, err
	}
	return pod.NewPod(
		controlPlane.etcdPodName(inquirer),
		[]pod.Container{
			etcdContainer,
		},
		map[int]int{
			etcdClientHostPort: 2379,
			etcdPeerHostPort:   2380,
		},
		pod.PrivilegesUnprivileged,
	), nil
}

func (controlPlane *ControlPlane) hasEtcdMember(inquirer inquirer.ReconcilerInquirer) (bool, error) {
	memberFound, _, err := controlPlane.etcdMemberID(inquirer)
	return memberFound, err
}

func (controlPlane *ControlPlane) etcdMembers(inquirer inquirer.ReconcilerInquirer) (*clientv3.MemberListResponse, error) {
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return nil, err
	}
	defer etcdClient.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	return etcdClient.MemberList(ctx)
}

// etcdMemberID returns whether this member was found and its ID
func (controlPlane *ControlPlane) etcdMemberID(inquirer inquirer.ReconcilerInquirer) (bool, uint64, error) {
	memberList, err := controlPlane.etcdMembers(inquirer)
	if err != nil {
		return false, 0, err
	}
	peerURL, err := controlPlane.etcdPeerURL(inquirer)
	if err != nil {
		return false, 0, err
	}
	for _, member := range memberList.Members {
		if reflect.DeepEqual([]string{peerURL}, member.PeerURLs) {
			return true, member.ID, nil
		}
	}
	return false, 0, nil
}

func (controlPlane *ControlPlane) etcdPeerURL(inquirer inquirer.ReconcilerInquirer) (string, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	etcdPeerHostPort, err := component.RequestPort(hypervisor, EtcdPeerHostPortName)
	if err != nil {
		return "", err
	}
	peerURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
	return peerURL.String(), nil
}

func (controlPlane *ControlPlane) reconcileEtcdCertificatesAndKeys(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	etcdPeerCertificate, err := component.ClientCertificate(
		cluster.CertificateAuthorities.EtcdPeer,
		"etcd-peer",
		fmt.Sprintf("%s.etcd.cluster", cluster.Name),
		[]string{cluster.Name},
		// Peer authentication via SANs
		[]string{hypervisor.IPAddress},
	)
	if err != nil {
		return err
	}
	etcdServerCertificate, err := component.ServerCertificate(
		cluster.EtcdServer.CA,
		"etcd",
		"etcd",
		[]string{"etcd"},
		[]string{hypervisor.IPAddress},
	)
	if err != nil {
		return err
	}
	return hypervisor.UploadFiles(
		cluster.Namespace,
		cluster.Name,
		component.Name,
		map[string]string{
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.crt"):           etcdServerCertificate.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.key"):           etcdServerCertificate.PrivateKey,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-client-ca.crt"): cluster.CertificateAuthorities.EtcdClient.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer-ca.crt"):   cluster.CertificateAuthorities.EtcdPeer.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.crt"):      etcdPeerCertificate.Certificate,
			componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.key"):      etcdPeerCertificate.PrivateKey,
		},
	)
}

func (controlPlane *ControlPlane) runEtcd(inquirer inquirer.ReconcilerInquirer) error {
	if err := controlPlane.reconcileEtcdCertificatesAndKeys(inquirer); err != nil {
		return err
	}
	hasEtcdMember, err := controlPlane.hasEtcdMember(inquirer)
	if err == nil && !hasEtcdMember {
		if err := controlPlane.setupEtcdLearner(inquirer); err != nil {
			klog.Warningf("setting up etcd learner failed: %v", err)
		}
	}
	if err := controlPlane.ensureEtcdPod(inquirer); err != nil {
		return err
	}
	hasEtcdLearner, err := controlPlane.hasEtcdLearner(inquirer)
	if err != nil {
		return err
	}
	if hasEtcdLearner {
		return controlPlane.promoteEtcdLearner(inquirer)
	}
	return nil
}

func (controlPlane *ControlPlane) ensureEtcdPod(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	cluster := inquirer.Cluster()
	hypervisor := inquirer.Hypervisor()
	etcdPod, err := controlPlane.etcdPod(inquirer)
	if err != nil {
		return err
	}
	if _, err = hypervisor.EnsurePod(cluster.Namespace, cluster.Name, component.Name, etcdPod); err != nil {
		return err
	}
	return nil
}

func (controlPlane *ControlPlane) etcdContainer(inquirer inquirer.ReconcilerInquirer, etcdClientHostPort, etcdPeerHostPort int) (pod.Container, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	kubernetesVersion := inquirer.Cluster().KubernetesVersion
	versionBundle, err := constants.KubernetesVersionBundle(kubernetesVersion)
	if err != nil {
		return pod.Container{}, errors.Errorf("could not retrieve version bundle for version %q", kubernetesVersion)
	}
	listenClientURLs := url.URL{Scheme: "https", Host: "0.0.0.0:2379"}
	advertiseClientURLs := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdClientHostPort))}
	listenPeerURLs := url.URL{Scheme: "https", Host: "0.0.0.0:2380"}
	initialAdvertisePeerURLs := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
	etcdContainer := pod.Container{
		Name:    "etcd",
		Image:   fmt.Sprintf(etcdImage, versionBundle.EtcdVersion),
		Command: []string{"etcd"},
		Args: component.ArgsFromMap(map[string]string{
			"name":                 component.Name,
			"client-cert-auth":     "true",
			"peer-cert-allowed-cn": fmt.Sprintf("%s.etcd.cluster", cluster.Name),
			"experimental-peer-skip-client-san-verification": "true",
			"cert-file":                   componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.crt"),
			"key-file":                    componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.key"),
			"trusted-ca-file":             componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-client-ca.crt"),
			"peer-trusted-ca-file":        componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer-ca.crt"),
			"peer-cert-file":              componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.crt"),
			"peer-key-file":               componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.key"),
			"data-dir":                    etcdDataDir,
			"listen-client-urls":          listenClientURLs.String(),
			"advertise-client-urls":       advertiseClientURLs.String(),
			"listen-peer-urls":            listenPeerURLs.String(),
			"initial-advertise-peer-urls": initialAdvertisePeerURLs.String(),
			"enable-grpc-gateway":         "false",
		}),
		Env: map[string]string{
			"ETCDCTL_CACERT":    componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-ca.crt"),
			"ETCDCTL_CERT":      componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.crt"),
			"ETCDCTL_KEY":       componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.key"),
			"ETCDCTL_ENDPOINTS": strings.Join(controlPlane.etcdClientEndpoints(inquirer), ","),
		},
		Mounts: map[string]string{
			componentSecretsPath(cluster.Namespace, cluster.Name, component.Name):            componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
			subcomponentStoragePath(cluster.Namespace, cluster.Name, component.Name, "etcd"): etcdDataDir,
		},
	}
	etcdMembers, err := controlPlane.etcdMembers(inquirer)
	if err != nil && len(cluster.StoragePeerEndpoints) == 0 {
		etcdContainer.Args = append(
			etcdContainer.Args,
			"--initial-cluster-state=new",
		)
	} else if err == nil {
		endpoints := []string{}
		peerURL, err := controlPlane.etcdPeerURL(inquirer)
		if err != nil {
			return pod.Container{}, err
		}
		for _, etcdMember := range etcdMembers.Members {
			memberName := etcdMember.Name
			if reflect.DeepEqual([]string{peerURL}, etcdMember.PeerURLs) {
				memberName = component.Name
			}
			endpoints = append(
				endpoints,
				fmt.Sprintf("%s=%s", memberName, etcdMember.PeerURLs[0]),
			)
		}
		etcdContainer.Args = append(
			etcdContainer.Args,
			component.ArgsFromMap(map[string]string{
				"initial-cluster":       strings.Join(endpoints, ","),
				"initial-cluster-state": "existing",
			})...,
		)
	}
	return etcdContainer, nil
}

func etcdEndpoint(inquirer inquirer.ReconcilerInquirer, hostPort int) string {
	return fmt.Sprintf(
		"%s=%s",
		inquirer.Component().Name,
		net.JoinHostPort(inquirer.Hypervisor().IPAddress, strconv.Itoa(hostPort)),
	)
}

func (controlPlane *ControlPlane) removeEtcdMember(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	if len(controlPlane.etcdClientEndpoints(inquirer)) > 0 {
		for i := 0; i < 60; i++ {
			klog.V(2).Infof("removing etcd member %s/%s", component.Namespace, component.Name)
			memberFound, memberID, err := controlPlane.etcdMemberID(inquirer)
			if err == nil && !memberFound {
				return nil
			}
			if err != nil {
				klog.Errorf("error retrieving etcd member ID for %s/%s: %v", component.Namespace, component.Name, err)
				time.Sleep(time.Second)
				continue
			}
			etcdClient, err := controlPlane.etcdClient(inquirer)
			if err != nil {
				klog.Errorf("could not create an etcd client to remove member %s/%s", component.Namespace, component.Name)
				time.Sleep(time.Second)
				continue
			}
			defer etcdClient.Close()
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			if _, err = etcdClient.MemberRemove(ctx, memberID); err != nil {
				klog.Errorf("failed to remove member %s/%s: %v", component.Namespace, component.Name, err)
				time.Sleep(time.Second)
				continue
			}
			klog.Infof("etcd member %s/%s correctly removed", component.Namespace, component.Name)
			return nil
		}
		return errors.Errorf("failed to remove etcd member %s/%s", component.Namespace, component.Name)
	}
	return nil
}

func (controlPlane *ControlPlane) etcdPeerHostPort(inquirer inquirer.ReconcilerInquirer) (int, error) {
	return inquirer.Component().RequestPort(inquirer.Hypervisor(), EtcdPeerHostPortName)
}

func (controlPlane *ControlPlane) etcdClientHostPort(inquirer inquirer.ReconcilerInquirer) (int, error) {
	return inquirer.Component().RequestPort(inquirer.Hypervisor(), EtcdClientHostPortName)
}

func (controlPlane *ControlPlane) etcdPodName(inquirer inquirer.ReconcilerInquirer) string {
	return fmt.Sprintf("etcd-%s", inquirer.Cluster().Name)
}

func (controlPlane *ControlPlane) stopEtcd(inquirer inquirer.ReconcilerInquirer) error {
	err := inquirer.Hypervisor().DeletePod(
		inquirer.Cluster().Namespace,
		inquirer.Cluster().Name,
		inquirer.Component().Name,
		controlPlane.etcdPodName(inquirer),
	)
	if err == nil {
		component := inquirer.Component()
		hypervisor := inquirer.Hypervisor()
		if err := component.FreePort(hypervisor, EtcdPeerHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", EtcdPeerHostPortName, hypervisor.Name)
		}
		if err := component.FreePort(hypervisor, EtcdClientHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", EtcdClientHostPortName, hypervisor.Name)
		}
	}
	return err
}
