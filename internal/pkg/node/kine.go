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
	dqliteImage = "oneinfra/dqlite:latest"
	kineImage   = "oneinfra/kine:latest"
)

type Kine struct {
	node *Node
}

func (kine *Kine) Reconcile() error {
	if err := kine.node.hypervisor.PullImages(dqliteImage, kineImage); err != nil {
		return err
	}
	return kine.node.hypervisor.RunPod(
		infra.NewRegularPod(
			"kine",
			[]infra.Container{
				{
					Name:    "dqlite",
					Image:   dqliteImage,
					Command: []string{"dqlite-demo", "start", "1", "--address", "0.0.0.0:9181"},
				},
				{
					Name:    "kine",
					Image:   kineImage,
					Command: []string{"kine", "--endpoint", "dqlite://127.0.0.1:9181"},
				},
			},
		),
	)
}
