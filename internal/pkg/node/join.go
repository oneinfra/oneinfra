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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

// Join joins a node to an existing cluster
func Join(nodename, apiServerEndpoint, caCertificate, token string) error {
	kubeConfig, err := cluster.KubeConfigWithToken("cluster", apiServerEndpoint, caCertificate, token)
	if err != nil {
		return err
	}
	scheme := runtime.NewScheme()
	if err := nodev1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	client, err := cluster.RESTClientFromKubeConfig(kubeConfig, &nodev1alpha1.GroupVersion, scheme)
	if err != nil {
		return err
	}
	keyPair, err := certificates.NewPrivateKey()
	if err != nil {
		return err
	}
	nodeJoinRequest := nodev1alpha1.NodeJoinRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodename,
			Namespace: constants.OneInfraNamespace,
		},
		Spec: nodev1alpha1.NodeJoinRequestSpec{
			PublicKey: keyPair.PublicKey,
		},
	}
	err = client.
		Post().
		Namespace(constants.OneInfraNamespace).
		Resource("nodejoinrequests").
		Body(&nodeJoinRequest).
		Do().
		Error()
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func setupSystemd() error {
	return nil
}

func retrieveKubeletConfig() (string, error) {
	return "", nil
}

func retrieveKubeConfig() (string, error) {
	return "", nil
}

func startKubelet() error {
	return nil
}
