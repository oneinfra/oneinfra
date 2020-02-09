# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# kubebuilder tools version
KUBEBUILDER_TOOLS_VERSION ?= 1.16.4

# Project top level folders
PROJECT_GO_FOLDERS = apis cmd controllers internal

# Project top level packages
PROJECT_GO_PACKAGES = $(foreach folder,${PROJECT_GO_FOLDERS},${folder}/...)

all: manager oi oi-local-cluster

# Install test deps
test-deps:
	wget https://go.kubebuilder.io/test-tools/${KUBEBUILDER_TOOLS_VERSION}/linux/amd64 -O kubebuilder-tools.tar.gz
	tar -C /usr/local -xf kubebuilder-tools.tar.gz
	rm kubebuilder-tools.tar.gz

# Run tests
test: fmt vet
	go test ./... -coverprofile cover.out

# Build and install manager binary
manager: generate fmt vet
	go install ./cmd/oi-manager

# Build and install oi binary
oi:
	go install ./cmd/oi

# Build and install oi-local-cluster
oi-local-cluster:
	go install ./cmd/oi-local-cluster

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
	golint -set_exit_status=1 ${PROJECT_GO_PACKAGES}

# Run gofmt against code
fmt:
	@test -z "$(shell gofmt -l ${PROJECT_GO_FOLDERS})" || (gofmt -d -l ${PROJECT_GO_FOLDERS} && exit 1)

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
