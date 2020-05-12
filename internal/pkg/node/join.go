/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
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

package node

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
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
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	podapi "github.com/oneinfra/oneinfra/internal/pkg/infra/pod"
	nodejoinrequests "github.com/oneinfra/oneinfra/internal/pkg/node-join-requests"
	oneinframanagedclientset "github.com/oneinfra/oneinfra/pkg/clientsets/managed"
	"github.com/oneinfra/oneinfra/pkg/constants"
)

// Join joins a node to an existing cluster
func Join(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint string, extraSANs []string) error {
	klog.Info("loading or generating symmetric key")
	symmetricKey, err := readOrGenerateSymmetricKey()
	if err != nil {
		return err
	}
	nodeJoinRequest, err := createAndWaitForJoinRequest(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint, symmetricKey, extraSANs)
	if err != nil {
		return err
	}
	if nodeJoinRequest.VPNEnabled {
		// TODO: set up wireguard
	}
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

func createOneInfraManagedClient(apiServerEndpoint, caCertificate, token string) (oneinframanagedclientset.Interface, error) {
	kubeConfig, err := cluster.KubeConfigWithToken("cluster", apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	return cluster.OneInfraManagedClientFromKubeConfig(kubeConfig)
}

func createJoinRequest(client oneinframanagedclientset.Interface, apiServerEndpoint, nodename, symmetricKey, containerRuntimeEndpoint, imageServiceEndpoint string, extraSANs []string) error {
	nodeJoinRequest := nodev1alpha1.NodeJoinRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodename,
		},
		Spec: nodev1alpha1.NodeJoinRequestSpec{
			SymmetricKey:             symmetricKey,
			APIServerEndpoint:        apiServerEndpoint,
			ContainerRuntimeEndpoint: containerRuntimeEndpoint,
			ImageServiceEndpoint:     imageServiceEndpoint,
			ExtraSANs:                extraSANs,
		},
	}
	_, err := client.NodeV1alpha1().NodeJoinRequests().Create(
		context.TODO(),
		&nodeJoinRequest,
		metav1.CreateOptions{},
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func waitForJoinRequestIssuedCondition(client oneinframanagedclientset.Interface, nodename string, timeout time.Duration) (*nodejoinrequests.NodeJoinRequest, error) {
	klog.Infof("waiting for join request %q to be issued; will timeout in %s", nodename, timeout)
	timeoutChan := time.After(timeout)
	tickChan := time.Tick(time.Second)
	for {
		select {
		case <-timeoutChan:
			return nil, errors.New("timed out waiting for issued condition")
		case <-tickChan:
			klog.Info("waiting for the node join request to be issued")
			nodeJoinRequest, err := client.NodeV1alpha1().NodeJoinRequests().Get(
				context.TODO(),
				nodename,
				metav1.GetOptions{},
			)
			if err != nil {
				continue
			}
			if nodeJoinRequest.Status.Conditions.IsCondition(nodev1alpha1.Issued, commonv1alpha1.ConditionTrue) {
				nodeJoinRequestInternal, err := nodejoinrequests.NewNodeJoinRequestFromv1alpha1(nodeJoinRequest, nil)
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

func writeKubeletServerCertificate(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	certificate, err := decrypt(symmetricKey, nodeJoinRequest.KubeletServerCertificate)
	if err != nil {
		return err
	}
	privateKey, err := decrypt(symmetricKey, nodeJoinRequest.KubeletServerPrivateKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(constants.OneInfraConfigDir, 0700); err != nil {
		return err
	}
	if err := ioutil.WriteFile(constants.KubeletServerCertificatePath, []byte(certificate), 0600); err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletServerPrivateKeyPath, []byte(privateKey), 0600)
}

func writeKubeletClientCACertificate(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	clientCACertificate, err := decrypt(symmetricKey, nodeJoinRequest.KubeletClientCACertificate)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(constants.OneInfraConfigDir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(constants.KubeletClientCACertificatePath, []byte(clientCACertificate), 0600)
}

func installKubelet(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	kubernetesVersion, err := decrypt(symmetricKey, nodeJoinRequest.KubernetesVersion)
	if err != nil {
		return err
	}
	kubeletImage := fmt.Sprintf(kubeletInstallerImage, kubernetesVersion)
	klog.Infof("installing the kubelet from %q", kubeletImage)
	hypervisorImageEndpoint := infra.NewLocalHypervisor(nodeJoinRequest.Name, nodeJoinRequest.ImageServiceEndpoint)
	hypervisorRuntimeEndpoint := infra.NewLocalHypervisor(nodeJoinRequest.Name, nodeJoinRequest.ContainerRuntimeEndpoint)
	if err := hypervisorImageEndpoint.EnsureImage(kubeletImage); err != nil {
		return err
	}
	return hypervisorRuntimeEndpoint.RunAndWaitForPod(
		"",
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
			// This will make the container runtime skip CNI when creating
			// the kubelet installer container.
			Privileges: podapi.PrivilegesNetworkPrivileged,
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

func createAndWaitForJoinRequest(nodename, apiServerEndpoint, caCertificate, token, containerRuntimeEndpoint, imageServiceEndpoint, symmetricKey string, extraSANs []string) (*nodejoinrequests.NodeJoinRequest, error) {
	oneinfraManagedClient, err := createOneInfraManagedClient(apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	kubernetesClient, err := createKubernetesClient(apiServerEndpoint, caCertificate, token)
	if err != nil {
		return nil, err
	}
	klog.Info("downloading oneinfra public ConfigMap")
	oneinfraPublicConfigMap, err := kubernetesClient.CoreV1().ConfigMaps(constants.OneInfraNamespace).Get(
		context.TODO(),
		constants.OneInfraJoinConfigMap,
		metav1.GetOptions{},
	)
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
	klog.Infof("creating node join request for nodename %q", nodename)
	if err := createJoinRequest(oneinfraManagedClient, apiServerEndpoint, nodename, cryptedSymmetricKey, containerRuntimeEndpoint, imageServiceEndpoint, extraSANs); err != nil {
		return nil, err
	}
	return waitForJoinRequestIssuedCondition(oneinfraManagedClient, nodename, 5*time.Minute)
}

func setupKubelet(nodeJoinRequest *nodejoinrequests.NodeJoinRequest, symmetricKey string) error {
	klog.Info("writing kubelet configuration and secrets")
	if err := writeKubeConfig(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := writeKubeletConfig(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := writeKubeletServerCertificate(nodeJoinRequest, symmetricKey); err != nil {
		return err
	}
	if err := writeKubeletClientCACertificate(nodeJoinRequest, symmetricKey); err != nil {
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
