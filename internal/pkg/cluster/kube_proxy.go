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
	"fmt"
	"net"
	"net/url"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	kubeProxyImage = "k8s.gcr.io/kube-proxy:v%s"
)

// ReconcileKubeProxy reconciles the kube-proxy daemonset in this
// cluster
func (cluster *Cluster) ReconcileKubeProxy() error {
	client, err := cluster.KubernetesClient()
	if err != nil {
		return err
	}
	if err := cluster.reconcileKubeProxyPermissions(client); err != nil {
		return err
	}
	apiserverEndpoint, err := url.Parse(cluster.APIServerEndpoint)
	if err != nil {
		return err
	}
	apiserverEndpointHost, apiserverEndpointPort, err := net.SplitHostPort(apiserverEndpoint.Host)
	if err != nil {
		return err
	}
	trueVar := true
	hostPathFileOrCreateVar := corev1.HostPathFileOrCreate
	maxUnavailable := intstr.FromString("10%")
	_, err = client.AppsV1().DaemonSets(metav1.NamespaceSystem).Create(
		context.TODO(),
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-proxy",
				Namespace: metav1.NamespaceSystem,
				Labels: map[string]string{
					"k8s-app": "kube-proxy",
				},
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "kube-proxy",
					},
				},
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
					Type: appsv1.RollingUpdateDaemonSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDaemonSet{
						MaxUnavailable: &maxUnavailable,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "kube-proxy",
						},
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "kube-proxy",
						PriorityClassName:  "system-node-critical",
						HostNetwork:        true,
						Containers: []corev1.Container{
							{
								Name:    "kube-proxy",
								Image:   fmt.Sprintf(kubeProxyImage, cluster.KubernetesVersion),
								Command: []string{"/bin/sh"},
								Args:    []string{"-c", "kube-proxy"},
								Env: []corev1.EnvVar{
									{
										Name:  "KUBERNETES_SERVICE_HOST",
										Value: apiserverEndpointHost,
									},
									{
										Name:  "KUBERNETES_SERVICE_PORT",
										Value: apiserverEndpointPort,
									},
									{
										Name:  "KUBERNETES_SERVICE_PORT_HTTPS",
										Value: apiserverEndpointPort,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "var-log",
										MountPath: "/var/log",
									},
									{
										Name:      "xtables-lock",
										MountPath: "/run/xtables.lock",
									},
									{
										Name:      "lib-modules",
										MountPath: "/lib/modules",
										ReadOnly:  true,
									},
								},
								SecurityContext: &corev1.SecurityContext{
									Privileged: &trueVar,
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "var-log",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/var/log",
									},
								},
							},
							{
								Name: "xtables-lock",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/run/xtables.lock",
										Type: &hostPathFileOrCreateVar,
									},
								},
							},
							{
								Name: "lib-modules",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/lib/modules",
									},
								},
							},
						},
						Tolerations: []corev1.Toleration{
							{
								Operator: corev1.TolerationOpExists,
							},
						},
					},
				},
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (cluster *Cluster) reconcileKubeProxyPermissions(client clientset.Interface) error {
	_, err := client.CoreV1().ServiceAccounts(metav1.NamespaceSystem).Create(
		context.TODO(),
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-proxy",
				Namespace: metav1.NamespaceSystem,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = client.RbacV1().ClusterRoleBindings().Create(
		context.TODO(),
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:kube-proxy",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "kube-proxy",
					Namespace: metav1.NamespaceSystem,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     "system:node-proxier",
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
