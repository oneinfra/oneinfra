FROM ubuntu:18.04

RUN apt-get update -qq \
  && apt-get install -qq -y software-properties-common \
  && add-apt-repository -y ppa:projectatomic/ppa \
  && apt-get update -qq \
  && apt-get install -qq -y containerd ca-certificates curl iptables containernetworking-plugins \
  && rm -rf /var/lib/apt/lists/*

ADD containerd-config.toml /etc/containerd/config.toml

RUN mkdir -p /etc/cni/net.d
ADD 10-containerd-net.conflist /etc/cni/net.d/10-containerd-net.conflist

CMD ["/usr/bin/containerd"]
