#!/usr/bin/env bash

docker run --rm -v $PWD:/app $EXTRA_FLAGS -v $PWD/bin:/go/bin -w /app -e GOFLAGS="-mod=vendor" oneinfra/builder:latest "$@"
