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
	./scripts/run.sh go test ./... -coverprofile cover.out

test-coverage: test
	go tool cover -html=cover.out

# Build and install manager binary
manager: go-generate
	./scripts/run.sh go install ${GO_INSTALL_FLAGS} ./cmd/oi-manager

# Build and install oi binary
oi: go-generate
	./scripts/run.sh go install ${GO_INSTALL_FLAGS} ./cmd/oi

# Build and install oi-local-hypervisor-set
oi-local-hypervisor-set: go-generate
	./scripts/run.sh go install ${GO_INSTALL_FLAGS} ./cmd/oi-local-hypervisor-set

# Build and install oi-releaser
oi-releaser: oi
	./scripts/run.sh go install ${GO_INSTALL_FLAGS} ./cmd/oi-releaser

clientsets-generate:
	rm -rf pkg/clientsets/*
	client-gen --input-base github.com/oneinfra/oneinfra/apis --input cluster/v1alpha1 --output-base pkg/clientsets -h hack/boilerplate.go.txt -p github.com/oneinfra/oneinfra/pkg/clientsets -n manager
	client-gen --input-base github.com/oneinfra/oneinfra/apis --input node/v1alpha1 --output-base pkg/clientsets -h hack/boilerplate.go.txt -p github.com/oneinfra/oneinfra/pkg/clientsets -n managed
	mv pkg/clientsets/github.com/oneinfra/oneinfra/pkg/clientsets/* pkg/clientsets
	rm -rf pkg/clientsets/github.com

pipelines: oi-releaser
	oi-releaser pipelines test dump > .azure-pipelines/test.yml
	oi-releaser pipelines release dump > .azure-pipelines/release.yml
	oi-releaser pipelines publish-tooling-images dump > .azure-pipelines/publish-tooling-images.yml
	oi-releaser pipelines publish-nightly-images dump > .azure-pipelines/publish-nightly-images.yml
	oi-releaser pipelines publish-testing-images dump > .azure-pipelines/publish-testing-images.yml

go-generate: RELEASE
	sh -c 'SKIP_CI=1 ./scripts/run.sh go generate ./...'

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
	./scripts/run.sh controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths=./apis/cluster/... paths=./apis/infra/... output:crd:artifacts:config=config/crd/bases

guest-manifests:
	sh -c 'CRD_OPTIONS=$(CRD_OPTIONS) RUN_EXTRA_OPTS="-e CRD_OPTIONS" ./scripts/run.sh ./scripts/openapi-gen.sh apis/node'

# Run golint against code
lint:
	./scripts/run.sh golint -set_exit_status=1 ${PROJECT_GO_PACKAGES}

# Run gofmt against code
fmt:
	@test -z "$(shell ./scripts/run.sh gofmt -l ${PROJECT_GO_FOLDERS})" || (./scripts/run.sh gofmt -d -l ${PROJECT_GO_FOLDERS} && exit 1)

# Run go vet against code
vet:
	./scripts/run.sh go vet ./...

# Generate code
generate: replace-text-placeholders manifests
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

generate-all: generate clientsets-generate pipelines

replace-text-placeholders: oi-releaser
	@./scripts/replace-text-placeholders.sh

os-information:
	cat /etc/os-release
	uname -a

deps: os-information pull-builder oi oi-releaser pull-or-build-hypervisor kubectl crictl

pull-or-build-hypervisor:
	KUBERNETES_VERSION=$(KUBERNETES_VERSION) ./scripts/install-requirements.sh pull-or-build-hypervisor

pull-builder:
	@docker pull oneinfra/builder:latest

builder-shell:
	sh -c 'CI="1" RUN_EXTRA_OPTS="-it" ./scripts/run.sh bash'

# Install kubectl
kubectl:
	KUBERNETES_VERSION=$(KUBERNETES_VERSION) ./scripts/install-requirements.sh kubectl

# Install crictl
crictl:
	KUBERNETES_VERSION=$(KUBERNETES_VERSION) ./scripts/install-requirements.sh crictl

# Run e2e with local CRI endpoints (to be moved to a proper e2e framework)
e2e: oi oi-local-hypervisor-set
	./scripts/e2e.sh

# Run e2e with remote CRI endpoints (to be moved to a proper e2e framework)
e2e-remote: oi oi-local-hypervisor-set
	./scripts/e2e.sh --tcp

create-fake-worker:
	./scripts/create-fake-worker.sh

kind-webhook-certs: $(TEST_WEBHOOK_CERTS_DIR) $(TEST_WEBHOOK_CERTS_DIR)/tls.crt

$(TEST_WEBHOOK_CERTS_DIR):
	mkdir -p $(TEST_WEBHOOK_CERTS_DIR)

$(TEST_WEBHOOK_CERTS_DIR)/ca.key:
	openssl genrsa -out $(TEST_WEBHOOK_CERTS_DIR)/ca.key 1024

$(TEST_WEBHOOK_CERTS_DIR)/ca.crt: $(TEST_WEBHOOK_CERTS_DIR)/ca.key
	openssl req -x509 -new -nodes -key $(TEST_WEBHOOK_CERTS_DIR)/ca.key -subj "/C=ES/ST=Madrid/O=oneinfra/CN=webhook" -sha256 -days 3650 -out $(TEST_WEBHOOK_CERTS_DIR)/ca.crt

$(TEST_WEBHOOK_CERTS_DIR)/tls.key:
	openssl genrsa -out $(TEST_WEBHOOK_CERTS_DIR)/tls.key 1024

$(TEST_WEBHOOK_CERTS_DIR)/tls.csr: $(TEST_WEBHOOK_CERTS_DIR)/tls.key
	openssl req -new -sha256 -key $(TEST_WEBHOOK_CERTS_DIR)/tls.key -subj "/C=ES/ST=Madrid/O=oneinfra/CN=$(shell .kind/scripts/docker-gateway.sh)" -out $(TEST_WEBHOOK_CERTS_DIR)/tls.csr

$(TEST_WEBHOOK_CERTS_DIR)/tls.crt: $(TEST_WEBHOOK_CERTS_DIR)/tls.csr $(TEST_WEBHOOK_CERTS_DIR)/ca.crt $(TEST_WEBHOOK_CERTS_DIR)/ca.key
	openssl x509 -req -in $(TEST_WEBHOOK_CERTS_DIR)/tls.csr -CA $(TEST_WEBHOOK_CERTS_DIR)/ca.crt -CAkey $(TEST_WEBHOOK_CERTS_DIR)/ca.key -CAcreateserial -out $(TEST_WEBHOOK_CERTS_DIR)/tls.crt -days 3650 -sha256

kind:
	kind create cluster --name oi-test-cluster --config .kind/config.yaml

kind-deploy:
	./.kind/scripts/write-runtime-patches.sh
	kubectl apply -k .kind/kustomize

kind-delete:
	kind delete cluster --name oi-test-cluster

build-container-image: oi-releaser
	./scripts/run-local.sh oi-releaser container-images build $(CONTAINER_BUILD_OPTIONS)

build-container-images: oi-releaser
	./scripts/run-local.sh oi-releaser container-images build

publish-container-image: oi-releaser
	./scripts/run-local.sh oi-releaser container-images publish $(CONTAINER_BUILD_OPTIONS)

publish-container-image-ci: docker-login publish-container-image

publish-container-images: oi-releaser
	./scripts/run-local.sh oi-releaser container-images publish

release: oi-releaser docker-login
	sh -c 'RUN_EXTRA_OPTS="-t" ./scripts/release.sh'

docker-login:
	echo $(DOCKER_HUB_TOKEN) | docker login -u oneinfrapublisher --password-stdin
