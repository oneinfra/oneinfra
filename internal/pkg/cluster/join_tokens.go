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

package cluster

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	tokenapi "k8s.io/cluster-bootstrap/token/api"
	tokenutil "k8s.io/cluster-bootstrap/token/util"
	utilsecrets "k8s.io/cluster-bootstrap/util/secrets"
	utiltokens "k8s.io/cluster-bootstrap/util/tokens"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

// ReconcileJoinTokens reconciles this cluster join tokens
func (cluster *Cluster) ReconcileJoinTokens() error {
	client, err := cluster.KubernetesClient()
	if err != nil {
		return err
	}
	cluster.CurrentJoinTokens = []string{}
	secretList, err := client.CoreV1().Secrets(metav1.NamespaceSystem).List(
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("type=%s", tokenapi.SecretTypeBootstrapToken),
		},
	)
	if err != nil {
		return err
	}
	for _, secret := range secretList.Items {
		cluster.CurrentJoinTokens = append(
			cluster.CurrentJoinTokens,
			tokenFromSecret(&secret),
		)
	}
	tokensSuccessfullyReconciled := true
	if err := cluster.createNewTokens(client); err != nil {
		tokensSuccessfullyReconciled = false
	}
	if err := cluster.removeExcessTokens(client); err != nil {
		tokensSuccessfullyReconciled = false
	}
	if tokensSuccessfullyReconciled {
		cluster.CurrentJoinTokens = cluster.DesiredJoinTokens
		return nil
	}
	return errors.New("some join tokens could not be successfully reconciled")
}

func (cluster *Cluster) newTokens() []string {
	return substractTokens(cluster.DesiredJoinTokens, cluster.CurrentJoinTokens)
}

func (cluster *Cluster) excessTokens() []string {
	return substractTokens(cluster.CurrentJoinTokens, cluster.DesiredJoinTokens)
}

func (cluster *Cluster) createNewTokens(client clientset.Interface) error {
	allSucceeded := true
	for _, newToken := range cluster.newTokens() {
		tokenID, tokenSecret, err := utiltokens.ParseToken(newToken)
		if err != nil {
			allSucceeded = false
			continue
		}
		tokenSecretName := tokenutil.BootstrapTokenSecretName(tokenID)
		_, err = client.CoreV1().Secrets(metav1.NamespaceSystem).Create(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tokenSecretName,
					Namespace: metav1.NamespaceSystem,
				},
				StringData: map[string]string{
					tokenapi.BootstrapTokenDescriptionKey:      "oneinfra node join token",
					tokenapi.BootstrapTokenIDKey:               tokenID,
					tokenapi.BootstrapTokenSecretKey:           tokenSecret,
					tokenapi.BootstrapTokenUsageAuthentication: "true",
					tokenapi.BootstrapTokenExtraGroupsKey:      constants.OneInfraNodeJoinTokenExtraGroups,
				},
				Type: corev1.SecretTypeBootstrapToken,
			},
		)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			allSucceeded = false
			klog.Warningf("could not create new join token %q in cluster %q: %v", tokenSecretName, cluster.Name, err)
			continue
		}
	}
	if allSucceeded {
		return nil
	}
	return errors.New("not all new join tokens could be reconciled")
}

func (cluster *Cluster) removeExcessTokens(client clientset.Interface) error {
	allSucceeded := true
	for _, excessToken := range cluster.excessTokens() {
		tokenID, _, err := utiltokens.ParseToken(excessToken)
		if err != nil {
			allSucceeded = false
			continue
		}
		tokenSecretName := tokenutil.BootstrapTokenSecretName(tokenID)
		err = client.CoreV1().Secrets(metav1.NamespaceSystem).Delete(
			tokenSecretName,
			&metav1.DeleteOptions{},
		)
		if err != nil {
			allSucceeded = false
			klog.Warningf("could not delete excess join token %q in cluster %q: %v", tokenSecretName, cluster.Name, err)
		}
	}
	if allSucceeded {
		return nil
	}
	return errors.New("not all excess join tokens could be deleted")
}

func substractTokens(list []string, listToSubstract []string) []string {
	res := []string{}
	toSubstract := map[string]struct{}{}
	for _, token := range listToSubstract {
		toSubstract[token] = struct{}{}
	}
	for _, token := range list {
		if _, exists := toSubstract[token]; !exists {
			res = append(res, token)
		}
	}
	return res
}

func tokenFromSecret(secret *corev1.Secret) string {
	tokenID := utilsecrets.GetData(secret, tokenapi.BootstrapTokenIDKey)
	tokenSecret := utilsecrets.GetData(secret, tokenapi.BootstrapTokenSecretKey)
	return tokenutil.TokenFromIDAndSecret(tokenID, tokenSecret)
}
