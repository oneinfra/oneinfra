FROM oneinfra/builder:latest as builder
RUN mkdir -p /oneinfra
WORKDIR /oneinfra
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY apis/ apis/
COPY controllers/ controllers/
COPY internal/ internal/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go install -mod=vendor github.com/oneinfra/oneinfra/cmd/oi-manager


FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/bin/oi-manager .
USER nonroot:nonroot
ENTRYPOINT ["/oi-manager"]
