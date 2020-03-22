# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Project top level folders
PROJECT_GO_FOLDERS = apis cmd controllers internal

# Project top level packages
PROJECT_GO_PACKAGES = $(foreach folder,${PROJECT_GO_FOLDERS},${folder}/...)

all: manager oi oi-local-cluster

# Run tests
test: lint fmt vet
	./scripts/run.sh go test ./... -coverprofile cover.out

# Build and install manager binary
manager: go-generate
	./scripts/run.sh go install ./cmd/oi-manager

# Build and install oi binary
oi: go-generate
	./scripts/run.sh go install ./cmd/oi

# Build and install oi-local-cluster
oi-local-cluster: go-generate
	./scripts/run.sh go install ./cmd/oi-local-cluster

go-generate: RELEASE
	sh -c "SKIP_CI=1 ./scripts/run.sh go generate ./..."
	sh -c "SKIP_CI=1 ./scripts/run.sh sh -c 'cd scripts/oi-releaser && go mod vendor'"

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run cmd/oi-manager/main.go

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
generate: manifests
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

deps: pull kubectl crictl wg

pull: check-kubernetes-version-provided pull-builder
	@docker pull oneinfra/hypervisor:$(KUBERNETES_VERSION)

pull-builder:
	@docker pull oneinfra/builder:latest

builder-shell:
	sh -c 'CI="1" RUN_EXTRA_OPTS="-it" ./scripts/run.sh bash'

# Install kubectl
kubectl: check-kubernetes-version-provided
	sudo wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl
	sudo chmod +x /usr/local/bin/kubectl

# Install crictl
crictl: check-cri-tools-version-provided
	wget -O cri-tools.tar.gz https://github.com/kubernetes-sigs/cri-tools/releases/download/v${CRI_TOOLS_VERSION}/crictl-v${CRI_TOOLS_VERSION}-linux-amd64.tar.gz
	sudo tar -C /usr/local/bin -xf cri-tools.tar.gz
	rm cri-tools.tar.gz

# Installs wireguard
wg:
	./scripts/install-wireguard.sh

# Run e2e with local CRI endpoints (to be moved to a proper e2e framework)
e2e: oi oi-local-cluster
	./scripts/e2e.sh

# Run e2e with remote CRI endpoints (to be moved to a proper e2e framework)
e2e-remote: oi oi-local-cluster
	./scripts/e2e.sh --remote

create-fake-worker:
	./scripts/create-fake-worker.sh

oi-releaser:
	./scripts/run.sh sh -c "cd scripts/oi-releaser && go install -mod=vendor ."

build-container-images: oi-releaser
	./scripts/run-local.sh oi-releaser container-images build

publish-container-images: oi-releaser build-container-images
	docker login -u oneinfrapublisher -p $(DOCKER_HUB_TOKEN)
	./scripts/run-local.sh oi-releaser container-images publish

check-kubernetes-version-provided:
ifndef KUBERNETES_VERSION
	$(error KUBERNETES_VERSION envvar is undefined)
endif

check-cri-tools-version-provided:
ifndef CRI_TOOLS_VERSION
	$(error CRI_TOOLS_VERSION envvar is undefined)
endif
