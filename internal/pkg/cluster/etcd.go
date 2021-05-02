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

package cluster

import (
	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	"github.com/oneinfra/oneinfra/internal/pkg/certificates"
)

// EtcdServer represents the etcd component
type EtcdServer struct {
	CA *certificates.Certificate
}

func newEtcdServer() (*EtcdServer, error) {
	certificateAuthority, err := certificates.NewCertificateAuthority("etcd-authority")
	if err != nil {
		return nil, err
	}
	return &EtcdServer{
		CA: certificateAuthority,
	}, nil
}

func newEtcdServerFromv1alpha1(etcdServer *clusterv1alpha1.EtcdServer) *EtcdServer {
	return &EtcdServer{
		CA: certificates.NewCertificateFromv1alpha1(etcdServer.CA),
	}
}

// Export exports this etcd server into a versioned etcd server
func (etcdServer *EtcdServer) Export() *clusterv1alpha1.EtcdServer {
	if etcdServer == nil {
		return nil
	}
	return &clusterv1alpha1.EtcdServer{
		CA: etcdServer.CA.Export(),
	}
}
