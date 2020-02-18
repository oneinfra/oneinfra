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
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"k8s.io/klog"

	"oneinfra.ereslibre.es/m/internal/pkg/infra/pod"
	"oneinfra.ereslibre.es/m/internal/pkg/inquirer"
)

const (
	etcdDialTimeout = 5 * time.Second
	etcdImage       = "oneinfra/etcd:3.4.3"
	etcdDataDir     = "/var/lib/etcd"
)

func (controlPlane *ControlPlane) etcdClient(inquirer inquirer.ReconcilerInquirer) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   controlPlane.etcdClientEndpoints(inquirer),
		DialTimeout: etcdDialTimeout,
	})
}

func (controlPlane *ControlPlane) etcdClientWithEndpoints(endpoints []string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: etcdDialTimeout,
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
	node := inquirer.Node()
	cluster := inquirer.Cluster()
	hypervisor := inquirer.Hypervisor()
	endpoints := []string{}
	for _, endpoint := range cluster.StoragePeerEndpoints {
		endpointURL := strings.Split(endpoint, "=")
		endpoints = append(endpoints, endpointURL[1])
	}
	if etcdPeerHostPort, ok := node.AllocatedHostPorts["etcd-peer"]; ok {
		endpoints = append(endpoints, fmt.Sprintf("http://%s:%d", hypervisor.IPAddress, etcdPeerHostPort))
	}
	return endpoints
}

func (controlPlane *ControlPlane) setupEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	node := inquirer.Node()
	hypervisor := inquirer.Hypervisor()
	etcdClient, err := controlPlane.etcdClient(inquirer)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	for {
		klog.V(2).Infof("adding etcd learner %s", node.Name)
		etcdPeerHostPort, ok := node.AllocatedHostPorts["etcd-peer"]
		if !ok {
			return errors.Errorf("etcd peer host port not found for node %s", node.Name)
		}
		_, err = etcdClient.MemberAddAsLearner(
			ctx,
			[]string{
				fmt.Sprintf("http://%s:%d", hypervisor.IPAddress, etcdPeerHostPort),
			},
		)
		if err == nil {
			break
		}
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	return nil
}

func (controlPlane *ControlPlane) promoteEtcdLearner(inquirer inquirer.ReconcilerInquirer) error {
	node := inquirer.Node()
	for {
		klog.V(2).Infof("trying to promote etcd learner %s", node.Name)
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
			klog.V(2).Infof("member %q not found", node.Name)
			continue
		}
		etcdClient, err = controlPlane.etcdClientWithEndpoints(endpoints)
		if err != nil {
			return err
		}
		_, err = etcdClient.MemberPromote(ctx, newMemberID)
		if err == nil {
			break
		}
		// TODO: retry timeout
		time.Sleep(time.Second)
	}
	return nil
}

func (controlPlane *ControlPlane) runEtcd(inquirer inquirer.ReconcilerInquirer) error {
	node := inquirer.Node()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	etcdPeerHostPort, err := hypervisor.RequestPort(cluster.Name, node.Name)
	if err != nil {
		return err
	}
	node.AllocatedHostPorts["etcd-peer"] = etcdPeerHostPort
	settingUpLearner := len(cluster.StoragePeerEndpoints) > 0
	if settingUpLearner {
		if err := controlPlane.setupEtcdLearner(inquirer); err != nil {
			return err
		}
	}
	etcdClientHostPort, err := hypervisor.RequestPort(cluster.Name, node.Name)
	if err != nil {
		return err
	}
	node.AllocatedHostPorts["etcd-client"] = etcdClientHostPort
	cluster.StoragePeerEndpoints = append(
		cluster.StoragePeerEndpoints,
		fmt.Sprintf("%s=%s:%d", node.Name, hypervisor.IPAddress, etcdPeerHostPort),
	)
	etcdContainer, err := controlPlane.etcdContainer(inquirer, etcdClientHostPort, etcdPeerHostPort)
	if err != nil {
		return err
	}
	_, err = hypervisor.RunPod(
		cluster,
		pod.NewPod(
			fmt.Sprintf("etcd-%s", cluster.Name),
			[]pod.Container{
				etcdContainer,
			},
			map[int]int{
				etcdClientHostPort: 2379,
				etcdPeerHostPort:   2380,
			},
		),
	)
	if err != nil {
		return err
	}
	if settingUpLearner {
		if err := controlPlane.promoteEtcdLearner(inquirer); err != nil {
			return err
		}
		klog.V(2).Infof("etcd learner %s successfully promoted", node.Name)
	}
	cluster.StorageClientEndpoints = append(
		cluster.StorageClientEndpoints,
		fmt.Sprintf("%s=%s:%d", node.Name, hypervisor.IPAddress, etcdClientHostPort),
	)
	return nil
}

func (controlPlane *ControlPlane) etcdContainer(inquirer inquirer.ReconcilerInquirer, etcdClientHostPort, etcdPeerHostPort int) (pod.Container, error) {
	node := inquirer.Node()
	hypervisor := inquirer.Hypervisor()
	cluster := inquirer.Cluster()
	etcdContainer := pod.Container{
		Name:    "etcd",
		Image:   etcdImage,
		Command: []string{"etcd"},
		Args: []string{
			"--name", node.Name,
			"--data-dir", etcdDataDir,
			"--listen-client-urls", "http://0.0.0.0:2379",
			"--advertise-client-urls", fmt.Sprintf("http://%s:%d", hypervisor.IPAddress, etcdClientHostPort),
			"--listen-peer-urls", "http://0.0.0.0:2380",
			"--initial-advertise-peer-urls", fmt.Sprintf("http://%s:%d", hypervisor.IPAddress, etcdPeerHostPort),
		},
		Mounts: map[string]string{
			filepath.Join(storagePath(cluster.Name), "etcd", node.Name): etcdDataDir,
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
			endpointURL := strings.Split(endpoint, "=")
			endpoints = append(endpoints, fmt.Sprintf("%s=http://%s", endpointURL[0], endpointURL[1]))
		}
		etcdContainer.Args = append(
			etcdContainer.Args,
			"--initial-cluster", strings.Join(endpoints, ","),
			"--initial-cluster-state", "existing",
		)
	}
	return etcdContainer, nil
}
