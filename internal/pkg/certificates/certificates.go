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

package certificates

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"net"
	"time"

	"k8s.io/klog"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

// Certificate represents a certificate
type Certificate struct {
	Certificate string
	PrivateKey  string
	certificate *x509.Certificate
	privateKey  *rsa.PrivateKey
}

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
	pem.Encode(publicKeyPEM, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})
	privateKeyPEM := new(bytes.Buffer)
	pem.Encode(privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
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
	res, err := NewKeyPairFromString(keyPair.PrivateKey)
	if err != nil {
		return nil, err
	}
	return res, nil
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

// Export exports the key pair to a versioned key pair
func (keyPair *KeyPair) Export() *commonv1alpha1.KeyPair {
	return &commonv1alpha1.KeyPair{
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
	}
}

// Decrypt decrypts the given base-64 contents using this private key
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

// NewCertificateAuthority creates a new certificate authority
func NewCertificateAuthority(authorityName string) (*Certificate, error) {
	privateKey, err := NewPrivateKey(constants.DefaultKeyBitSize)
	if err != nil {
		return nil, err
	}
	serialNumber, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return nil, err
	}
	caCertificate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:    authorityName,
			Organization:  []string{"Some Company"},
			Country:       []string{"Some Country"},
			Province:      []string{"Some Province"},
			Locality:      []string{"Some Locality"},
			StreetAddress: []string{"Some StreetAddress"},
			PostalCode:    []string{"Some PostalCode"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caCertificateBytes, err := x509.CreateCertificate(rand.Reader, &caCertificate, &caCertificate, &privateKey.key.PublicKey, privateKey.key)
	if err != nil {
		return nil, err
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificateBytes,
	})
	return &Certificate{
		Certificate: caPEM.String(),
		PrivateKey:  privateKey.PrivateKey,
		certificate: &caCertificate,
		privateKey:  privateKey.key,
	}, nil
}

// NewCertificateFromv1alpha1 returns a certificate from a versioned certificate
func NewCertificateFromv1alpha1(certificate *commonv1alpha1.Certificate) *Certificate {
	res := &Certificate{
		Certificate: certificate.Certificate,
		PrivateKey:  certificate.PrivateKey,
	}
	if err := res.init(); err != nil {
		klog.Warningf("error when decoding certificate authority: %v", err)
	}
	return res
}

// Export exports the certificate to a versioned certificate
func (certificate *Certificate) Export() *commonv1alpha1.Certificate {
	return &commonv1alpha1.Certificate{
		Certificate: certificate.Certificate,
		PrivateKey:  certificate.PrivateKey,
	}
}

func (certificate *Certificate) init() error {
	if certificate.certificate != nil && certificate.privateKey != nil {
		return nil
	}
	block, _ := pem.Decode([]byte(certificate.Certificate))
	if block == nil {
		return errors.New("could not decode PEM encoded certificate")
	}
	parsedCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	certificate.certificate = parsedCertificate
	if len(certificate.PrivateKey) > 0 {
		block, _ = pem.Decode([]byte(certificate.PrivateKey))
		if block == nil {
			return errors.New("could not decode PEM private key")
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}
		certificate.privateKey = privateKey
	}
	return nil
}

// CreateCertificate generates a new certificate and key signed with the current CA
func (certificate *Certificate) CreateCertificate(commonName string, organization []string, extraSANs []string) (string, string, error) {
	serialNumber, err := rand.Int(rand.Reader, (&big.Int{}).Exp(big.NewInt(2), big.NewInt(159), nil))
	if err != nil {
		return "", "", err
	}
	sansHosts := []string{"localhost"}
	sansIps := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	for _, extraSAN := range extraSANs {
		if ip := net.ParseIP(extraSAN); ip != nil {
			sansIps = append(sansIps, ip)
		} else {
			sansHosts = append(sansHosts, extraSAN)
		}
	}
	newCertificate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:    commonName,
			Organization:  organization,
			Country:       []string{"Some Country"},
			Province:      []string{"Some Province"},
			Locality:      []string{"Some Locality"},
			StreetAddress: []string{"Some StreetAddress"},
			PostalCode:    []string{"Some PostalCode"},
		},
		DNSNames:     sansHosts,
		IPAddresses:  sansIps,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certificateBytes, err := x509.CreateCertificate(rand.Reader, &newCertificate, certificate.certificate, &certificate.privateKey.PublicKey, certificate.privateKey)
	if err != nil {
		return "", "", err
	}
	certificatePEM := new(bytes.Buffer)
	pem.Encode(certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateBytes,
	})
	certificatePrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certificatePrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certificate.privateKey),
	})
	return certificatePEM.String(), certificatePrivKeyPEM.String(), nil
}
