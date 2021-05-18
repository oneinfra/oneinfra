# Kubernetes version to use
KUBERNETES_VERSION ?= default

# The build version of oneinfra
BUILD_VERSION ?= $(shell git describe --tags --dirty)

GO_INSTALL_FLAGS ?= -ldflags='-X github.com/oneinfra/oneinfra/internal/pkg/constants.BuildVersion=${BUILD_VERSION}'

# Image URL to use all building/pushing image targets
IMG ?= controller:latest

TEST_WEBHOOK_CERTS_DIR ?= /tmp/k8s-webhook-server/serving-certs

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Project top level folders
PROJECT_GO_FOLDERS = apis cmd controllers internal pkg

# Project top level packages
PROJECT_GO_PACKAGES = $(foreach folder,${PROJECT_GO_FOLDERS},${folder}/...)

oi-binaries: manager oi oi-local-hypervisor-set

all: oi-binaries generate pipelines

# Run tests
test: lint fmt vet
	go test ./... -coverprofile cover.out

test-coverage: test
	go tool cover -html=cover.out

# Build and install manager binary
manager: go-generate
	go build -o bin/manager ${GO_INSTALL_FLAGS} ./cmd/oi-manager

# Build and install oi binary
oi: go-generate
	go build -o bin/oi ${GO_INSTALL_FLAGS} ./cmd/oi

# Build and install oi-local-hypervisor-set
oi-local-hypervisor-set: go-generate
	go build -o bin/oi-local-hypervisor-set ${GO_INSTALL_FLAGS} ./cmd/oi-local-hypervisor-set

clientsets-generate:
	rm -rf pkg/clientsets/*
	client-gen --input-base github.com/oneinfra/oneinfra/apis --input cluster/v1alpha1 --output-base pkg/clientsets -h hack/boilerplate.go.txt -p github.com/oneinfra/oneinfra/pkg/clientsets -n manager
	client-gen --input-base github.com/oneinfra/oneinfra/apis --input node/v1alpha1 --output-base pkg/clientsets -h hack/boilerplate.go.txt -p github.com/oneinfra/oneinfra/pkg/clientsets -n managed
	mv pkg/clientsets/github.com/oneinfra/oneinfra/pkg/clientsets/* pkg/clientsets
	rm -rf pkg/clientsets/github.com

go-generate: RELEASE
	go generate ./...

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run cmd/oi-manager/main.go -verbosity 10

# Run against a kind cluster with webhooks set up with generated certificates
run-kind: kind-webhook-certs kind kind-deploy run

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: platform-manifests guest-manifests
	kustomize build config/default > config/generated/all.yaml
	kustomize build config/nightly > config/generated/nightly.yaml

platform-manifests:
	controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths=./apis/cluster/... paths=./apis/infra/... output:crd:artifacts:config=config/crd/bases

guest-manifests:
	scripts/openapi-gen.sh apis/node

# Run golint against code
lint:
	golint -set_exit_status=1 ${PROJECT_GO_PACKAGES}

# Run gofmt against code
fmt:
	@test -z "$(shell gofmt -l ${PROJECT_GO_FOLDERS})" || (gofmt -d -l ${PROJECT_GO_FOLDERS} && exit 1)

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: replace-text-placeholders manifests
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

generate-all: generate clientsets-generate pipelines

# Run e2e with local CRI endpoints (to be moved to a proper e2e framework)
e2e: oi oi-local-hypervisor-set
	./scripts/e2e.sh

# Run e2e with remote CRI endpoints (to be moved to a proper e2e framework)
e2e-remote: oi oi-local-hypervisor-set
	./scripts/e2e.sh --tcp

create-fake-worker:
	./scripts/create-fake-worker.sh
