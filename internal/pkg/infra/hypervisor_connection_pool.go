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

package infra

// HypervisorConnectionPool represents an hypervisor connection pool
type HypervisorConnectionPool map[string]hypervisorEndpoint

// connection returns a gRPC client connection for the given
// hypervisor name; if a connection does not exist, it will be
// established using the provided hypervisor endpoint
func (connectionPool HypervisorConnectionPool) connection(hypervisorName string, hypervisorEndpoint hypervisorEndpoint) hypervisorEndpoint {
	if hypervisorConnection, exists := connectionPool[hypervisorName]; exists {
		return hypervisorConnection
	}
	connectionPool[hypervisorName] = hypervisorEndpoint
	return hypervisorEndpoint
}
