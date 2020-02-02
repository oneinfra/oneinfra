FROM ubuntu:18.04

RUN apt-get update -y \
  && apt-get install -y containerd ca-certificates \
  && rm -rf /var/lib/apt/lists/*

ADD containerd-config.toml /etc/containerd/config.toml

CMD ["/usr/bin/containerd"]
