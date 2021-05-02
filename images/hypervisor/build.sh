#!/usr/bin/env bash

# Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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

if [ -z "${CONTAINERD_VERSION}" ]; then
    echo 'Please, set $CONTAINERD_VERSION environment variable setting the desired containerd version to include'
    exit 1
fi

if [ -z "${ETCD_VERSION}" ]; then
    echo 'Please, set $ETCD_VERSION environment variable setting the desired etcd version to include'
    exit 1
fi

if [ -z "${PAUSE_VERSION}" ]; then
    echo 'Please, set $PAUSE_VERSION environment variable setting the desired pause image version to include'
    exit 1
fi

if [ -z "${KUBERNETES_VERSION}" ]; then
    echo 'Please, set $KUBERNETES_VERSION environment variable setting the desired kubernetes version to include'
    exit 1
fi

CONTAINER_ID=$(docker run --privileged -d -it oneinfra/containerd:${CONTAINERD_VERSION})
IMAGES=(oneinfra/haproxy:latest
        oneinfra/tooling:latest
        oneinfra/etcd:${ETCD_VERSION}
        k8s.gcr.io/pause:${PAUSE_VERSION}
        k8s.gcr.io/kube-apiserver:v${KUBERNETES_VERSION}
        k8s.gcr.io/kube-controller-manager:v${KUBERNETES_VERSION}
        k8s.gcr.io/kube-scheduler:v${KUBERNETES_VERSION})
for image in "${IMAGES[@]}"; do
    docker exec "${CONTAINER_ID}" crictl pull "${image}"
done

docker commit "${CONTAINER_ID}" oneinfra/hypervisor:${KUBERNETES_VERSION}
docker rm -f "${CONTAINER_ID}"
