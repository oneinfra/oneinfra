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

set -e

export PATH=${GOPATH}/bin:./bin:${PATH}

CLUSTER_NAME=${CLUSTER_NAME:-cluster}

if [ -z "${CLUSTER_CONF}" ]; then
    echo 'Please, set $CLUSTER_CONF environment variable pointing to your cluster manifests'
    exit 1
fi

NETWORK_ARG=""
if [ -n "${NETWORK_NAME}" ]; then
    if ! docker network inspect "${NETWORK_NAME}" &> /dev/null; then
        echo "Network name ${NETWORK_NAME} not found, please set an existing network name or leave it blank"
        exit 1
    fi
else
    if docker network inspect kind &> /dev/null; then
        NETWORK_NAME="kind"
    fi
fi

if [ -n "${NETWORK_NAME}" ]; then
    NETWORK_ARG="--network=${NETWORK_NAME}"
fi

OI_BIN=$(which oi)
CONTAINERD_LOCAL_ENDPOINT="unix:///containerd-socket/containerd.sock"
APISERVER_ENDPOINT=$(cat ${CLUSTER_CONF} | oi-local-hypervisor-set endpoint --cluster "${CLUSTER_NAME}")
KUBERNETES_VERSION=$(cat ${CLUSTER_CONF} | oi cluster version --cluster "${CLUSTER_NAME}" kubernetes)
CONTAINERD_VERSION=$(oi version component --component containerd --kubernetes-version ${KUBERNETES_VERSION})
CONTAINER_ID=$(docker run --privileged ${NETWORK_ARG} -v /dev/null:/proc/swaps:ro -v /etc/resolv.conf:/etc/resolv.conf:ro -v ${OI_BIN}:/usr/local/bin/oi:ro -v $(realpath "${CLUSTER_CONF}"):/etc/oneinfra/cluster.conf:ro -d oneinfra/containerd:${CONTAINERD_VERSION})
KUBECONFIG=/tmp/kubeconfig-${CLUSTER_NAME}.conf
cat ${CLUSTER_CONF} | oi cluster admin-kubeconfig --cluster "${CLUSTER_NAME}" > ${KUBECONFIG}

docker exec ${CONTAINER_ID} sh -c "rm /etc/cni/net.d/*"

echo "creating new join token"
JOIN_TOKEN=$(cat ${CLUSTER_CONF} | oi join-token inject --cluster "${CLUSTER_NAME}" 3> "${CLUSTER_CONF}.new" 2>&1 >&3 | tr -d '\n')
NODENAME=$(echo ${CONTAINER_ID} | head -c 10)

echo "reconciling join tokens"
cat "${CLUSTER_CONF}.new" | oi reconcile > ${CLUSTER_CONF} 2> /dev/null

# Get the IP address of the fake worker, so we can request a SAN for
# its kubelet server certificate
FAKE_WORKER_IP_SAN=$(docker inspect ${CONTAINER_ID} | jq -r '.[0].NetworkSettings.IPAddress')
if [ -z "${FAKE_WORKER_IP_SAN}" ]; then
    FAKE_WORKER_IP_SAN=$(docker inspect ${CONTAINER_ID} | jq -r ".[0].NetworkSettings.Networks.${NETWORK_NAME}.IPAddress")
fi

echo "joining new node in background"
docker exec ${CONTAINER_ID} sh -c "cat /etc/oneinfra/cluster.conf | oi cluster apiserver-ca --cluster \"${CLUSTER_NAME}\" > /etc/oneinfra/apiserver-ca.crt"
docker exec ${CONTAINER_ID} sh -c "oi node join --nodename ${NODENAME} --extra-san ${FAKE_WORKER_IP_SAN} --apiserver-endpoint ${APISERVER_ENDPOINT} --apiserver-ca-cert-file /etc/oneinfra/apiserver-ca.crt --container-runtime-endpoint ${CONTAINERD_LOCAL_ENDPOINT} --image-service-endpoint ${CONTAINERD_LOCAL_ENDPOINT} --join-token ${JOIN_TOKEN}" &

echo -n "waiting for node join request to be created by the new node"
until kubectl --kubeconfig=${KUBECONFIG} get njr ${NODENAME} -n oneinfra-system &> /dev/null
do
    echo -n "."
    sleep 1
done
echo

echo "reconciling node join requests"
cat ${CLUSTER_CONF} | oi reconcile > "${CLUSTER_CONF}.new" 2> /dev/null
mv "${CLUSTER_CONF}.new" ${CLUSTER_CONF}
