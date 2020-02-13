#!/usr/bin/env bash

if [ -z "$CI" ]; then
    GOFLAGS="-mod=vendor" "$@"
else
    set -x
    docker run --rm -v $PWD:/app -v $PWD/bin:/go/bin -v /tmp/go-build-cache:/root/.cache/go-build $EXTRA_FLAGS -w /app -e GOFLAGS="-mod=vendor" oneinfra/builder:latest "$@"
fi
