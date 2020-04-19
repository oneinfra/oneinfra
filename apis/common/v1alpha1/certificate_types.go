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

// Certificate represents a Certificate
type Certificate struct {
	// Certificate is a PEM encoded certificate
	Certificate string `json:"certificate,omitempty"`

	// PrivateKey is a PEM encoded private key
	PrivateKey string `json:"privateKey,omitempty"`
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	// PublicKey is a PEM encoded public key
	PublicKey string `json:"publicKey,omitempty"`

	// PrivateKey is a PEM encoded private key
	PrivateKey string `json:"privateKey,omitempty"`
}
