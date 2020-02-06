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
	"context"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"oneinfra.ereslibre.es/m/internal/pkg/infra"
)

type Node struct {
	hypervisor *infra.Hypervisor
}

func NewNode(hypervisor *infra.Hypervisor) Node {
	return Node{
		hypervisor: hypervisor,
	}
}

func (node *Node) Reconcile() error {
	if err := node.PullImage("k8s.gcr.io/kube-apiserver:v1.17.0"); err != nil {
		return err
	}
	if err := node.PullImage("k8s.gcr.io/kube-controller-manager:v1.17.0"); err != nil {
		return err
	}
	if err := node.PullImage("k8s.gcr.io/kube-scheduler:v1.17.0"); err != nil {
		return err
	}
	return nil
}

func (node *Node) PullImage(image string) error {
	_, err := node.hypervisor.CRIImage.PullImage(context.Background(), &criapi.PullImageRequest{
		Image: &criapi.ImageSpec{
			Image: image,
		},
	})
	return err
}
