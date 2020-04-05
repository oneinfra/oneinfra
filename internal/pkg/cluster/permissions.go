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
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

const (
	oneInfraNodeJoinRequesterRoleName = "oneinfra:node:join-requester"
)

// ReconcilePermissions reconciles this cluster namespaces
func (cluster *Cluster) ReconcilePermissions() error {
	client, err := cluster.KubernetesClient()
	if err != nil {
		return err
	}
	_, err = client.RbacV1().Roles(constants.OneInfraNamespace).Create(&rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oneInfraNodeJoinRequesterRoleName,
			Namespace: constants.OneInfraNamespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{nodev1alpha1.GroupVersion.Group},
				Resources: []string{"nodejoinrequests"},
				Verbs:     []string{"get", "create"},
			},
			{
				APIGroups:     []string{""},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{constants.OneInfraJoinConfigMap},
				Verbs:         []string{"get"},
			},
		},
	})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = client.RbacV1().RoleBindings(constants.OneInfraNamespace).Create(&rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oneInfraNodeJoinRequesterRoleName,
			Namespace: constants.OneInfraNamespace,
		},
		Subjects: []rbacv1.Subject{
			{
				APIGroup: rbacv1.GroupName,
				Kind:     "Group",
				Name:     constants.OneInfraNodeJoinTokenExtraGroups,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     oneInfraNodeJoinRequesterRoleName,
		},
	})
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
