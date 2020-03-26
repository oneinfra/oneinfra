#!/usr/bin/env bash

# Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

kubernetes_version() {
    if [ "${KUBERNETES_VERSION}" = "default" ]; then
        ${SCRIPT_DIR}/run-local.sh oi version kubernetes | tail -n1
    else
        echo "${KUBERNETES_VERSION}"
    fi
}

cri_tools_version() {
    ${SCRIPT_DIR}/run-local.sh oi version component --kubernetes-version $(kubernetes_version) --component cri-tools
}

install_kubectl() {
    sudo wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v$(kubernetes_version)/bin/linux/amd64/kubectl
    sudo chmod +x /usr/local/bin/kubectl
}

install_crictl() {
    wget -O cri-tools.tar.gz https://github.com/kubernetes-sigs/cri-tools/releases/download/v$(cri_tools_version)/crictl-v$(cri_tools_version)-linux-amd64.tar.gz
    sudo tar -C /usr/local/bin -xf cri-tools.tar.gz
    rm cri-tools.tar.gz
}

case "$1" in
    pull-hypervisor)
        docker pull oneinfra/hypervisor:$(kubernetes_version)
        shift
        ;;
    kubectl)
        install_kubectl
        shift
        ;;
    crictl)
        install_crictl
        shift
        ;;
    *)
        echo "unknown argument: ${1}"
        exit 1
        ;;
esac
