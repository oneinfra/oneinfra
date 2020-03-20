FROM alpine:latest

ARG KUBERNETES_VERSION

COPY /fs /

RUN echo "Downloading the kubelet" \
  && wget -O /usr/bin/kubelet https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubelet \
  && chmod +x /usr/bin/kubelet

ENTRYPOINT ["/usr/bin/entrypoint"]