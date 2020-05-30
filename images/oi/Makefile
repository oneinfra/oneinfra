ONEINFRA_VERSION ?= $(shell git describe --tags --dirty)
ONEINFRA_VERSION_TAG ?= $(shell git describe --tags)

image:
	docker build --build-arg ONEINFRA_VERSION=${ONEINFRA_VERSION} -t oneinfra/oi:${ONEINFRA_VERSION_TAG} -f Dockerfile ../..

publish:
	docker push oneinfra/oi:${ONEINFRA_VERSION_TAG}
