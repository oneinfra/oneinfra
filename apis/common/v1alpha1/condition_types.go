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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ConditionType represents a condition type
type ConditionType string

// ConditionStatus represents a condition status
type ConditionStatus string

const (
	// ConditionTrue is a condition that holds true
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse is a condition that holds false
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown is a condition in an unknown state
	ConditionUnknown ConditionStatus = "Unknown"
)

// +kubebuilder:object:root=true

// Condition represents a condition
type Condition struct {
	Type               ConditionType   `json:"type,omitempty"`
	Status             ConditionStatus `json:"status,omitempty"`
	LastTransitionTime metav1.Time     `json:"lastTransitionTime,omitempty"`
	LastSetTime        metav1.Time     `json:"lastSetTime,omitempty"`
	Reason             string          `json:"reason,omitempty"`
	Message            string          `json:"message,omitempty"`
}

// +kubebuilder:object:root=true

// ConditionList represents a list of conditions
type ConditionList []Condition

// IsCondition checks whether a condition type matches a condition
// status
func (conditionList ConditionList) IsCondition(conditionType ConditionType, conditionStatus ConditionStatus) bool {
	for _, condition := range conditionList {
		if condition.Type == conditionType {
			return condition.Status == conditionStatus
		}
	}
	return false
}

// GetObjectKind partially implements the runtime.Object interface for
// Condition
func (condition *Condition) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

// GetObjectKind partially implements the runtime.Object interface for
// ConditionList
func (conditionList ConditionList) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}
