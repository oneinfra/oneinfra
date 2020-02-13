FROM oneinfra/containerd:latest

COPY /fs /

RUN echo "Installing dependencies" \
  && apt-get update -qq \
  && apt-get install --no-install-recommends -qq -y \
      apt-transport-https gpg gpg-agent \
  && echo 'deb https://apt.kubernetes.io/ kubernetes-xenial main' > /etc/apt/sources.list.d/kubernetes.list \
  && curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - \
  && apt-get update -qq \
&& echo "Installing the kubelet" \
  && apt-get install -y kubelet \
  && systemctl enable kubelet \
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