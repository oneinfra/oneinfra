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
	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
)

const (
	// ReconcileStarted represents a condition type signaling whether a
	// reconcile has been started
	ReconcileStarted commonv1alpha1.ConditionType = "ReconcileStarted"

	// ReconcileSucceeded represents a condition type signaling that a
	// reconcile has succeeded
	ReconcileSucceeded commonv1alpha1.ConditionType = "ReconcileSucceeded"
)
