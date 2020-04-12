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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
	"github.com/oneinfra/oneinfra/internal/pkg/utils"
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
	etcdClientCert, etcdClientPrivateKey, err := cluster.CertificateAuthorities.EtcdClient.CreateCertificate(
		"oneinfra-client",
		[]string{cluster.Name},
		[]string{},
	)
	if err != nil {
		return nil, err
	}
	etcdClient, err := tls.X509KeyPair([]byte(etcdClientCert), []byte(etcdClientPrivateKey))
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

func (controlPlane *ControlPlane) etcdPeerEndpoints(inquirer inquirer.ReconcilerInquirer) []string {
	component := inquirer.Component()
	cluster := inquirer.Cluster()
	hypervisor := inquirer.Hypervisor()
	endpoints := []string{}
	for _, endpoint := range cluster.StoragePeerEndpoints {
		endpointURL := strings.Split(endpoint, "=")
		endpoints = append(endpoints, endpointURL[1])
	}
	if etcdPeerHostPort, exists := component.AllocatedHostPorts[etcdPeerHostPortName]; exists {
		endpointURL := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
		endpoints = append(endpoints, endpointURL.String())
	}
	return endpoints
}

func (controlPlane *ControlPlane) setupEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	for {
		klog.V(2).Infof("adding etcd learner %q", component.Name)
		etcdPeerHostPort, exists := component.AllocatedHostPorts[etcdPeerHostPortName]
		if !exists {
			return errors.Errorf("etcd peer host port not found for component %q", component.Name)
		}
		peerURLs := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
		_, err = etcdClient.MemberAddAsLearner(
			ctx,
			[]string{peerURLs.String()},
		)
		if err == nil {
			break
		}
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	klog.V(2).Infof("etcd learner %q added", component.Name)
	return nil
}

func (controlPlane *ControlPlane) promoteEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	for {
		klog.V(2).Infof("trying to promote etcd learner %s", component.Name)
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
		var newMemberID uint64
		memberFound := false
		endpoints := []string{}
		for _, etcdMember := range etcdMembers.Members {
			if etcdMember.IsLearner {
				// We are creating only one learner at a time, so it's safe to
				// assume that if it's a learner it's the new member
				// (otherwise we have to compare peer URL's)
				memberFound = true
				newMemberID = etcdMember.ID
			} else if len(etcdMember.ClientURLs) > 0 {
				endpoints = append(endpoints, etcdMember.ClientURLs...)
			}
		}
		if !memberFound {
			klog.V(2).Infof("member %q not found", component.Name)
			continue
		}
		etcdClient, err = controlPlane.etcdClientWithEndpoints(inquirer, endpoints)
		if err != nil {
			return err
		}
		if _, err = etcdClient.MemberPromote(ctx, newMemberID); err == nil {
			break
		}
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	return nil
}

func (controlPlane *ControlPlane) etcdPod(inquirer inquirer.ReconcilerInquirer) (pod.Pod, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	etcdPeerHostPort, err := component.RequestPort(hypervisor, etcdPeerHostPortName)
	if err != nil {
		return pod.Pod{}, err
	}
	etcdClientHostPort, err := component.RequestPort(hypervisor, etcdClientHostPortName)
	if err != nil {
		return pod.Pod{}, err
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

// etcdMemberID returns whether this member was found and its ID
func (controlPlane *ControlPlane) etcdMemberID(inquirer inquirer.ReconcilerInquirer) (bool, uint64, error) {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return false, 0, err
	}
	etcdPeerHostPort, exists := component.AllocatedHostPorts[etcdPeerHostPortName]
	if !exists {
		return false, 0, errors.Errorf("etcd peer host port not found for component %s", component.Name)
	}
	peerURLs := url.URL{Scheme: "https", Host: net.JoinHostPort(hypervisor.IPAddress, strconv.Itoa(etcdPeerHostPort))}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	memberList, err := etcdClient.MemberList(ctx)
	if err != nil {
		return false, 0, err
	}
	for _, member := range memberList.Members {
		if reflect.DeepEqual([]string{peerURLs.String()}, member.PeerURLs) {
			return true, member.ID, nil
		}
	}
	return false, 0, nil
}

func (controlPlane *ControlPlane) reconcileCertificatesAndKeys(inquirer inquirer.ReconcilerInquirer) error {
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
		cluster.Name,
		component.Name,
		map[string]string{
			secretsPathFile(cluster.Name, component.Name, "etcd.crt"):           etcdServerCertificate.Certificate,
			secretsPathFile(cluster.Name, component.Name, "etcd.key"):           etcdServerCertificate.PrivateKey,
			secretsPathFile(cluster.Name, component.Name, "etcd-client-ca.crt"): cluster.CertificateAuthorities.EtcdClient.Certificate,
			secretsPathFile(cluster.Name, component.Name, "etcd-peer-ca.crt"):   cluster.CertificateAuthorities.EtcdPeer.Certificate,
			secretsPathFile(cluster.Name, component.Name, "etcd-peer.crt"):      etcdPeerCertificate.Certificate,
			secretsPathFile(cluster.Name, component.Name, "etcd-peer.key"):      etcdPeerCertificate.PrivateKey,
		},
	)
}

func (controlPlane *ControlPlane) runEtcdLearnerAndPromote(inquirer inquirer.ReconcilerInquirer, etcdPeerHostPort int) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	settingUpLearner := len(cluster.StoragePeerEndpoints) > 0
	if settingUpLearner {
		if err := controlPlane.setupEtcdLearner(inquirer); err != nil {
			return err
		}
	}
	etcdClientHostPort, err := component.RequestPort(hypervisor, etcdClientHostPortName)
	if err != nil {
		return err
	}
	cluster.StoragePeerEndpoints = append(
		cluster.StoragePeerEndpoints,
		etcdEndpoint(inquirer, etcdPeerHostPort),
	)
	etcdPod, err := controlPlane.etcdPod(inquirer)
	if err != nil {
		return err
	}
	if _, err = hypervisor.RunPod(cluster.Name, component.Name, etcdPod); err != nil {
		return err
	}
	if settingUpLearner {
		if err := controlPlane.promoteEtcdLearner(inquirer); err != nil {
			return err
		}
		klog.V(2).Infof("etcd learner %s successfully promoted", component.Name)
	}
	cluster.StorageClientEndpoints = append(
		cluster.StorageClientEndpoints,
		etcdEndpoint(inquirer, etcdClientHostPort),
	)
	return nil
}

func (controlPlane *ControlPlane) runEtcd(inquirer inquirer.ReconcilerInquirer) error {
	component := inquirer.Component()
	hypervisor := inquirer.Hypervisor()
	etcdPod, err := controlPlane.etcdPod(inquirer)
	if err != nil {
		return err
	}
	isEtcdRunning, _, err := hypervisor.IsPodRunning(etcdPod)
	if err != nil {
		return err
	}
	if isEtcdRunning {
		return nil
	}
	if err := controlPlane.reconcileCertificatesAndKeys(inquirer); err != nil {
		return err
	}
	etcdPeerHostPort, err := component.RequestPort(hypervisor, etcdPeerHostPortName)
	if err != nil {
		return err
	}
	hasEtcdMember := false
	if len(controlPlane.etcdClientEndpoints(inquirer)) > 0 {
		hasEtcdMember, err = controlPlane.hasEtcdMember(inquirer)
		if err != nil {
			return err
		}
	}
	if hasEtcdMember {
		return nil
	}
	return controlPlane.runEtcdLearnerAndPromote(inquirer, etcdPeerHostPort)
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
			"--cert-file", secretsPathFile(cluster.Name, component.Name, "etcd.crt"),
			"--key-file", secretsPathFile(cluster.Name, component.Name, "etcd.key"),
			"--trusted-ca-file", secretsPathFile(cluster.Name, component.Name, "etcd-client-ca.crt"),
			"--peer-trusted-ca-file", secretsPathFile(cluster.Name, component.Name, "etcd-peer-ca.crt"),
			"--peer-cert-file", secretsPathFile(cluster.Name, component.Name, "etcd-peer.crt"),
			"--peer-key-file", secretsPathFile(cluster.Name, component.Name, "etcd-peer.key"),
			"--data-dir", etcdDataDir,
			"--listen-client-urls", listenClientURLs.String(),
			"--advertise-client-urls", advertiseClientURLs.String(),
			"--listen-peer-urls", listenPeerURLs.String(),
			"--initial-advertise-peer-urls", initialAdvertisePeerURLs.String(),
			"--enable-grpc-gateway=false",
		},
		Mounts: map[string]string{
			secretsPath(cluster.Name, component.Name):                        secretsPath(cluster.Name, component.Name),
			filepath.Join(storagePath(cluster.Name), "etcd", component.Name): etcdDataDir,
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

func (controlPlane *ControlPlane) etcdPodName(inquirer inquirer.ReconcilerInquirer) string {
	return fmt.Sprintf("etcd-%s", inquirer.Cluster().Name)
}

func (controlPlane *ControlPlane) stopEtcd(inquirer inquirer.ReconcilerInquirer) error {
	return inquirer.Hypervisor().DeletePod(
		inquirer.Cluster().Name,
		inquirer.Component().Name,
		controlPlane.etcdPodName(inquirer),
	)
}
