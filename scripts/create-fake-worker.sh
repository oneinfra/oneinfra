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

set -e

export PATH=${GOPATH}/bin:./bin:${PATH}

if [ -z "${CLUSTER_CONF}" ]; then
    echo 'Please, set $CLUSTER_CONF environment variable pointing to your cluster manifests'
    exit 1
fi

if [ -z "${CLUSTER_NAME}" ]; then
    echo 'Please, set $CLUSTER_NAME environment variable setting the name of the cluster you want to join this fake worker to'
    exit 1
fi

# This is a bit hacky for this specific Docker based environment. On
# the real world, worker nodes will connect to the public facing
# interface of the public hypervisors, but since we are mimicking this
# in Docker as forwarding the ports to localhost, containers in their
# own networking namespaces (as fake workers) won't be able to access
# to forwarded ports. Thus, we will connect them to the Docker IP
# address of the public hypervisors, that in this model resembles the
# private interface that we would have in reality, so we are not
# having a perfect match with the intended real world production
# design.

INGRESS_CONTAINER_NAME="${CLUSTER_NAME}-$(cat "${CLUSTER_CONF}" | oi cluster ingress-node-name --cluster "${CLUSTER_NAME}")"
INGRESS_CONTAINER_IP=$(docker inspect -f '{{ .NetworkSettings.IPAddress }}' "${INGRESS_CONTAINER_NAME}")

KUBECONFIG=$(mktemp /tmp/kubeconfig-XXXXXXX)
cat "${CLUSTER_CONF}" | oi cluster kubeconfig --cluster "${CLUSTER_NAME}" --endpoint-host-override "${INGRESS_CONTAINER_IP}" > "${KUBECONFIG}"

CONTAINERD_LOCAL_ENDPOINT="unix:///containerd-socket/containerd.sock"

docker run --privileged -v /dev/null:/proc/swaps:ro -v "${KUBECONFIG}":/kubeconfig -e CONTAINER_RUNTIME_ENDPOINT=${CONTAINERD_LOCAL_ENDPOINT} -e IMAGE_SERVICE_ENDPOINT=${CONTAINERD_LOCAL_ENDPOINT} -d -i oneinfra/kubelet:latest
