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

// FIXME: remove this script

BOILERPLATE="$(cat hack/boilerplate.go.txt)"

for crdPath in $(find "${1}" -maxdepth 1 -mindepth 1 -type d); do
    rawCRDs="$(controller-gen ${CRD_OPTIONS} paths="./${crdPath}/..." output:crd:stdout)"
    crdVersion="$(basename "${crdPath}")"
    allKinds=$(echo -e "${rawCRDs}" | yq eval '.spec.names.kind' -)
    cat <<EOF > ${crdPath}/zz_generated.openapi.go
$BOILERPLATE

package $crdVersion

// Code generated by openapi-gen script. DO NOT EDIT.

const (
EOF

    i=0
    for kind in ${allKinds}; do
        openAPISchema=$(echo -e "${rawCRDs}" | yq eval "select(documentIndex == ${i}) | .spec.versions[].schema.openAPIV3Schema" - | sed 's/`/"/g')
cat <<EOF >> ${crdPath}/zz_generated.openapi.go

	// ${kind}OpenAPISchema represents the OpenAPI schema for kind ${kind}
	${kind}OpenAPISchema = \`$openAPISchema\`
EOF
        i=$((i+1))
    done

    echo ")" >> ${crdPath}/zz_generated.openapi.go
done
