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

if [ -z "$CI" ]; then
    export PATH=${GOPATH}/bin:${PATH}
else
    export PATH=${PWD}/bin:${PATH}
fi

INFRA_TEST_CLUSTER_NAME="${INFRA_TEST_CLUSTER_NAME:-test}"
KUBERNETES_VERSION="${KUBERNETES_VERSION:-default}"
CLUSTER_CONF="${CLUSTER_CONF:-cluster.conf}"
CLUSTER_NAME="${CLUSTER_NAME:-cluster}"

mkdir -p ~/.kube

echo "Creating infrastructure"
oi-local-hypervisor-set create --name "${INFRA_TEST_CLUSTER_NAME}" --kubernetes-version "${KUBERNETES_VERSION}" "$@" > ${CLUSTER_CONF}
docker ps -a

RECONCILED_CLUSTER_CONF=$(mktemp /tmp/reconciled-cluster-${INFRA_TEST_CLUSTER_NAME}-XXXXXXX.conf)

echo "Reconciling infrastructure"
cat ${CLUSTER_CONF} | \
    oi cluster inject --name "${CLUSTER_NAME}" --kubernetes-version "${KUBERNETES_VERSION}" --control-plane-replicas 3 | \
    oi reconcile -v 2 > ${RECONCILED_CLUSTER_CONF}

mv ${RECONCILED_CLUSTER_CONF} ${CLUSTER_CONF}

cat ${CLUSTER_CONF} | oi cluster admin-kubeconfig --cluster "${CLUSTER_NAME}" > ~/.kube/config

# Tests

echo "Running tests"

set +e

RETRIES=1
MAX_RETRIES=5
while ! kubectl cluster-info &> /dev/null; do
    echo "API server not accessible; retrying..."
    if [ ${RETRIES} -eq ${MAX_RETRIES} ]; then
        exit 1
    fi
    ((RETRIES++))
    sleep 1
done

set -ex

find "/tmp/oneinfra-hypervisor-sets/${INFRA_TEST_CLUSTER_NAME}/" -type s -name "*.sock" | xargs -I{} -- bash -c 'echo {}; crictl --runtime-endpoint unix://{} pods'

find "/tmp/oneinfra-hypervisor-sets/${INFRA_TEST_CLUSTER_NAME}/" -type s -name "*.sock" | xargs -I{} -- bash -c 'echo {}; crictl --runtime-endpoint unix://{} ps -a'

kubectl cluster-info

kubectl version
