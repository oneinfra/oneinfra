FROM ubuntu:18.04

COPY /fs /

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       software-properties-common \
       systemd \
  && add-apt-repository -y ppa:projectatomic/ppa \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
       conmon \
       containerd \
       ca-certificates \
       curl \
       iptables \
       containernetworking-plugins \
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
