FROM ubuntu:18.04

ARG CRI_TOOLS_VERSION
ARG CNI_PLUGINS_VERSION
ARG CONTAINERD_VERSION

COPY /fs /

ENV IMAGE_SERVICE_ENDPOINT unix:///containerd-socket/containerd.sock
ENV CONTAINER_RUNTIME_ENDPOINT unix:///containerd-socket/containerd.sock

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       systemd \
       runc \
       ca-certificates \
       curl \
       iptables \
       wget \
&& echo "Installing CRI tools" \
  && wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v${CRI_TOOLS_VERSION}/crictl-v${CRI_TOOLS_VERSION}-linux-amd64.tar.gz -O cri-tools.tar.gz \
  && tar -C /usr/local/bin -xf cri-tools.tar.gz \
  && rm cri-tools.tar.gz \
&& echo "Installing CNI plugins" \
  && mkdir -p /opt/cni/bin \
  && wget https://github.com/containernetworking/plugins/releases/download/v${CNI_PLUGINS_VERSION}/cni-plugins-linux-amd64-v${CNI_PLUGINS_VERSION}.tgz -O cni.tgz \
  && tar -C /opt/cni/bin -xf cni.tgz \
  && rm cni.tgz \
&& echo "Installing containerd" \
  && wget https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz -O containerd.tar.gz \
  && tar -C /usr/local -xf containerd.tar.gz \
  && rm containerd.tar.gz \
  && rm /usr/local/bin/containerd-stress /usr/local/bin/ctr \
  && systemctl enable containerd \
&& echo "Ensuring /etc/cni/net.d" \
  && mkdir -p /etc/cni/net.d \
&& echo "Cleaning up image" \
  && apt-get clean -y \
  && rm -rf /var/cache/debconf/* \
            /var/lib/apt/lists/* \
            /var/log/* \
            /tmp/* \
            /var/tmp/* \
            /usr/share/doc/* \
            /usr/share/man/* \
            /usr/share/local/*

# tell systemd that it is in docker (it will check for the container env)
# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
ENV container docker

# systemd exits on SIGRTMIN+3, not SIGTERM (which re-executes it)
# https://bugzilla.redhat.com/show_bug.cgi?id=1201657
STOPSIGNAL SIGRTMIN+3

ENTRYPOINT ["/usr/local/bin/entrypoint", "/bin/systemd"]
