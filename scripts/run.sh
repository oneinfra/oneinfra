#!/usr/bin/env bash

# Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ -z "$CI" ]; then
    GOFLAGS="-mod=vendor" "$@"
else
    if [ ! -z "$SKIP_CI" ]; then
        exit 0
    fi
    set -x
    docker run --rm -v ${PWD}:/app -v ${PWD}/bin:/go/bin -v /tmp/go-build-cache:/root/.cache/go-build -w /app -e GOFLAGS="-mod=vendor" ${RUN_EXTRA_OPTS} oneinfra/builder:latest "$@"
fi
