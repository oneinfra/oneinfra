image:
	docker build --build-arg KUBERNETES_VERSION=$(KUBERNETES_VERSION) -t oneinfra/kubelet-installer:$(KUBERNETES_VERSION) .

publish:
	docker push oneinfra/kubelet-installer:$(KUBERNETES_VERSION)
