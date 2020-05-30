BUILD_VERSION ?= $(shell git describe --tags --dirty)

image:
	docker build --build-arg ONEINFRA_VERSION=${BUILD_VERSION} -t oneinfra/oi:${ONEINFRA_VERSION} -f Dockerfile ../..

publish:
	docker push oneinfra/oi:${ONEINFRA_VERSION}
