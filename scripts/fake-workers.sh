#!/usr/bin/env bash

docker run --net=host --privileged -v /dev/null:/proc/swaps:ro -e CONTAINER_RUNTIME_ENDPOINT=unix:///containerd-socket/containerd.sock -e IMAGE_SERVICE_ENDPOINT=unix:///containerd-socket/containerd.sock -d -i oneinfra/kubelet:latest
