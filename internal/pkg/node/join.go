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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
)

// Join joins a node to an existing cluster
func Join(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint string) error {
	client, err := createClient(apiServerEndpoint, caCertificate, token)
	if err != nil {
		return err
	}
	keyPair, err := readOrGenerateKeyPair()
	if err != nil {
		return err
	}
	if err := createJoinRequest(client, apiServerEndpoint, nodename, keyPair, containerRuntimeEndpoint, imageServiceEndpoint); err != nil {
		return err
	}
	nodeJoinRequest, err := waitForJoinRequestIssuedCondition(client, nodename, 5*time.Minute)
	if err != nil {
		return err
	}
	if err := writeKubeConfig(nodeJoinRequest, keyPair); err != nil {
		return err
	}
	if err := writeKubeletConfig(nodeJoinRequest, keyPair); err != nil {
		return err
	}
	if err := setupSystemd(nodeJoinRequest); err != nil {
		return err
	}
	if err := startKubelet(); err != nil {
		return err
	}
	return nil
}

func createClient(apiServerEndpoint, caCertificate, token string) (*restclient.RESTClient, error) {
	kubeConfig, err := cluster.KubeConfigWithToken("cluster", apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	if err := nodev1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	client, err := cluster.RESTClientFromKubeConfig(kubeConfig, &nodev1alpha1.GroupVersion, scheme)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func createJoinRequest(client *restclient.RESTClient, apiServerEndpoint, nodename string, keyPair *certificates.KeyPair, containerRuntimeEndpoint, imageServiceEndpoint string) error {
	nodeJoinRequest := nodev1alpha1.NodeJoinRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodename,
			Namespace: constants.OneInfraNamespace,
		},
		Spec: nodev1alpha1.NodeJoinRequestSpec{
			PublicKey:                keyPair.PublicKey,
			APIServerEndpoint:        apiServerEndpoint,
			ContainerRuntimeEndpoint: containerRuntimeEndpoint,
			ImageServiceEndpoint:     imageServiceEndpoint,
		},
	}
	err := client.
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

func waitForJoinRequestIssuedCondition(client *restclient.RESTClient, nodename string, timeout time.Duration) (*nodejoinrequests.NodeJoinRequest, error) {
	timeoutChan := time.After(timeout)
	tickChan := time.Tick(time.Second)
	for {
		select {
		case <-timeoutChan:
			return nil, errors.New("timed out waiting for issued condition")
		case <-tickChan:
			klog.V(2).Info("checking if the node join request has been issued")
			nodeJoinRequest := nodev1alpha1.NodeJoinRequest{}
			err := client.
				Get().
				Namespace(constants.OneInfraNamespace).
				Resource("nodejoinrequests").
				Name(nodename).
				Do().
				Into(&nodeJoinRequest)
			if err != nil {
				continue
			}
			if nodeJoinRequest.HasCondition(nodev1alpha1.Issued) {
				nodeJoinRequestInternal, err := nodejoinrequests.NewNodeJoinRequestFromv1alpha1(&nodeJoinRequest)
				if err != nil {
					return nil, errors.New("could not convert node join request")
				}
				return nodeJoinRequestInternal, nil
			}
		}
	}
}

func readOrGenerateKeyPair() (*certificates.KeyPair, error) {
	var keyPair *certificates.KeyPair
	if _, err := os.Stat(filepath.Join(constants.OneInfraConfigDir, "join.key")); os.IsNotExist(err) {
		var err error
		keyPair, err = certificates.NewPrivateKey()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(constants.OneInfraConfigDir, 0700); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(filepath.Join(constants.OneInfraConfigDir, "join.key"), []byte(keyPair.PrivateKey), 0600); err != nil {
			return nil, err
		}
	} else {
		var err error
		keyPair, err = certificates.NewPrivateKeyFromFile(filepath.Join(constants.OneInfraConfigDir, "join.key"))
		if err != nil {
			return nil, err
		}
	}
	return keyPair, nil
}

func writeKubeConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, keyPair *certificates.KeyPair) error {
	kubeConfig, err := keyPair.Decrypt(nodeJoinRequest.KubeConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletKubeConfigPath, []byte(kubeConfig), 0600)
}

func writeKubeletConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, keyPair *certificates.KeyPair) error {
	kubeletConfig, err := keyPair.Decrypt(nodeJoinRequest.KubeletConfig)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletConfigPath, []byte(kubeletConfig), 0600)
}

func setupSystemd(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	return nil
}

func startKubelet() error {
	return nil
}
