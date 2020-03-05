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

package infra

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"

	commonv1alpha1 "github.com/oneinfra/oneinfra/apis/common/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
)

type hypervisorEndpoint interface {
	Connection() (*grpc.ClientConn, error)
	Export() (*infrav1alpha1.LocalHypervisorCRIEndpoint, *infrav1alpha1.RemoteHypervisorCRIEndpoint)
}

type localHypervisorEndpoint struct {
	CRIEndpoint string
}

type remoteHypervisorEndpoint struct {
	CRIEndpoint       string
	CACertificate     *certificates.Certificate
	ClientCertificate *certificates.Certificate
}

func (endpoint *localHypervisorEndpoint) Connection() (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("passthrough:///unix://%s", endpoint.CRIEndpoint), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
}

func (endpoint *localHypervisorEndpoint) Export() (*infrav1alpha1.LocalHypervisorCRIEndpoint, *infrav1alpha1.RemoteHypervisorCRIEndpoint) {
	return &infrav1alpha1.LocalHypervisorCRIEndpoint{
		CRIEndpoint: endpoint.CRIEndpoint,
	}, nil
}

func (endpoint *remoteHypervisorEndpoint) Connection() (*grpc.ClientConn, error) {
	clientCert, err := tls.X509KeyPair(
		[]byte(endpoint.ClientCertificate.Certificate),
		[]byte(endpoint.ClientCertificate.PrivateKey),
	)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM([]byte(endpoint.CACertificate.Certificate)); !ok {
		return nil, errors.New("could not add CA certificate to the pool of known certificates")
	}
	transportCredentials := grpc.WithTransportCredentials(
		grpccredentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}),
	)
	return grpc.Dial(endpoint.CRIEndpoint, transportCredentials, grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
}

func (endpoint *remoteHypervisorEndpoint) Export() (*infrav1alpha1.LocalHypervisorCRIEndpoint, *infrav1alpha1.RemoteHypervisorCRIEndpoint) {
	return nil, &infrav1alpha1.RemoteHypervisorCRIEndpoint{
		CRIEndpoint:       endpoint.CRIEndpoint,
		CACertificate:     endpoint.CACertificate.Certificate,
		ClientCertificate: endpoint.ClientCertificate.Export(),
	}
}

func setHypervisorEndpointFromv1alpha1(hypervisor *infrav1alpha1.Hypervisor, resHypervisor *Hypervisor) error {
	if hypervisor.Spec.LocalCRIEndpoint != nil && hypervisor.Spec.RemoteCRIEndpoint != nil {
		return errors.Errorf("hypervisor %q has both a local and a remote CRI endpoint, can only have one", hypervisor.ObjectMeta.Name)
	} else if hypervisor.Spec.LocalCRIEndpoint != nil {
		resHypervisor.Endpoint = &localHypervisorEndpoint{
			CRIEndpoint: hypervisor.Spec.LocalCRIEndpoint.CRIEndpoint,
		}
	} else if hypervisor.Spec.RemoteCRIEndpoint != nil {
		resHypervisor.Endpoint = &remoteHypervisorEndpoint{
			CRIEndpoint: hypervisor.Spec.RemoteCRIEndpoint.CRIEndpoint,
			CACertificate: certificates.NewCertificateFromv1alpha1(&commonv1alpha1.Certificate{
				Certificate: hypervisor.Spec.RemoteCRIEndpoint.CACertificate,
				PrivateKey:  "",
			}),
			ClientCertificate: certificates.NewCertificateFromv1alpha1(hypervisor.Spec.RemoteCRIEndpoint.ClientCertificate),
		}
	} else {
		return errors.Errorf("hypervisor %q is missing a local or a remote CRI endpoint", hypervisor.ObjectMeta.Name)
	}
	return nil
}
