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

for file in $(find . -type f -not -path "./dhall/*" -not -path "./apis/*"  -name "*.dhall" -not -name "*.yaml.dhall"); do
    TARGET_FILE="$(echo $file | sed s/.dhall$//)"
    dhall text --file "$file" > ${TARGET_FILE}
done

for file in $(find . -type f -name "*.yaml.dhall"); do
    TARGET_FILE="$(echo $file | sed s/.dhall$//)"
    dhall-to-yaml --file "$file" > ${TARGET_FILE}
done
