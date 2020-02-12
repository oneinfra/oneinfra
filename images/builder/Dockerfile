ARG GOLANG_VERSION=1.13

FROM golang:$GOLANG_VERSION

ARG CONTROLLER_GEN_VERSION=0.2.4
ARG KUBEBUILDER_TOOLS_VERSION=1.16.4

ENV GO111MODULE=on

RUN echo "Installing dependencies" \
  && go get -u golang.org/x/lint/golint \
  && go get sigs.k8s.io/controller-tools/cmd/controller-gen@v$CONTROLLER_GEN_VERSION \
  && wget https://go.kubebuilder.io/test-tools/$KUBEBUILDER_TOOLS_VERSION/linux/amd64 -O kubebuilder-tools.tar.gz \
  && tar -C /usr/local -xf kubebuilder-tools.tar.gz \
  && rm kubebuilder-tools.tar.gz