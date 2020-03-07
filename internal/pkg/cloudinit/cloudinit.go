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

package cloudinit

import (
	"strings"

	"github.com/pkg/errors"

	"sigs.k8s.io/yaml"
)

// CloudConfig represents a cloud-config file
type CloudConfig struct {
	CACerts    CACerts    `json:"ca-certs,omitempty"`
	WriteFiles WriteFiles `json:"write_files,omitempty"`
	RunCmd     []string   `json:"runcmd,omitempty"`
	BootCmd    []string   `json:"bootcmd,omitempty"`
}

// CACerts reprsents cloud-config's ca-certs
type CACerts struct {
	Trusted []string `json:"trusted,omitempty"`
}

// WriteFiles represents cloud-config's write_files
type WriteFiles struct {
	Encoding    string `json:"encoding,omitempty"`
	Content     string `json:"content,omitempty"`
	Path        string `json:"path,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Owner       string `json:"owner,omitempty"`
}

// Export exports the cloud-config contents
func (cloudConfig *CloudConfig) Export() (string, error) {
	cloudConfigContents, err := yaml.Marshal(cloudConfig)
	if err != nil {
		return "", errors.Wrap(err, "could not marshal cloud-config file")
	}
	return strings.Join([]string{"#cloud-config", string(cloudConfigContents)}, "\n"), nil
}
