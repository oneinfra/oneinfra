/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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

package cluster

import (
	"context"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oneinfra/oneinfra/pkg/constants"
)

// ReconcileNamespaces reconciles this cluster namespaces
func (cluster *Cluster) ReconcileNamespaces() error {
	client, err := cluster.KubernetesClient()
	if err != nil {
		return err
	}
	_, err = client.CoreV1().Namespaces().Create(
		context.TODO(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.OneInfraNamespace,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
