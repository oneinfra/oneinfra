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

package cluster

import (
	"fmt"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	versions "github.com/oneinfra/oneinfra/pkg/versions"
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
	coreDNSImage     = "coredns/coredns:%s"
	corefileContents = `.:53 {
    errors
    health {
      lameduck 5s
    }
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      fallthrough in-addr.arpa ip6.arpa
    }
    prometheus :9153
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}`
)

// ReconcileCoreDNS reconciles the CoreDNS deployment in this cluster
func (cluster *Cluster) ReconcileCoreDNS() error {
	client, err := cluster.KubernetesClient()
	if err != nil {
		return err
	}
	if err := cluster.reconcileCoreDNSPermissions(client); err != nil {
		return err
	}
	if err := cluster.reconcileCoreDNSConfigMap(client); err != nil {
		return err
	}
	if err := cluster.reconcileCoreDNSDeployment(client); err != nil {
		return err
	}
	if err := cluster.reconcileCoreDNSService(client); err != nil {
		return err
	}
	return nil
}

func (cluster *Cluster) reconcileCoreDNSPermissions(client clientset.Interface) error {
	_, err := client.CoreV1().ServiceAccounts(metav1.NamespaceSystem).Create(
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: metav1.NamespaceSystem,
			},
		},
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = client.RbacV1().ClusterRoles().Create(
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:coredns",
				Labels: map[string]string{
					"kubernetes.io/bootstrapping": "rbac-defaults",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"endpoints", "services", "pods", "namespaces"},
					Verbs:     []string{"list", "watch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"nodes"},
					Verbs:     []string{"get"},
				},
			},
		},
	)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	_, err = client.RbacV1().ClusterRoleBindings().Create(
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:coredns",
				Annotations: map[string]string{
					"rbac.authorization.kubernetes.io/autoupdate": "true",
				},
				Labels: map[string]string{
					"kubernetes.io/bootstrapping": "rbac-defaults",
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     "system:coredns",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "coredns",
					Namespace: metav1.NamespaceSystem,
				},
			},
		},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (cluster *Cluster) reconcileCoreDNSConfigMap(client clientset.Interface) error {
	_, err := client.CoreV1().ConfigMaps(metav1.NamespaceSystem).Create(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: metav1.NamespaceSystem,
			},
			Data: map[string]string{
				"Corefile": corefileContents,
			},
		},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (cluster *Cluster) reconcileCoreDNSDeployment(client clientset.Interface) error {
	oneVar := intstr.FromInt(1)
	trueVar := true
	falseVar := false
	coreDNSVersion, err := constants.KubernetesComponentVersion(cluster.KubernetesVersion, versions.CoreDNS)
	if err != nil {
		return err
	}
	_, err = client.AppsV1().Deployments(metav1.NamespaceSystem).Create(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coredns",
				Namespace: metav1.NamespaceSystem,
				Labels: map[string]string{
					"k8s-app":            "kube-dns",
					"kubernetes.io/name": "CoreDNS",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RollingUpdateDeploymentStrategyType,
					RollingUpdate: &appsv1.RollingUpdateDeployment{
						MaxUnavailable: &oneVar,
					},
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "kube-dns",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "kube-dns",
						},
					},
					Spec: corev1.PodSpec{
						PriorityClassName:  "system-cluster-critical",
						ServiceAccountName: "coredns",
						Tolerations: []corev1.Toleration{
							{
								Key:      "CriticalAddonsOnly",
								Operator: corev1.TolerationOpExists,
							},
						},
						NodeSelector: map[string]string{
							"kubernetes.io/os": "linux",
						},
						Affinity: &corev1.Affinity{
							PodAntiAffinity: &corev1.PodAntiAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
									{
										LabelSelector: &metav1.LabelSelector{
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "k8s-app",
													Operator: metav1.LabelSelectorOpIn,
													Values:   []string{"kube-dns"},
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "coredns",
								Image: fmt.Sprintf(coreDNSImage, coreDNSVersion),
								Args:  []string{"-conf", "/etc/coredns/Corefile"},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "config-volume",
										MountPath: "/etc/coredns",
										ReadOnly:  true,
									},
								},
								Ports: []corev1.ContainerPort{
									{
										Name:          "dns",
										Protocol:      corev1.ProtocolUDP,
										ContainerPort: 53,
									},
									{
										Name:          "dns-tcp",
										Protocol:      corev1.ProtocolTCP,
										ContainerPort: 53,
									},
									{
										Name:          "metrics",
										Protocol:      corev1.ProtocolTCP,
										ContainerPort: 9153,
									},
								},
								SecurityContext: &corev1.SecurityContext{
									AllowPrivilegeEscalation: &falseVar,
									Capabilities: &corev1.Capabilities{
										Add:  []corev1.Capability{"NET_BIND_SERVICE"},
										Drop: []corev1.Capability{"all"},
									},
									ReadOnlyRootFilesystem: &trueVar,
								},
								LivenessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path:   "/health",
											Port:   intstr.FromInt(8080),
											Scheme: corev1.URISchemeHTTP,
										},
									},
									InitialDelaySeconds: 60,
									TimeoutSeconds:      5,
									SuccessThreshold:    1,
									FailureThreshold:    5,
								},
								ReadinessProbe: &corev1.Probe{
									Handler: corev1.Handler{
										HTTPGet: &corev1.HTTPGetAction{
											Path:   "/ready",
											Port:   intstr.FromInt(8181),
											Scheme: corev1.URISchemeHTTP,
										},
									},
								},
							},
						},
						DNSPolicy: corev1.DNSDefault,
						Volumes: []corev1.Volume{
							{
								Name: "config-volume",
								VolumeSource: corev1.VolumeSource{
									ConfigMap: &corev1.ConfigMapVolumeSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "coredns",
										},
										Items: []corev1.KeyToPath{
											{
												Key:  "Corefile",
												Path: "Corefile",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (cluster *Cluster) reconcileCoreDNSService(client clientset.Interface) error {
	coreDNSServiceIP, err := cluster.CoreDNSServiceIP()
	if err != nil {
		return err
	}
	client.CoreV1().Services(metav1.NamespaceSystem).Create(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-dns",
				Namespace: metav1.NamespaceSystem,
				Annotations: map[string]string{
					"prometheus.io/port":   "9153",
					"prometheus.io/scrape": "true",
				},
				Labels: map[string]string{
					"k8s-app":                       "kube-dns",
					"kubernetes.io/cluster-service": "true",
					"kubernetes.io/name":            "CoreDNS",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"k8s-app": "kube-dns",
				},
				ClusterIP: coreDNSServiceIP,
				Ports: []corev1.ServicePort{
					{
						Name:     "dns",
						Protocol: corev1.ProtocolUDP,
						Port:     53,
					},
					{
						Name:     "dns-tcp",
						Protocol: corev1.ProtocolTCP,
						Port:     53,
					},
					{
						Name:     "metrics",
						Protocol: corev1.ProtocolTCP,
						Port:     9153,
					},
				},
			},
		},
	)
	return nil
}
