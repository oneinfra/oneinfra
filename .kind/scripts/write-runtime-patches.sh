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

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
TEST_WEBHOOK_CERTS_DIR=${TEST_WEBHOOK_CERTS_DIR:-/tmp/k8s-webhook-server/serving-certs}
DOCKER_GATEWAY=$(${SCRIPT_DIR}/docker-gateway.sh)
CA_CERT_CONTENTS=$(base64 -w0 ${TEST_WEBHOOK_CERTS_DIR}/ca.crt)

cat <<EOF > ${SCRIPT_DIR}/../kustomize/runtime-patches/mutating_webhook_configuration.yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: ${CA_CERT_CONTENTS}
    service: null
    url: https://${DOCKER_GATEWAY}:9443/mutate-cluster-oneinfra-ereslibre-es-v1alpha1-cluster
  name: mcluster.kb.io
- clientConfig:
    caBundle: ${CA_CERT_CONTENTS}
    service: null
    url: https://${DOCKER_GATEWAY}:9443/mutate-cluster-oneinfra-ereslibre-es-v1alpha1-component
  name: mcomponent.kb.io
EOF

cat <<EOF > ${SCRIPT_DIR}/../kustomize/runtime-patches/validating_webhook_configuration.yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  name: validating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: ${CA_CERT_CONTENTS}
    service: null
    url: https://${DOCKER_GATEWAY}:9443/validate-cluster-oneinfra-ereslibre-es-v1alpha1-cluster
  name: vcluster.kb.io
- clientConfig:
    caBundle: ${CA_CERT_CONTENTS}
    service: null
    url: https://${DOCKER_GATEWAY}:9443/validate-cluster-oneinfra-ereslibre-es-v1alpha1-component
  name: vcomponent.kb.io
EOF
