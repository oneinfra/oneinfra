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

package utils

// AddElementsToListIfNotExists adds the provided elements to the list
// only if they are not present
func AddElementsToListIfNotExists(list []string, elements ...string) []string {
	res := []string{}
	elementMap := map[string]struct{}{}
	for _, element := range list {
		elementMap[element] = struct{}{}
		res = append(res, element)
	}
	for _, element := range elements {
		if _, found := elementMap[element]; !found {
			res = append(res, element)
		}
	}
	return res
}

// HasListAnyElement returns whether the list has any of the provided
// elements
func HasListAnyElement(list []string, elements ...string) bool {
	elementMap := map[string]struct{}{}
	for _, element := range elements {
		elementMap[element] = struct{}{}
	}
	for _, element := range list {
		if _, found := elementMap[element]; found {
			return true
		}
	}
	return false
}

// RemoveElementsFromList removes elements from a given list. The
// resulting list preserves the original order
func RemoveElementsFromList(list []string, elements ...string) []string {
	elementMap := map[string]struct{}{}
	for _, element := range elements {
		elementMap[element] = struct{}{}
	}
	res := []string{}
	for _, element := range list {
		if _, found := elementMap[element]; !found {
			res = append(res, element)
		}
	}
	return res
}
