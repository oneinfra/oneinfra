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

package reconciler

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// ReconcileErrors represents a fine grained report of the series of
// errors that happened during a reconciliation cycle
type ReconcileErrors map[string][]error

// Error returns the errors that happened during this reconciliation
// cycle in string format
func (reconcileErrors ReconcileErrors) Error() string {
	allErrors := []string{}
	for clusterName, errors := range reconcileErrors {
		for _, error := range errors {
			allErrors = append(
				allErrors,
				fmt.Sprintf("%s[%v]", clusterName, error),
			)
		}
	}
	return strings.Join(allErrors, ", ")
}

// IsClusterErrorFree returns whether the cluster provided has at
// least one error
func (reconcileErrors ReconcileErrors) IsClusterErrorFree(clusterName string) bool {
	clusterErrors, exists := reconcileErrors[clusterName]
	if !exists {
		return true
	}
	return len(clusterErrors) == 0
}

func (reconcileErrors ReconcileErrors) addClusterError(clusterName string, err error) {
	reconcileErrors.ensureClusterEntry(clusterName)
	reconcileErrors[clusterName] = append(
		reconcileErrors[clusterName],
		err,
	)
}

func (reconcileErrors ReconcileErrors) addComponentError(clusterName, componentName string, err error) {
	reconcileErrors.ensureClusterEntry(clusterName)
	reconcileErrors[clusterName] = append(
		reconcileErrors[clusterName],
		errors.Wrap(err, componentName),
	)
}

func (reconcileErrors ReconcileErrors) ensureClusterEntry(clusterName string) {
	if reconcileErrors[clusterName] == nil {
		reconcileErrors[clusterName] = []error{}
	}
}
