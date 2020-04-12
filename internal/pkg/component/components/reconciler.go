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

package components

import (
	"github.com/oneinfra/oneinfra/internal/pkg/inquirer"
)

// Component is an interface that allows a component implementing this
// interface to be reconciled
type Component interface {
	// PreReconcile allocates resources that need to be reserved for the
	// reconcile process. If conflicts arise on resources, the
	// PreReconcile will be retried, instead of a full Reconcile cycle
	PreReconcile(inquirer.ReconcilerInquirer) error
	Reconcile(inquirer.ReconcilerInquirer) error
	ReconcileDeletion(inquirer.ReconcilerInquirer) error
}
