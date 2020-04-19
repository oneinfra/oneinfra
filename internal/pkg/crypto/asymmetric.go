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

package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
)

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  string
	PrivateKey string
	key        *rsa.PrivateKey
}

// PublicKey represents a public key
type PublicKey struct {
	PublicKey string
	key       *rsa.PublicKey
}

// NewPrivateKey generates a new key pair
func NewPrivateKey(keyBitSize int) (*KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyBitSize)
	if err != nil {
		return nil, err
	}
	publicKey, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := new(bytes.Buffer)
	err = pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})
	if err != nil {
		return nil, err
	}
	privateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return nil, err
	}
	return &KeyPair{
		PublicKey:  publicKeyPEM.String(),
		PrivateKey: privateKeyPEM.String(),
		key:        key,
	}, nil
}

// NewPublicKeyFromFile returns a public key from a PEM encoded public
// key file in the given path
func NewPublicKeyFromFile(publicKeyPEMPath string) (*PublicKey, error) {
	publicKeyPEM, err := ioutil.ReadFile(publicKeyPEMPath)
	if err != nil {
		return nil, err
	}
	return NewPublicKeyFromString(string(publicKeyPEM))
}

// NewPublicKeyFromString returns a public key from a PEM encoded
// public key
func NewPublicKeyFromString(publicKeyPEM string) (*PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("could not parse public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	if publicKey, ok := publicKey.(*rsa.PublicKey); ok {
		return &PublicKey{
			PublicKey: publicKeyPEM,
			key:       publicKey,
		}, nil
	}
	return nil, errors.New("could not identify public key as an RSA public key")
}

// NewKeyPairFromFile returns a key pair from a PEM encoded private
// key file in the given path
func NewKeyPairFromFile(privateKeyPEMPath string) (*KeyPair, error) {
	privateKeyPEM, err := ioutil.ReadFile(privateKeyPEMPath)
	if err != nil {
		return nil, err
	}
	return NewKeyPairFromString(string(privateKeyPEM))
}

// NewKeyPairFromString returns a key pair from a PEM encoded private
// key
func NewKeyPairFromString(privateKeyPEM string) (*KeyPair, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("could not parse private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	publicKeyPEM := new(bytes.Buffer)
	pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})
	return &KeyPair{
		PublicKey:  publicKeyPEM.String(),
		PrivateKey: string(privateKeyPEM),
		key:        privateKey,
	}, nil
}

// NewKeyPairFromv1alpha1 returns a key pair from a versioned key pair
func NewKeyPairFromv1alpha1(keyPair *commonv1alpha1.KeyPair) (*KeyPair, error) {
	if keyPair == nil {
		return nil, nil
	}
	res, err := NewKeyPairFromString(keyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Export exports the key pair to a versioned key pair
func (keyPair *KeyPair) Export() *commonv1alpha1.KeyPair {
	if keyPair == nil {
		return nil
	}
	return &commonv1alpha1.KeyPair{
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
	}
}

// Encrypt encrypts the given content using this public key,
// producing a base64 result
func (publicKey *PublicKey) Encrypt(content string) (string, error) {
	encryptedContents, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey.key, []byte(content), []byte(""))
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(encryptedContents), nil
}

// Encrypt encrypts the given base-64 contents using the public key in
// the key pair
func (keyPair *KeyPair) Encrypt(content string) (string, error) {
	encryptedContents, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &keyPair.key.PublicKey, []byte(content), []byte(""))
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(encryptedContents), nil
}

// Decrypt decrypts the given base-64 contents using the private key
// in the key pair
func (keyPair *KeyPair) Decrypt(content string) (string, error) {
	rawContent, err := base64.RawStdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}
	decryptedContents, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, keyPair.key, rawContent, []byte(""))
	if err != nil {
		return "", err
	}
	return string(decryptedContents), nil
}

// Key returns the RSA private key for this private key pair
func (keyPair *KeyPair) Key() *rsa.PrivateKey {
	return keyPair.key
}
