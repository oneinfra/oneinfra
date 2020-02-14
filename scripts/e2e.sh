#!/usr/bin/env bash

export PATH=${GOPATH}/bin:./bin:${PATH}

CLUSTER_CONF="${CLUSTER_CONF:-cluster.conf}"
CLUSTER_NAME="${CLUSTER_NAME:-cluster}"

mkdir -p ~/.kube
echo "Creating infrastructure"
oi-local-cluster cluster create > "${CLUSTER_CONF}"

# Get all IP addresses from docker containers, we don't care being
# picky here. This is required because of how fake workers will
# connect to the infrastructure, read more on the
# `create-fake-worker.sh` script
APISERVER_EXTRA_SANS="$(docker ps -aq | xargs docker inspect -f '{{ .NetworkSettings.IPAddress }}' | xargs -I{} echo "--apiserver-extra-sans {}" | paste -sd " " -)"

cat "${CLUSTER_CONF}" | \
    oi cluster inject --name "${CLUSTER_NAME}" ${APISERVER_EXTRA_SANS} | \
    oi node inject --name controlplane --cluster "${CLUSTER_NAME}" --role controlplane | \
    oi node inject --name loadbalancer --cluster "${CLUSTER_NAME}" --role controlplane-ingress | \
    oi reconcile | \
    tee "${CLUSTER_CONF}" | \
    oi cluster kubeconfig --cluster "${CLUSTER_NAME}" > ~/.kube/config

# Tests

docker ps -a

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

kubectl cluster-info
kubectl version
