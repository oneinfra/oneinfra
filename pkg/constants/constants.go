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

package constants

import (
	"path/filepath"
)

const (
	// DefaultCAKeyBitSize is the default key bit size for CA certificates
	DefaultCAKeyBitSize = 4096
	// DefaultKeyBitSize is the default key bit size for non-CA certificates
	DefaultKeyBitSize = 2048
	// DefaultClusterCIDR is the default cluster CIDR
	DefaultClusterCIDR = "10.244.0.0/16"
	// DefaultServiceCIDR is the default service CIDR
	DefaultServiceCIDR = "10.96.0.0/12"
	// OneInfraNamespace is the namespace for storing OneInfra resources
	OneInfraNamespace = "oneinfra-system"
	// OneInfraVersionConfigMap is the oneinfra version ConfigMap
	OneInfraVersionConfigMap = "oneinfra-version"
	// OneInfraVersionsKeyName is the oneinfra version ConfigMap key name for the version structure
	OneInfraVersionsKeyName = "versions"
	// OneInfraJoinConfigMap is the name of the ConfigMap used to
	// store join information
	OneInfraJoinConfigMap = "oneinfra-join"
	// OneInfraJoinConfigMapJoinKey is the name of the key that holds
	// the join key inside the join ConfigMap
	OneInfraJoinConfigMapJoinKey = "joinKey"
	// OneInfraNodeJoinTokenExtraGroups represents the bootstrap token
	// extra groups used to identify oneinfra bootstrap tokens
	OneInfraNodeJoinTokenExtraGroups = "system:bootstrappers:oneinfra"
	// OneInfraKubeletProxierExtraGroups represents the kubelet proxier
	// extra groups used to identify kubelet proxying requests
	OneInfraKubeletProxierExtraGroups = "oneinfra:kubelet-proxier"
	// OneInfraConfigDir represents the configuration directory for oneinfra
	OneInfraConfigDir = "/etc/oneinfra"
	// OneInfraClusterNameLabelName is the name of the label for the
	// cluster name
	OneInfraClusterNameLabelName = "oneinfra/cluster-name"
	// OneInfraClusterUninitializedCertificates is the name of the label
	// when a cluster needs certificates or keys to be initialized
	OneInfraClusterUninitializedCertificates = "oneinfra/uninitialized-certificates"
	// OneInfraControlPlaneIngressVPNPeerName represents the control
	// plane ingress peer VPN name
	OneInfraControlPlaneIngressVPNPeerName = "control-plane-ingress"
	// OneInfraCleanupFinalizer is a finalizer for cleaning up resources
	OneInfraCleanupFinalizer = "oneinfra/cleanup"
	// KubeletDir is the kubelet configuration dir
	KubeletDir = "/var/lib/kubelet"
)

var (
	// KubeletKubeConfigPath represents the kubelet kubeconfig path
	KubeletKubeConfigPath = filepath.Join(OneInfraConfigDir, "kubelet.conf")
	// KubeletServerCertificatePath represents the kubelet server certificate path
	KubeletServerCertificatePath = filepath.Join(OneInfraConfigDir, "kubelet.crt")
	// KubeletServerPrivateKeyPath represents the kubelet server private key path
	KubeletServerPrivateKeyPath = filepath.Join(OneInfraConfigDir, "kubelet.key")
	// KubeletClientCACertificatePath represents the kubelet server certificate path
	KubeletClientCACertificatePath = filepath.Join(OneInfraConfigDir, "kubelet-client-ca.crt")
	// KubeletConfigPath represents the kubelet configuration path
	KubeletConfigPath = filepath.Join(KubeletDir, "config.yaml")
)
