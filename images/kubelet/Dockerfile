FROM oneinfra/containerd:latest

ARG KUBERNETES_VERSION=1.17.0

COPY /fs /

RUN echo "Installing the kubelet" \
  && wget -O /usr/bin/kubelet https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubelet \
  && chmod +x /usr/bin/kubelet \
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