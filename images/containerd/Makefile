image:
	docker build --build-arg CRI_TOOLS_VERSION=$(CRI_TOOLS_VERSION) --build-arg CNI_PLUGINS_VERSION=$(CNI_PLUGINS_VERSION) --build-arg CONTAINERD_VERSION=$(CONTAINERD_VERSION) -t oneinfra/containerd:$(CONTAINERD_VERSION) .

publish:
	docker push oneinfra/containerd:$(CONTAINERD_VERSION)
