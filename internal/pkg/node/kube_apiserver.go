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

import "oneinfra.ereslibre.es/m/internal/pkg/infra"

const (
	dqliteImage        = "oneinfra/dqlite:latest"
	kineImage          = "oneinfra/kine:latest"
	kubeApiServerImage = "k8s.gcr.io/kube-apiserver:v1.17.0"
)

type KubeAPIServer struct {
	node *Node
}

func (kubeApiServer *KubeAPIServer) Reconcile() error {
	if err := kubeApiServer.node.hypervisor.PullImages(kineImage, kubeApiServerImage); err != nil {
		return err
	}
	return kubeApiServer.node.hypervisor.RunPod(
		infra.NewPod(
			"kube-apiserver",
			[]infra.Container{
				{
					Name:    "kine",
					Image:   kineImage,
					Command: []string{"kine"},
				},
				{
					Name:    "kube-apiserver",
					Image:   kubeApiServerImage,
					Command: []string{"kube-apiserver", "--etcd-servers", "http://127.0.0.1:2379"},
				},
			},
		),
	)
}
