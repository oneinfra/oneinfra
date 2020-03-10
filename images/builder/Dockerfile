ARG GOLANG_VERSION=1.13

FROM golang:${GOLANG_VERSION}

ARG CONTROLLER_GEN_VERSION=0.2.5

ENV GO111MODULE=on

RUN echo "Installing dependencies" \
  && GOBIN=/usr/local/bin go get -u github.com/mikefarah/yq/v3 \
  && GOBIN=/usr/local/bin go get -u golang.org/x/lint/golint \
  && GOBIN=/usr/local/bin go get -u sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION} \
&& echo "Cleaning up image" \
  && go clean -cache -testcache -modcache
