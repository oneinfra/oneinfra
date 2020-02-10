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

package cluster

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

	clusterv1alpha1 "oneinfra.ereslibre.es/m/apis/cluster/v1alpha1"
)

type certificateAuthorities struct {
	apiServerClient   *certificateAuthority
	certificateSigner *certificateAuthority
	kubelet           *certificateAuthority
}

type certificateAuthority struct {
	caCertificateContents string
	caPrivateKeyContents  string
	caCertificate         *x509.Certificate
	caPrivateKey          *rsa.PrivateKey
}

func newCertificateAuthorities() (*certificateAuthorities, error) {
	apiserverClientAuthority, err := newCertificateAuthority()
	if err != nil {
		return nil, err
	}
	certificateSignerAuthority, err := newCertificateAuthority()
	if err != nil {
		return nil, err
	}
	kubeletAuthority, err := newCertificateAuthority()
	if err != nil {
		return nil, err
	}
	return &certificateAuthorities{
		apiServerClient:   apiserverClientAuthority,
		certificateSigner: certificateSignerAuthority,
		kubelet:           kubeletAuthority,
	}, nil
}

func newCertificateAuthority() (*certificateAuthority, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	caCertificate := x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
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
	caCertificateBytes, err := x509.CreateCertificate(rand.Reader, &caCertificate, &caCertificate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertificateBytes,
	})
	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	return &certificateAuthority{
		caCertificateContents: caPEM.String(),
		caPrivateKeyContents:  caPrivKeyPEM.String(),
		caCertificate:         &caCertificate,
		caPrivateKey:          privateKey,
	}, nil
}

func newCertificateAuthorityFromv1alpha1(ca *clusterv1alpha1.CertificateAuthority) *certificateAuthority {
	res := &certificateAuthority{
		caCertificateContents: ca.CACertificate,
		caPrivateKeyContents:  ca.CAPrivateKey,
	}
	if err := res.init(); err != nil {
		klog.Warningf("error when decoding certificate authority: %v", err)
	}
	return res
}

func (ca *certificateAuthority) init() error {
	if ca.caCertificate != nil && ca.caPrivateKey != nil {
		return nil
	}
	block, _ := pem.Decode([]byte(ca.caCertificateContents))
	if block == nil {
		return errors.New("could not decode PEM CA certificate")
	}
	caCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}
	ca.caCertificate = caCertificate
	block, _ = pem.Decode([]byte(ca.caPrivateKeyContents))
	if block == nil {
		return errors.New("could not decode PEM private key")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	ca.caPrivateKey = privateKey
	return nil
}

func (ca *certificateAuthority) createCertificate() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}
	certificate := x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{"Some Company"},
			Country:       []string{"Some Country"},
			Province:      []string{"Some Province"},
			Locality:      []string{"Some Locality"},
			StreetAddress: []string{"Some StreetAddress"},
			PostalCode:    []string{"Some PostalCode"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certificateBytes, err := x509.CreateCertificate(rand.Reader, &certificate, ca.caCertificate, &ca.caPrivateKey.PublicKey, ca.caPrivateKey)
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
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	return certificatePEM.String(), certificatePrivKeyPEM.String(), nil
}
