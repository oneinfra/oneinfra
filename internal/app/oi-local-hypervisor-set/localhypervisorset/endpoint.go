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

package localhypervisorset

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"

	"github.com/pkg/errors"

	componentapi "github.com/oneinfra/oneinfra/internal/pkg/component"
	localhypervisorsetpkg "github.com/oneinfra/oneinfra/internal/pkg/local-hypervisor-set"
	"github.com/oneinfra/oneinfra/internal/pkg/manifests"
)

// Endpoint prints the provided cluster endpoint
func Endpoint(clusterName string) error {
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	clusters := manifests.RetrieveClusters(string(stdin))
	components := manifests.RetrieveComponents(string(stdin))

	cluster, exists := clusters[clusterName]
	if !exists {
		return errors.Errorf("cluster %q not found", clusterName)
	}

	uri, err := url.Parse(cluster.APIServerEndpoint)
	if err != nil {
		return errors.Errorf("could not parse API server endpoint %q", cluster.APIServerEndpoint)
	}
	_, apiServerPort, err := net.SplitHostPort(uri.Host)
	if err != nil {
		return errors.Errorf("could not split host and port in %q", uri.Host)
	}
	for _, component := range components {
		if component.ClusterName == clusterName && component.Role == componentapi.ControlPlaneIngressRole {
			internalHypervisorIP, err := localhypervisorsetpkg.InternalIPAddress(component.HypervisorName)
			if err != nil {
				return errors.Errorf("could not retrieve the internal IP address for hypervisor %q", component.HypervisorName)
			}
			uri := url.URL{Scheme: uri.Scheme, Host: net.JoinHostPort(internalHypervisorIP, apiServerPort)}
			fmt.Println(uri.String())
			return nil
		}
	}

	return errors.Errorf("could not find ingress component for cluster %q", clusterName)
}
