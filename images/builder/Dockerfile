ARG GOLANG_VERSION=1.14

FROM golang:${GOLANG_VERSION} AS builder
ARG CONTROLLER_GEN_VERSION=0.2.5
ENV GO111MODULE=on
RUN echo "Installing dependencies" \
  && GOBIN=/usr/local/bin go get -u github.com/mikefarah/yq/v3 \
  && GOBIN=/usr/local/bin go get -u golang.org/x/lint/golint \
  && GOBIN=/usr/local/bin go get -u k8s.io/code-generator/cmd/client-gen \
  && GOBIN=/usr/local/bin go get -u k8s.io/code-generator/cmd/openapi-gen \
  && GOBIN=/usr/local/bin go get -u sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION}


FROM golang:${GOLANG_VERSION}
COPY --from=builder /usr/local/bin /usr/local/bin