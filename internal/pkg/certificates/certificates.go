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
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"time"

	"k8s.io/klog"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/internal/pkg/crypto"
)

// Certificate represents a certificate
type Certificate struct {
	Certificate string
	PrivateKey  string
	certificate *x509.Certificate
	privateKey  *rsa.PrivateKey
}

// NewCertificateAuthority creates a new certificate authority
func NewCertificateAuthority(authorityName string) (*Certificate, error) {
	privateKey, err := crypto.NewPrivateKey(constants.DefaultKeyBitSize)
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
	caCertificateBytes, err := x509.CreateCertificate(rand.Reader, &caCertificate, &caCertificate, &privateKey.Key().PublicKey, privateKey.Key())
	if err != nil {
		return nil, err
	}
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificateBytes,
	})
	if err != nil {
		return nil, err
	}
	return &Certificate{
		Certificate: caPEM.String(),
		PrivateKey:  privateKey.PrivateKey,
		certificate: &caCertificate,
		privateKey:  privateKey.Key(),
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
	knownSans := map[string]struct{}{
		"localhost":               {},
		"127.0.0.1":               {},
		net.IPv6loopback.String(): {},
	}
	for _, extraSAN := range extraSANs {
		if _, exists := knownSans[extraSAN]; exists {
			continue
		}
		if ip := net.ParseIP(extraSAN); ip != nil {
			sansIps = append(sansIps, ip)
		} else {
			sansHosts = append(sansHosts, extraSAN)
		}
		knownSans[extraSAN] = struct{}{}
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
	err = pem.Encode(certificatePEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateBytes,
	})
	if err != nil {
		return "", "", err
	}
	certificatePrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certificatePrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certificate.privateKey),
	})
	if err != nil {
		return "", "", err
	}
	return certificatePEM.String(), certificatePrivKeyPEM.String(), nil
}
