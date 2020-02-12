#!/usr/bin/env bash

if [ -z "$SKIP_BIN_MOUNT" ]; then
    EXTRA_FLAGS="-v $PWD/bin:/go/bin"
fi

docker run --rm -v $PWD:/usr/src/oneinfra $EXTRA_FLAGS -w /usr/src/oneinfra -e GOFLAGS="-mod=vendor" oneinfra/builder:latest "$@"
