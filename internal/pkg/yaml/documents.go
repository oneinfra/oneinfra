/**
 * Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package yaml

import (
	"regexp"
	"strings"
)

const (
	trimmer = "\n\t "
)

var (
	splitter = regexp.MustCompile("(?m)^---$")
)

// SplitDocuments splits a multi-document YAML into a list of single
// YAML documents, ignoring empty documents
func SplitDocuments(manifests string) []string {
	documents := []string{}
	for _, document := range splitter.Split(manifests, -1) {
		if len(strings.Trim(document, trimmer)) > 0 {
			documents = append(documents, document)
		}
	}
	return documents
}
