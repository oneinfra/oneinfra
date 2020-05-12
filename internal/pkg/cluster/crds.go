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
	"context"
	"fmt"

	"github.com/pkg/errors"

	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionapi "k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/yaml"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
)

// ReconcileCustomResourceDefinitions reconciles this cluster custom resource definitions
func (cluster *Cluster) ReconcileCustomResourceDefinitions() error {
	client, err := cluster.KubernetesExtensionsClient()
	if err != nil {
		return err
	}
	version, err := versionapi.ParseSemantic(cluster.KubernetesVersion)
	if err != nil {
		return errors.Wrapf(err, "could not parse version %q", cluster.KubernetesVersion)
	}
	versionCompare, err := version.Compare("1.16.0")
	if err != nil {
		return err
	}
	if versionCompare < 0 {
		return cluster.reconcileNodeJoinRequestsCRDLegacy(client)
	}
	return cluster.reconcileNodeJoinRequestsCRD(client)
}

func (cluster *Cluster) reconcileNodeJoinRequestsCRD(client apiextensionsclientset.Interface) error {
	openAPISchema := extensionsv1.JSONSchemaProps{}
	if err := yaml.Unmarshal([]byte(nodev1alpha1.NodeJoinRequestOpenAPISchema), &openAPISchema); err != nil {
		return err
	}
	_, err := client.ApiextensionsV1().CustomResourceDefinitions().Create(
		context.TODO(),
		&extensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("nodejoinrequests.%s", nodev1alpha1.GroupVersion.Group),
			},
			Spec: extensionsv1.CustomResourceDefinitionSpec{
				Group: nodev1alpha1.GroupVersion.Group,
				Names: extensionsv1.CustomResourceDefinitionNames{
					Plural:     "nodejoinrequests",
					Singular:   "nodejoinrequest",
					ShortNames: []string{"njr", "njrs"},
					Kind:       "NodeJoinRequest",
					ListKind:   "NodeJoinRequestList",
				},
				Scope: extensionsv1.ClusterScoped,
				Versions: []extensionsv1.CustomResourceDefinitionVersion{
					{
						Name:    nodev1alpha1.GroupVersion.Version,
						Served:  true,
						Storage: true,
						Schema: &extensionsv1.CustomResourceValidation{
							OpenAPIV3Schema: &openAPISchema,
						},
						Subresources: &extensionsv1.CustomResourceSubresources{
							Status: &extensionsv1.CustomResourceSubresourceStatus{},
						},
						AdditionalPrinterColumns: []extensionsv1.CustomResourceColumnDefinition{
							{
								Name:     "SANs",
								Type:     "string",
								JSONPath: ".spec.extraSANs",
							},
							{
								Name:     "Age",
								Type:     "date",
								JSONPath: ".metadata.creationTimestamp",
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

func (cluster *Cluster) reconcileNodeJoinRequestsCRDLegacy(client apiextensionsclientset.Interface) error {
	openAPISchema := extensionsv1beta1.JSONSchemaProps{}
	if err := yaml.Unmarshal([]byte(nodev1alpha1.NodeJoinRequestOpenAPISchema), &openAPISchema); err != nil {
		return err
	}
	falseVar := false
	_, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(
		context.TODO(),
		&extensionsv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("nodejoinrequests.%s", nodev1alpha1.GroupVersion.Group),
			},
			Spec: extensionsv1beta1.CustomResourceDefinitionSpec{
				Group:   nodev1alpha1.GroupVersion.Group,
				Version: nodev1alpha1.GroupVersion.Version,
				Names: extensionsv1beta1.CustomResourceDefinitionNames{
					Plural:     "nodejoinrequests",
					Singular:   "nodejoinrequest",
					ShortNames: []string{"njr", "njrs"},
					Kind:       "NodeJoinRequest",
					ListKind:   "NodeJoinRequestList",
				},
				Scope: extensionsv1beta1.ClusterScoped,
				Validation: &extensionsv1beta1.CustomResourceValidation{
					OpenAPIV3Schema: &openAPISchema,
				},
				Subresources: &extensionsv1beta1.CustomResourceSubresources{
					Status: &extensionsv1beta1.CustomResourceSubresourceStatus{},
				},
				AdditionalPrinterColumns: []extensionsv1beta1.CustomResourceColumnDefinition{
					{
						Name:     "SANs",
						Type:     "string",
						JSONPath: ".spec.extraSANs",
					},
					{
						Name:     "Age",
						Type:     "date",
						JSONPath: ".metadata.creationTimestamp",
					},
				},
				PreserveUnknownFields: &falseVar,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
