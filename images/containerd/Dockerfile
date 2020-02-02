FROM ubuntu:18.04

ARG ARCH="amd64"
ARG CNI_VERSION="v0.8.5"

RUN apt-get update -y \
  && apt-get install -y containerd ca-certificates curl iptables \
  && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /etc/cni/net.d

ADD containerd-config.toml /etc/containerd/config.toml
ADD 10-containerd-net.conflist /etc/cni/net.d/10-containerd-net.conflist

RUN export CNI_URL="https://github.com/containernetworking/plugins/releases/download/${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" \
  && curl -sSL --retry 5 --output /tmp/cni.tgz "${CNI_URL}" \
  && mkdir -p /opt/cni/bin \
  && tar -C /opt/cni/bin -xzf /tmp/cni.tgz \
  && rm -f /tmp/cni.tgz

CMD ["/usr/bin/containerd"]
