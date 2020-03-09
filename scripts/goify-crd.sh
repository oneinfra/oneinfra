#!/usr/bin/env bash

# Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
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

BOILERPLATE="$(cat hack/boilerplate.go.txt)"

for crdPath in $(find "${1}" -maxdepth 1 -mindepth 1 -type d); do
    rawCRD="$(controller-gen ${CRD_OPTIONS} paths="./${crdPath}/..." output:crd:stdout | yq read - "spec.validation.openAPIV3Schema")"
    crdVersion="$(basename "${crdPath}")"
    cat <<EOF > ${crdPath}/crd.go
$BOILERPLATE

package $crdVersion

const (
	// OpenAPISchema represents this CRD OpenAPI schema
	OpenAPISchema = \`$rawCRD
\`
)
EOF
done
