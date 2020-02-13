# Image URL to use all building/pushing image targets
IMG ?= controller:latest

# kubectl version
KUBECTL_VERSION ?= 1.17.0

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Project top level folders
PROJECT_GO_FOLDERS = apis cmd controllers internal

# Project top level packages
PROJECT_GO_PACKAGES = $(foreach folder,${PROJECT_GO_FOLDERS},${folder}/...)

all: manager oi oi-local-cluster

# Run tests
test: lint fmt vet
	./run.sh go test ./... -coverprofile cover.out

# Build and install manager binary
manager: fmt vet
	./run.sh go install ./cmd/oi-manager

# Build and install oi binary
oi:
	./run.sh go install ./cmd/oi

# Build and install oi-local-cluster
oi-local-cluster:
	./run.sh go install ./cmd/oi-local-cluster

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
	./run.sh golint -set_exit_status=1 ${PROJECT_GO_PACKAGES}

# Run gofmt against code
fmt:
	@test -z "$(shell ./run.sh gofmt -l ${PROJECT_GO_FOLDERS})" || (./run.sh gofmt -d -l ${PROJECT_GO_FOLDERS} && exit 1)

# Run go vet against code
vet:
	./run.sh go vet ./...

# Generate code
generate:
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

pull: pull-builder
	@docker pull oneinfra/containerd:latest

pull-builder:
	@docker pull oneinfra/builder:latest

# Install kubectl
kubectl:
	sudo wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl
	sudo chmod +x /usr/local/bin/kubectl

# Run e2e (to be moved to a proper e2e framework)
e2e: oi oi-local-cluster
	mkdir -p ~/.kube
	bin/oi-local-cluster cluster create | bin/oi cluster inject --name test | bin/oi node inject --name test --cluster test --role controlplane | bin/oi node inject --name loadbalancer --cluster test --role controlplane-ingress | tee cluster.txt | bin/oi reconcile
	cat cluster.txt | bin/oi cluster kubeconfig --cluster test > ~/.kube/config
	docker ps -a
	kubectl cluster-info
	kubectl version
