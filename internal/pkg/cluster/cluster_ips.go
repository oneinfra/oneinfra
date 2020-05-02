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

package cluster

import (
	"net"
)

const (
	kubernetesServiceIPOffset = 1
	coreDNSServiceIPOffset    = 10
)

// KubernetesServiceIP returns the Kubernetes IP inside the services
// CIDR
func (cluster *Cluster) KubernetesServiceIP() (string, error) {
	_, kubernetesServiceIP, err := net.ParseCIDR(cluster.ServiceCIDR)
	if err != nil {
		return "", err
	}
	kubernetesServiceIP.IP[len(kubernetesServiceIP.IP)-1] = kubernetesServiceIP.IP[len(kubernetesServiceIP.IP)-1] + kubernetesServiceIPOffset
	return kubernetesServiceIP.IP.String(), nil
}

// CoreDNSServiceIP returns the CoreDNS IP inside the services CIDR
func (cluster *Cluster) CoreDNSServiceIP() (string, error) {
	_, coreDNSServiceIP, err := net.ParseCIDR(cluster.ServiceCIDR)
	if err != nil {
		return "", err
	}
	coreDNSServiceIP.IP[len(coreDNSServiceIP.IP)-1] = coreDNSServiceIP.IP[len(coreDNSServiceIP.IP)-1] + coreDNSServiceIPOffset
	return coreDNSServiceIP.IP.String(), nil
}
