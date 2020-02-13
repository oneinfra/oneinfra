#!/usr/bin/env bash

export PATH=$GOPATH/bin:./bin:$PATH

if [ -z "$CLUSTER_CONF" ]; then
    echo 'Please, set $CLUSTER_CONF environment variable pointing to your cluster manifests'
    exit 1
fi

if [ -z "$CLUSTER_NAME" ]; then
    echo 'Please, set $CLUSTER_NAME environment variable setting the name of the cluster you want to join this fake worker to'
    exit 1
fi

kubeconfig=$(mktemp /tmp/kubeconfig-XXXXXXX)
cat "$CLUSTER_CONF" | oi cluster kubeconfig --cluster "$CLUSTER_NAME" > "$kubeconfig"

docker run --privileged -v /dev/null:/proc/swaps:ro -v "$kubeconfig":/kubeconfig -e CONTAINER_RUNTIME_ENDPOINT=unix:///containerd-socket/containerd.sock -e IMAGE_SERVICE_ENDPOINT=unix:///containerd-socket/containerd.sock -d -i oneinfra/kubelet:latest
