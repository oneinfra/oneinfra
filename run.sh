#!/usr/bin/env bash

docker run --rm -v $PWD:/app -v $PWD/bin:/go/bin -v /tmp/go-build-cache:/root/.cache/go-build $EXTRA_FLAGS -w /app -e GOFLAGS="-mod=vendor" oneinfra/builder:latest "$@"
