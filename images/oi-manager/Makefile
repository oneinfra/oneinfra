ONEINFRA_VERSION ?= $(shell git describe --tags)
ONEINFRA_VERSION_TAG ?= $(shell git describe --tags)

image:
	docker build --build-arg ONEINFRA_VERSION=${ONEINFRA_VERSION} -t oneinfra/oi-manager:${ONEINFRA_VERSION_TAG} -f Dockerfile ../..

publish:
	docker push oneinfra/oi-manager:${ONEINFRA_VERSION_TAG}
