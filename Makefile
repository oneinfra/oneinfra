# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Kubernetes version
KUBERNETES_VERSION ?= 1.17.0

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
manager:
	./scripts/run.sh go install ./cmd/oi-manager

# Build and install oi binary
oi:
	./scripts/run.sh go install ./cmd/oi

# Build and install oi-local-cluster
oi-local-cluster:
	./scripts/run.sh go install ./cmd/oi-local-cluster

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
manifests:
	controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

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
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

deps: pull kubectl crictl wg

pull: pull-builder
	@docker pull oneinfra/hypervisor:latest

pull-builder:
	@docker pull oneinfra/builder:latest

# Install kubectl
kubectl:
	sudo wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl
	sudo chmod +x /usr/local/bin/kubectl

# Install crictl
crictl:
	wget -O cri-tools.tar.gz https://github.com/kubernetes-sigs/cri-tools/releases/download/v${KUBERNETES_VERSION}/crictl-v${KUBERNETES_VERSION}-linux-amd64.tar.gz
	sudo tar -C /usr/local/bin -xf cri-tools.tar.gz
	rm cri-tools.tar.gz

# Installs wireguard
wg:
	./scripts/install-wireguard.sh

# Build an hypervisor image with many images already present (for faster local testing cycles)
e2e-build-hypervisor-image:
	./scripts/build-hypervisor-image.sh

# Run e2e (to be moved to a proper e2e framework)
e2e: oi oi-local-cluster
	./scripts/e2e.sh

e2e-remote: oi oi-local-cluster
	./scripts/e2e.sh --remote

# Creates a fake worker
create-fake-worker:
	./scripts/create-fake-worker.sh
