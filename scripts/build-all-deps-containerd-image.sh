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

set -ex

KUBERNETES_VERSION=${KUBERNETES_VERSION:-1.17.0}
CONTAINERD_LOCAL_ENDPOINT="unix:///containerd-socket/containerd.sock"

docker pull oneinfra/containerd:latest
CONTAINER_ID=$(docker run --privileged -d -it oneinfra/containerd:latest)
IMAGES=(oneinfra/haproxy:latest
        oneinfra/tooling:latest
        oneinfra/etcd:3.4.3
        k8s.gcr.io/pause:3.1
        k8s.gcr.io/kube-apiserver:v${KUBERNETES_VERSION}
        k8s.gcr.io/kube-controller-manager:v${KUBERNETES_VERSION}
        k8s.gcr.io/kube-scheduler:v${KUBERNETES_VERSION})
for image in "${IMAGES[@]}"; do
    docker exec -e IMAGE_SERVICE_ENDPOINT="${CONTAINERD_LOCAL_ENDPOINT}" -it "${CONTAINER_ID}" crictl pull "${image}"
done
docker commit "${CONTAINER_ID}" oneinfra/hypervisor:latest
docker rm -f "${CONTAINER_ID}"
