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

package localcluster

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
)

const (
	haProxyImage            = "oneinfra/haproxy:latest"
	haProxyPort             = 3000
	haProxyCertBundlePath   = "/etc/oneinfra/hypervisor.crt"
	haProxyClientCACertPath = "/etc/oneinfra/hypervisor-client-ca.crt"
	haProxyConfigPath       = "/etc/oneinfra/haproxy.cfg"
	haProxyTemplate         = `global
  chroot /var/lib/haproxy
  daemon
defaults
  log global
  mode tcp
  timeout connect 10s
  timeout client  60s
  timeout server  60s
frontend cri_frontend
  bind *:{{ .HAProxyPort }} ssl crt {{ .HAProxyCertBundlePath }} ca-file {{ .HAProxyClientCACertPath }} verify required
  default_backend cri_backend
backend cri_backend
  server cri unix@containerd.sock
`
)

// Hypervisor represents a local hypervisor
type Hypervisor struct {
	Name                 string
	Public               bool
	HypervisorCluster    *HypervisorCluster
	CRIEndpoint          string
	ExposedPortRangeLow  int
	ExposedPortRangeHigh int
	CACertificate        *certificates.Certificate
	ClientCACertificate  *certificates.Certificate
}

// Create creates the local hypervisor
func (hypervisor *Hypervisor) Create() error {
	if err := hypervisor.createRuntimeDirectory(); err != nil {
		return err
	}
	if hypervisor.HypervisorCluster.Remote {
		if err := hypervisor.createCACertificates(); err != nil {
			return err
		}
	}
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	if err := exec.Command("docker", "inspect", hypervisor.HypervisorCluster.NodeImage).Run(); err != nil {
		klog.Infof("pulling image %q; this can take a while, please wait...\n", hypervisor.HypervisorCluster.NodeImage)
		if err := exec.Command("docker", "pull", hypervisor.HypervisorCluster.NodeImage).Run(); err != nil {
			return err
		}
	}
	klog.Infof("running fake hypervisor with name %q\n", hypervisor.fullName())
	return exec.Command("docker", []string{
		"run", "-d", "--privileged",
		"--name", hypervisor.fullName(),
		"-v", fmt.Sprintf("%s:%s", hypervisor.runtimeDirectory(), hypervisor.localContainerdSockDirectory()),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_UID=%s", currentUser.Uid),
		"-e", fmt.Sprintf("CONTAINERD_SOCK_GID=%s", currentUser.Gid),
		"-e", fmt.Sprintf("CONTAINER_RUNTIME_ENDPOINT=%s", hypervisor.localContainerdSockPath()),
		"-e", fmt.Sprintf("IMAGE_SERVICE_ENDPOINT=%s", hypervisor.localContainerdSockPath()),
		hypervisor.HypervisorCluster.NodeImage,
	}...).Run()
}

// StartRemoteCRIEndpoint initializes the remote CRI endpoint on this
// hypervisor, using the local endpoint (UNIX socket) in order to set
// up the required components and perform the required configuration
func (hypervisor *Hypervisor) StartRemoteCRIEndpoint() error {
	haProxyCfg, err := hypervisor.haProxyTemplate()
	if err != nil {
		return err
	}
	infraHypervisor := infra.NewLocalHypervisor(
		hypervisor.Name,
		hypervisor.containerdSockPath(),
	)
	hypervisorIPAddress, err := hypervisor.internalIPAddress()
	if err != nil {
		return err
	}
	criEndpointCertificate, criEndpointPrivateKey, err := hypervisor.CACertificate.CreateCertificate("oneinfra-cri", []string{"oneinfra"}, []string{hypervisorIPAddress})
	if err != nil {
		klog.Fatalf("error while creating oneinfra server certificate for hypervisor %q: %v", hypervisor.fullName(), err)
	}
	err = infraHypervisor.UploadFiles(
		"",
		"",
		map[string]string{
			haProxyCertBundlePath:   strings.Join([]string{criEndpointCertificate, criEndpointPrivateKey}, ""),
			haProxyClientCACertPath: hypervisor.ClientCACertificate.Certificate,
			haProxyConfigPath:       haProxyCfg,
		},
	)
	if err != nil {
		return err
	}
	if err := infraHypervisor.EnsureImage(haProxyImage); err != nil {
		return err
	}
	_, err = infraHypervisor.RunPod(nil, podapi.Pod{
		Name: "cri-endpoint",
		Containers: []podapi.Container{
			{
				Name:  "cri-endpoint",
				Image: haProxyImage,
				Mounts: map[string]string{
					haProxyCertBundlePath:                haProxyCertBundlePath,
					haProxyClientCACertPath:              haProxyClientCACertPath,
					haProxyConfigPath:                    "/etc/haproxy/haproxy.cfg",
					"/containerd-socket/containerd.sock": "/var/lib/haproxy/containerd.sock",
				},
			},
		},
		Ports: map[int]int{
			haProxyPort: haProxyPort,
		},
		Privileges: podapi.PrivilegesUnprivileged,
	})
	return err
}

func (hypervisor *Hypervisor) haProxyTemplate() (string, error) {
	template, err := template.New("").Parse(haProxyTemplate)
	if err != nil {
		return "", err
	}
	haProxyConfigData := struct {
		HAProxyPort             int
		HAProxyCertBundlePath   string
		HAProxyClientCACertPath string
	}{
		HAProxyPort:             haProxyPort,
		HAProxyCertBundlePath:   haProxyCertBundlePath,
		HAProxyClientCACertPath: haProxyClientCACertPath,
	}
	var rendered bytes.Buffer
	err = template.Execute(&rendered, haProxyConfigData)
	return rendered.String(), err
}

// Destroy destroys the current hypervisor
func (hypervisor *Hypervisor) Destroy() error {
	exec.Command(
		"docker", "rm", "-f", fmt.Sprintf("%s-%s", hypervisor.HypervisorCluster.Name, hypervisor.Name),
	).Run()
	return os.RemoveAll(hypervisor.runtimeDirectory())
}

func (hypervisor *Hypervisor) localContainerdSockDirectory() string {
	return "/containerd-socket"
}

func (hypervisor *Hypervisor) localContainerdSockPath() string {
	return fmt.Sprintf("unix://%s/containerd.sock", hypervisor.localContainerdSockDirectory())
}

func (hypervisor *Hypervisor) containerdSockPath() string {
	return filepath.Join(hypervisor.runtimeDirectory(), "containerd.sock")
}

func (hypervisor *Hypervisor) createRuntimeDirectory() error {
	return os.MkdirAll(hypervisor.runtimeDirectory(), 0700)
}

func (hypervisor *Hypervisor) runtimeDirectory() string {
	return filepath.Join(hypervisor.HypervisorCluster.directory(), hypervisor.Name)
}

func (hypervisor *Hypervisor) fullName() string {
	return fmt.Sprintf("%s-%s", hypervisor.HypervisorCluster.Name, hypervisor.Name)
}

// InternalIPAddress returns the internal IP address for the given
// container name
func InternalIPAddress(containerName string) (string, error) {
	ipAddress, err := exec.Command(
		"docker",
		"inspect", "-f", "{{ .NetworkSettings.IPAddress }}",
		containerName,
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(ipAddress), "\n"), nil
}

func (hypervisor *Hypervisor) internalIPAddress() (string, error) {
	return InternalIPAddress(hypervisor.fullName())
}

func (hypervisor *Hypervisor) createCACertificates() error {
	caCertificate, err := certificates.NewCertificateAuthority(hypervisor.fullName())
	if err != nil {
		return err
	}
	hypervisor.CACertificate = caCertificate
	clientCACertificate, err := certificates.NewCertificateAuthority(fmt.Sprintf("client-%s", hypervisor.fullName()))
	if err != nil {
		return err
	}
	hypervisor.ClientCACertificate = clientCACertificate
	return nil
}

// Export exports the local hypervisor to a versioned hypervisor
func (hypervisor *Hypervisor) Export() *infrav1alpha1.Hypervisor {
	ipAddress, err := hypervisor.internalIPAddress()
	if err != nil {
		klog.Fatalf("error while retrieving hypervisor IP address: %v", err)
	}
	res := infrav1alpha1.Hypervisor{
		ObjectMeta: metav1.ObjectMeta{
			Name: hypervisor.fullName(),
		},
		Spec: infrav1alpha1.HypervisorSpec{
			Public:    hypervisor.Public,
			IPAddress: ipAddress,
			PortRange: infrav1alpha1.HypervisorPortRange{
				Low:  hypervisor.ExposedPortRangeLow,
				High: hypervisor.ExposedPortRangeHigh,
			},
		},
	}
	if hypervisor.HypervisorCluster.Remote {
		internalIPAddress, err := hypervisor.internalIPAddress()
		if err != nil {
			klog.Fatalf("error while retrieving hypervisor internal IP address: %v", err)
		}
		clientCert, clientKey, err := hypervisor.ClientCACertificate.CreateCertificate("oneinfra-client", []string{"oneinfra"}, []string{})
		if err != nil {
			klog.Fatalf("error while creating oneinfra client certificate for hypervisor %q: %v", hypervisor.fullName(), err)
		}
		res.Spec.RemoteCRIEndpoint = &infrav1alpha1.RemoteHypervisorCRIEndpoint{
			CRIEndpoint:   net.JoinHostPort(internalIPAddress, strconv.Itoa(haProxyPort)),
			CACertificate: hypervisor.CACertificate.Certificate,
			ClientCertificate: &commonv1alpha1.Certificate{
				Certificate: clientCert,
				PrivateKey:  clientKey,
			},
		}
	} else {
		res.Spec.LocalCRIEndpoint = &infrav1alpha1.LocalHypervisorCRIEndpoint{
			CRIEndpoint: hypervisor.containerdSockPath(),
		}
	}
	return &res
}

// Wait waits for the local hypervisor to be created
func (hypervisor *Hypervisor) Wait() error {
	infraHypervisor := infra.NewLocalHypervisor(
		hypervisor.Name,
		hypervisor.containerdSockPath(),
	)
	for {
		_, runtimeErr := infraHypervisor.CRIRuntime()
		_, imageErr := infraHypervisor.CRIImage()
		if runtimeErr == nil && imageErr == nil {
			break
		}
	}
	return nil
}
