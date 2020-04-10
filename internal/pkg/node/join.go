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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
)

// Join joins a node to an existing cluster
func Join(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint string) error {
	symmetricKey, err := readOrGenerateSymmetricKey()
	if err != nil {
		return err
	}
	nodeJoinRequest, err := createAndWaitForJoinRequest(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint, symmetricKey)
	if err != nil {
		return err
	}
	// TODO: set up wireguard
	return setupKubelet(nodeJoinRequest, symmetricKey)
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

func createKubernetesClient(apiServerEndpoint, caCertificate, token string) (clientset.Interface, error) {
	kubeConfig, err := cluster.KubeConfigWithToken("cluster", apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	return cluster.KubernetesClientFromKubeConfig(kubeConfig)
}

func createJoinRequest(client *restclient.RESTClient, apiServerEndpoint, nodename, symmetricKey, containerRuntimeEndpoint, imageServiceEndpoint string) error {
	nodeJoinRequest := nodev1alpha1.NodeJoinRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodename,
			Namespace: constants.OneInfraNamespace,
		},
		Spec: nodev1alpha1.NodeJoinRequestSpec{
			SymmetricKey:             symmetricKey,
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
			if nodeJoinRequest.Status.Conditions.IsCondition(nodev1alpha1.Issued, commonv1alpha1.ConditionTrue) {
				nodeJoinRequestInternal, err := nodejoinrequests.NewNodeJoinRequestFromv1alpha1(&nodeJoinRequest, nil)
				if err != nil {
					return nil, errors.New("could not convert node join request")
				}
				return nodeJoinRequestInternal, nil
			}
		}
	}
}

func readOrGenerateSymmetricKey() (string, error) {
	var symmetricKey string
	if _, err := os.Stat(filepath.Join(constants.OneInfraConfigDir, "join.key")); os.IsNotExist(err) {
		symmetricKeyRaw := make([]byte, 16)
		_, err := rand.Read(symmetricKeyRaw)
		if err != nil {
			return "", err
		}
		symmetricKey = fmt.Sprintf("%x", symmetricKeyRaw)
		if err := os.MkdirAll(constants.OneInfraConfigDir, 0700); err != nil {
			return "", err
		}
		if err := ioutil.WriteFile(filepath.Join(constants.OneInfraConfigDir, "join.key"), []byte(symmetricKey), 0600); err != nil {
			return "", err
		}
	} else {
		symmetricKeyBytes, err := ioutil.ReadFile(filepath.Join(constants.OneInfraConfigDir, "join.key"))
		if err != nil {
			return "", err
		}
		symmetricKey = string(symmetricKeyBytes)
	}
	return symmetricKey, nil
}

func writeKubeConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	kubeConfig, err := decrypt(symmetricKey, nodeJoinRequest.KubeConfig)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(constants.OneInfraConfigDir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletKubeConfigPath, []byte(kubeConfig), 0600)
}

func writeKubeletConfig(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	kubeletConfig, err := decrypt(symmetricKey, nodeJoinRequest.KubeletConfig)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(constants.KubeletDir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletConfigPath, []byte(kubeletConfig), 0600)
}

func writeKubeletCertificate(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	certificate, err := decrypt(symmetricKey, nodeJoinRequest.KubeletServerCertificate)
	if err != nil {
		return err
	}
	privateKey, err := decrypt(symmetricKey, nodeJoinRequest.KubeletServerPrivateKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(constants.KubeletDir, 0700); err != nil {
		return err
	}
	if err := ioutil.WriteFile(constants.KubeletServerCertificatePath, []byte(certificate), 0600); err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletServerPrivateKeyPath, []byte(privateKey), 0600)
}

func installKubelet(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	kubernetesVersion, err := decrypt(symmetricKey, nodeJoinRequest.KubernetesVersion)
	if err != nil {
		return err
	}
	hypervisorImageEndpoint := infra.NewLocalHypervisor(nodeJoinRequest.Name, nodeJoinRequest.ImageServiceEndpoint)
	hypervisorRuntimeEndpoint := infra.NewLocalHypervisor(nodeJoinRequest.Name, nodeJoinRequest.ContainerRuntimeEndpoint)
	err = hypervisorImageEndpoint.EnsureImage(
		fmt.Sprintf(kubeletInstallerImage, kubernetesVersion),
	)
	if err != nil {
		return err
	}
	return hypervisorRuntimeEndpoint.RunAndWaitForPod(
		nodeJoinRequest.Name,
		"join",
		podapi.Pod{
			Name: "kubelet-installer",
			Containers: []podapi.Container{
				{
					Name:  "kubelet-installer",
					Image: fmt.Sprintf(kubeletInstallerImage, kubernetesVersion),
					Mounts: map[string]string{
						"/usr/local/bin": "/host",
					},
				},
			},
		})
}

func setupSystemd(nodeJoinRequest *nodejoinrequests.NodeJoinRequest) error {
	kubeletSystemdServiceTpl, err := template.New("").Parse(kubeletSystemdServiceTemplate)
	if err != nil {
		return err
	}
	var kubeletSystemdService bytes.Buffer
	err = kubeletSystemdServiceTpl.Execute(&kubeletSystemdService, struct {
		Nodename                 string
		KubeletKubeConfigPath    string
		KubeletConfigPath        string
		ImageServiceEndpoint     string
		ContainerRuntimeEndpoint string
	}{
		Nodename:                 nodeJoinRequest.Name,
		KubeletKubeConfigPath:    constants.KubeletKubeConfigPath,
		KubeletConfigPath:        constants.KubeletConfigPath,
		ImageServiceEndpoint:     nodeJoinRequest.ImageServiceEndpoint,
		ContainerRuntimeEndpoint: nodeJoinRequest.ContainerRuntimeEndpoint,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(systemdDir, "kubelet.service"), kubeletSystemdService.Bytes(), 0644)
}

func startKubelet() error {
	return exec.Command("systemctl", "enable", "--now", "kubelet").Run()
}

func createAndWaitForJoinRequest(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint, symmetricKey string) (*nodejoinrequests.NodeJoinRequest, error) {
	client, err := createClient(apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	kubernetesClient, err := createKubernetesClient(apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	oneinfraPublicConfigMap, err := kubernetesClient.CoreV1().ConfigMaps(constants.OneInfraNamespace).Get(constants.OneInfraJoinConfigMap, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	joinPublicKeyPEM, exists := oneinfraPublicConfigMap.Data[constants.OneInfraJoinConfigMapJoinKey]
	if !exists {
		return nil, errors.Errorf("could not find field %q in ConfigMap %q (in namespace %q)", constants.OneInfraJoinConfigMapJoinKey, constants.OneInfraJoinConfigMap, metav1.NamespacePublic)
	}
	joinPublicKey, err := crypto.NewPublicKeyFromString(joinPublicKeyPEM)
	if err != nil {
		return nil, errors.New("could not read a public key")
	}
	cryptedSymmetricKey, err := joinPublicKey.Encrypt(symmetricKey)
	if err != nil {
		return nil, err
	}
	if err := createJoinRequest(client, apiServerEndpoint, nodename, cryptedSymmetricKey, containerRuntimeEndpoint, imageServiceEndpoint); err != nil {
		return nil, err
	}
	return waitForJoinRequestIssuedCondition(client, nodename, 5*time.Minute)
}

func setupKubelet(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	if err := writeKubeConfig(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := writeKubeletConfig(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := writeKubeletCertificate(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := installKubelet(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := setupSystemd(nodeJoinRequest); err != nil {
		return err
	}
	return startKubelet()
}

func decrypt(symmetricKey string, base64Data string) (string, error) {
	block, err := aes.NewCipher([]byte(symmetricKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	data, err := base64.RawStdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
