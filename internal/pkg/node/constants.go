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

package node

const (
	systemdDir = "/etc/systemd/system"

	kubeletInstallerImage = "oneinfra/kubelet-installer:1.17.4"

	kubeletSystemdServiceTemplate = `[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/home/

[Service]
Environment="KUBELET_ARGS=--hostname-override={{.Nodename}}"
Environment="KUBELET_KUBECONFIG_ARGS=--kubeconfig={{.KubeletKubeConfigPath}}"
Environment="KUBELET_CONFIG_ARGS=--config={{.KubeletConfigPath}}"
Environment="SERVICE_ENDPOINTS_ARGS=--container-runtime=remote --image-service-endpoint={{.ImageServiceEndpoint}} --container-runtime-endpoint={{.ContainerRuntimeEndpoint}}"
ExecStart=/usr/local/bin/kubelet $KUBELET_ARGS $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $SERVICE_ENDPOINTS_ARGS
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`
)
