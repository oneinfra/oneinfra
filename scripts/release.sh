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

HUB_VERSION=2.14.2
CURRENT_TAG=$(git describe --tags)

if ! git describe --exact-match HEAD &> /dev/null; then
    echo "HEAD is currently at ${CURRENT_TAG}. Skipping since ${CURRENT_TAG} is not a tag object"
    exit 0
fi

if [ -z "${GITHUB_TOKEN}" ]; then
    echo "Please, set GITHUB_TOKEN envvar"
    exit 1
fi

if [ -z "${DOCKER_HUB_TOKEN}" ]; then
    echo "Please, set DOCKER_HUB_TOKEN envvar"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

HUB_PATH=/tmp/hub-${HUB_VERSION}
export PATH=${HUB_PATH}:${PATH}
if ! which hub &> /dev/null; then
    wget -O hub.tgz https://github.com/github/hub/releases/download/v${HUB_VERSION}/hub-linux-amd64-${HUB_VERSION}.tgz
    tar --strip-components 2 -xf hub.tgz hub-linux-amd64-${HUB_VERSION}/bin/hub
    rm hub.tgz
    mkdir -p ${HUB_PATH}
    mv hub ${HUB_PATH}
fi

TARGET_COMMITISH=$(git show-ref -d ${CURRENT_TAG} | tail -n1 | awk '{print $1}')
CHANGELOG_FILE=/tmp/oneinfra-${CURRENT_TAG}-changelog.txt

echo ${CURRENT_TAG} > ${CHANGELOG_FILE}
echo >> ${CHANGELOG_FILE}

echo "## :bulb: Release notes" >> ${CHANGELOG_FILE}
RELEASE_NOTES=$(git log $(git tag --sort=-version:refname | head -n2 | tail -n1)..${CURRENT_TAG} | oi-releaser git release-notes)
if [ -n "${RELEASE_NOTES}" ]; then
    echo "${RELEASE_NOTES}" >> ${CHANGELOG_FILE}
else
    echo "There are no release notes for this release. Enjoy!" >> ${CHANGELOG_FILE}
fi
echo >> ${CHANGELOG_FILE}

echo "## :pencil: Changelog" >> ${CHANGELOG_FILE}
git log $(git tag --sort=-version:refname | head -n2 | tail -n1)..${CURRENT_TAG} --format="- %h: %s" >> ${CHANGELOG_FILE}

echo "Creating release ${CURRENT_TAG}"

hub release create -p -t "${TARGET_COMMITISH}" -F "${CHANGELOG_FILE}" "${CURRENT_TAG}"

echo "Publishing container images"

oi-releaser container-images publish

echo "Publishing release assets"

RUN_EXTRA_OPTS="-e GITHUB_TOKEN" ${SCRIPT_DIR}/run.sh oi-releaser binaries publish --binary oi --binary oi-local-hypervisor-set
