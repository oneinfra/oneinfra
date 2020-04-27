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
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"github.com/oneinfra/oneinfra/internal/pkg/utils"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

const (
	etcdDialTimeout        = 5 * time.Second
	etcdImage              = "oneinfra/etcd:%s"
	etcdDataDir            = "/var/lib/etcd"
	etcdPeerHostPortName   = "etcd-peer"
	etcdClientHostPortName = "etcd-client"
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

func (controlPlane *ControlPlane) etcdClientEndpoints(inquirer inquirer.ReconcilerInquirer) []string {
	cluster := inquirer.Cluster()
	endpoints := []string{}
	for _, endpoint := range cluster.StorageClientEndpoints {
		endpointURL := strings.Split(endpoint, "=")
		endpoints = append(endpoints, endpointURL[1])
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
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	for {
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
			break
		}
		if goerrors.Is(err, context.DeadlineExceeded) {
			return errors.Errorf("failed to add etcd learner %q: %v", component.Name, err)
		}
		klog.V(2).Infof("failed to add etcd learner: %v", err)
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	klog.V(2).Infof("etcd learner %q added", component.Name)
	return nil
}

func (controlPlane *ControlPlane) hasEtcdLearner(inquirer inquirer.ReconcilerInquirer) (bool, error) {
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return false, err
	}
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
	for {
		klog.V(2).Infof("trying to promote etcd learner %q", component.Name)
		etcdClient, err := controlPlane.etcdClient(inquirer)
		if err != nil {
			return err
		}
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
		if _, err = etcdClient.MemberPromote(ctx, newMemberID); err == nil {
			klog.V(2).Infof("learner member %q successfully promoted", component.Name)
			break
		}
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	return nil
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

func (controlPlane *ControlPlane) hasEtcdMember(inquirer inquirer.ReconcilerInquirer) bool {
	if memberFound, _, err := controlPlane.etcdMemberID(inquirer); err == nil {
		return memberFound
	}
	// If we couldn't connect to etcd (our main source of truth,
	// fallback to our current knowledge of the system)
	return utils.HasListAnyElement(
		inquirer.Cluster().StoragePeerEndpoints,
		fmt.Sprintf("%s=%s", inquirer.Component().Name, controlPlane.etcdPeerEndpoint(inquirer)),
	)
}

// etcdMemberID returns whether this member was found and its ID
func (controlPlane *ControlPlane) etcdMemberID(inquirer inquirer.ReconcilerInquirer) (bool, uint64, error) {
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return false, 0, err
	}
	peerURL, err := controlPlane.etcdPeerURL(inquirer)
	if err != nil {
		return false, 0, err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	memberList, err := etcdClient.MemberList(ctx)
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
	etcdPeerHostPort, exists := component.AllocatedHostPorts[etcdPeerHostPortName]
	if !exists {
		return "", errors.Errorf("could not retrieve etcd peer host port for component %s", component.Name)
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
	cluster := inquirer.Cluster()
	if len(cluster.StoragePeerEndpoints) > 0 {
		if hasEtcdMember := controlPlane.hasEtcdMember(inquirer); !hasEtcdMember {
			if err := controlPlane.setupEtcdLearner(inquirer); err != nil {
				klog.Warningf("setting up etcd learner failed: %v", err)
			}
		}
	}
	etcdPeerHostPort, err := controlPlane.etcdPeerHostPort(inquirer)
	if err != nil {
		return err
	}
	cluster.StoragePeerEndpoints = utils.AddElementsToListIfNotExists(
		cluster.StoragePeerEndpoints,
		etcdEndpoint(inquirer, etcdPeerHostPort),
	)
	if err := controlPlane.ensureEtcdPod(inquirer); err != nil {
		return err
	}
	etcdClientHostPort, err := controlPlane.etcdClientHostPort(inquirer)
	if err != nil {
		return err
	}
	cluster.StorageClientEndpoints = utils.AddElementsToListIfNotExists(
		cluster.StorageClientEndpoints,
		etcdEndpoint(inquirer, etcdClientHostPort),
	)
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
		Args: []string{
			"--name", component.Name,
			"--client-cert-auth",
			"--peer-cert-allowed-cn", fmt.Sprintf("%s.etcd.cluster", cluster.Name),
			"--experimental-peer-skip-client-san-verification",
			"--cert-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.crt"),
			"--key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd.key"),
			"--trusted-ca-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-client-ca.crt"),
			"--peer-trusted-ca-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer-ca.crt"),
			"--peer-cert-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.crt"),
			"--peer-key-file", componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-peer.key"),
			"--data-dir", etcdDataDir,
			"--listen-client-urls", listenClientURLs.String(),
			"--advertise-client-urls", advertiseClientURLs.String(),
			"--listen-peer-urls", listenPeerURLs.String(),
			"--initial-advertise-peer-urls", initialAdvertisePeerURLs.String(),
			"--enable-grpc-gateway=false",
		},
		Env: map[string]string{
			"ETCDCTL_CACERT": componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "etcd-ca.crt"),
			"ETCDCTL_CERT":   componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.crt"),
			"ETCDCTL_KEY":    componentSecretsPathFile(cluster.Namespace, cluster.Name, component.Name, "apiserver-etcd-client.key"),
		},
		Mounts: map[string]string{
			componentSecretsPath(cluster.Namespace, cluster.Name, component.Name):            componentSecretsPath(cluster.Namespace, cluster.Name, component.Name),
			subcomponentStoragePath(cluster.Namespace, cluster.Name, component.Name, "etcd"): etcdDataDir,
		},
	}
	if len(cluster.StoragePeerEndpoints) == 1 {
		etcdContainer.Args = append(
			etcdContainer.Args,
			"--initial-cluster-state", "new",
		)
	} else {
		endpoints := []string{}
		for _, endpoint := range cluster.StoragePeerEndpoints {
			endpointURLRaw := strings.Split(endpoint, "=")
			endpointURL := url.URL{Scheme: "https", Host: endpointURLRaw[1]}
			endpoints = append(endpoints, fmt.Sprintf("%s=%s", endpointURLRaw[0], endpointURL.String()))
		}
		etcdContainer.Args = append(
			etcdContainer.Args,
			"--initial-cluster", strings.Join(endpoints, ","),
			"--initial-cluster-state", "existing",
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
	if len(controlPlane.etcdClientEndpoints(inquirer)) > 0 {
		memberFound, memberID, err := controlPlane.etcdMemberID(inquirer)
		if err != nil || !memberFound {
			return err
		}
		etcdClient, err := controlPlane.etcdClient(inquirer)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		if _, err = etcdClient.MemberRemove(ctx, memberID); err != nil {
			return errors.Wrap(err, "could not remove etcd member")
		}
	}
	component := inquirer.Component()
	cluster := inquirer.Cluster()
	hypervisor := inquirer.Hypervisor()
	if etcdPeerHostPort, exists := component.AllocatedHostPorts[etcdPeerHostPortName]; exists {
		cluster.StoragePeerEndpoints = utils.RemoveElementsFromList(
			cluster.StoragePeerEndpoints,
			etcdEndpoint(inquirer, etcdPeerHostPort),
		)
	}
	if etcdClientHostPort, exists := component.AllocatedHostPorts[etcdClientHostPortName]; exists {
		cluster.StorageClientEndpoints = utils.RemoveElementsFromList(
			cluster.StorageClientEndpoints,
			etcdEndpoint(inquirer, etcdClientHostPort),
		)
	}
	if err := component.FreePort(hypervisor, etcdPeerHostPortName); err != nil {
		return errors.Wrapf(err, "could not free port %q for hypervisor %q", etcdPeerHostPortName, hypervisor.Name)
	}
	if err := component.FreePort(hypervisor, etcdClientHostPortName); err != nil {
		return errors.Wrapf(err, "could not free port %q for hypervisor %q", etcdClientHostPortName, hypervisor.Name)
	}
	return nil
}

func (controlPlane *ControlPlane) etcdPeerHostPort(inquirer inquirer.ReconcilerInquirer) (int, error) {
	return inquirer.Component().RequestPort(inquirer.Hypervisor(), etcdPeerHostPortName)
}

func (controlPlane *ControlPlane) etcdClientHostPort(inquirer inquirer.ReconcilerInquirer) (int, error) {
	return inquirer.Component().RequestPort(inquirer.Hypervisor(), etcdClientHostPortName)
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
		if err := component.FreePort(hypervisor, etcdPeerHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", etcdPeerHostPortName, hypervisor.Name)
		}
		if err := component.FreePort(hypervisor, etcdClientHostPortName); err != nil {
			return errors.Wrapf(err, "could not free port %q for hypervisor %q", etcdClientHostPortName, hypervisor.Name)
		}
	}
	return err
}
