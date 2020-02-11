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

package node

import (
	"fmt"
	"path/filepath"

	"oneinfra.ereslibre.es/m/internal/pkg/cluster"
	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

const (
	dqliteImage        = "oneinfra/dqlite:latest"
	kineImage          = "oneinfra/kine:latest"
	kubeAPIServerImage = "k8s.gcr.io/kube-apiserver:v1.17.0"
)

// KubeAPIServer represents the kube-apiserver
type KubeAPIServer struct{}

func (kubeAPIServer *KubeAPIServer) secretsPath(cluster *cluster.Cluster) string {
	return filepath.Join("/etc/kubernetes/clusters", cluster.Name)
}

func (kubeAPIServer *KubeAPIServer) secretsPathFile(cluster *cluster.Cluster, file string) string {
	return filepath.Join(kubeAPIServer.secretsPath(cluster), file)
}

// Reconcile reconciles the kube-apiserver
func (kubeAPIServer *KubeAPIServer) Reconcile(hypervisor *infra.Hypervisor, cluster *cluster.Cluster) error {
	if err := hypervisor.PullImages(kineImage, kubeAPIServerImage); err != nil {
		return err
	}
	err := hypervisor.UploadFiles(
		map[string]string{
			kubeAPIServer.secretsPathFile(cluster, "apiserver-client-ca.crt"): cluster.CertificateAuthorities.APIServerClient.Certificate,
			kubeAPIServer.secretsPathFile(cluster, "apiserver.crt"):           cluster.APIServer.TLSCert,
			kubeAPIServer.secretsPathFile(cluster, "apiserver.key"):           cluster.APIServer.TLSPrivateKey,
		},
	)
	if err != nil {
		return err
	}
	_, err = hypervisor.RunPod(
		infra.NewPod(
			fmt.Sprintf("kube-apiserver-%s", cluster.Name),
			[]infra.Container{
				{
					Name:    "kine",
					Image:   kineImage,
					Command: []string{"kine"},
				},
				{
					Name:    "kube-apiserver",
					Image:   kubeAPIServerImage,
					Command: []string{"kube-apiserver"},
					Args: []string{
						"--etcd-servers", "http://127.0.0.1:2379",
						"--tls-cert-file", kubeAPIServer.secretsPathFile(cluster, "apiserver.crt"),
						"--tls-private-key-file", kubeAPIServer.secretsPathFile(cluster, "apiserver.key"),
						"--client-ca-file", kubeAPIServer.secretsPathFile(cluster, "apiserver-client-ca.crt"),
					},
					Mounts: map[string]string{
						kubeAPIServer.secretsPath(cluster): kubeAPIServer.secretsPath(cluster),
					},
				},
			},
		),
	)
	return err
}
