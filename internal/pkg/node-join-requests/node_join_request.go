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

package nodejoinrequests

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
)

// Condition represents a node join request condition
type Condition string

// ConditionList represents a node join request condition list
type ConditionList []Condition

const (
	// Issued represents a join request that has been completed
	Issued Condition = "issued"
)

// NodeJoinRequest represents a node join request
type NodeJoinRequest struct {
	Name                     string
	SymmetricKey             string
	APIServerEndpoint        string
	ContainerRuntimeEndpoint string
	ImageServiceEndpoint     string
	VPNAddress               string
	VPNPeer                  string
	KubeConfig               string
	KubeletConfig            string
	Conditions               ConditionList
	ResourceVersion          string
}

// NewNodeJoinRequestFromv1alpha1 returns a node join request based on a versioned node join request
func NewNodeJoinRequestFromv1alpha1(nodeJoinRequest *nodev1alpha1.NodeJoinRequest, joinKey *certificates.KeyPair) (*NodeJoinRequest, error) {
	symmetricKey := ""
	if joinKey != nil {
		var err error
		symmetricKey, err = joinKey.Decrypt(nodeJoinRequest.Spec.SymmetricKey)
		if err != nil {
			return nil, err
		}
	}
	return &NodeJoinRequest{
		Name:                     nodeJoinRequest.ObjectMeta.Name,
		SymmetricKey:             symmetricKey,
		APIServerEndpoint:        nodeJoinRequest.Spec.APIServerEndpoint,
		ContainerRuntimeEndpoint: nodeJoinRequest.Spec.ContainerRuntimeEndpoint,
		ImageServiceEndpoint:     nodeJoinRequest.Spec.ImageServiceEndpoint,
		VPNAddress:               nodeJoinRequest.Status.VPNAddress,
		VPNPeer:                  nodeJoinRequest.Status.VPNPeer,
		KubeConfig:               nodeJoinRequest.Status.KubeConfig,
		KubeletConfig:            nodeJoinRequest.Status.KubeletConfig,
		Conditions:               newConditionsFromv1alpha1(nodeJoinRequest.Status.Conditions),
		ResourceVersion:          nodeJoinRequest.ObjectMeta.ResourceVersion,
	}, nil
}

func newConditionsFromv1alpha1(conditions []nodev1alpha1.Condition) ConditionList {
	res := ConditionList{}
	for _, condition := range conditions {
		switch condition {
		case nodev1alpha1.Issued:
			res = append(res, Issued)
		}
	}
	return res
}

// Export exports this node join request to a versioned node join request
func (nodeJoinRequest *NodeJoinRequest) Export() *nodev1alpha1.NodeJoinRequest {
	return &nodev1alpha1.NodeJoinRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nodeJoinRequest.Name,
			ResourceVersion: nodeJoinRequest.ResourceVersion,
		},
		Spec: nodev1alpha1.NodeJoinRequestSpec{
			SymmetricKey:             nodeJoinRequest.SymmetricKey,
			APIServerEndpoint:        nodeJoinRequest.APIServerEndpoint,
			ContainerRuntimeEndpoint: nodeJoinRequest.ContainerRuntimeEndpoint,
			ImageServiceEndpoint:     nodeJoinRequest.ImageServiceEndpoint,
		},
		Status: nodev1alpha1.NodeJoinRequestStatus{
			VPNAddress:    nodeJoinRequest.VPNAddress,
			VPNPeer:       nodeJoinRequest.VPNPeer,
			KubeConfig:    nodeJoinRequest.KubeConfig,
			KubeletConfig: nodeJoinRequest.KubeletConfig,
			Conditions:    nodeJoinRequest.Conditions.export(),
		},
	}
}

func (conditionList ConditionList) export() []nodev1alpha1.Condition {
	res := []nodev1alpha1.Condition{}
	for _, condition := range conditionList {
		switch condition {
		case Issued:
			res = append(res, nodev1alpha1.Issued)
		}
	}
	return res
}

// HasCondition returns whether this node join request has a given condition
func (nodeJoinRequest *NodeJoinRequest) HasCondition(condition Condition) bool {
	for _, nodeJoinRequestCondition := range nodeJoinRequest.Conditions {
		if nodeJoinRequestCondition == condition {
			return true
		}
	}
	return false
}

// Encrypt encrypts the given content using this node join request symmetric key
func (nodeJoinRequest *NodeJoinRequest) Encrypt(content string) (string, error) {
	block, err := aes.NewCipher([]byte(nodeJoinRequest.SymmetricKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(content), nil)), nil
}
