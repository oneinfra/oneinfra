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

package text

import (
	"fmt"
	"strings"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

// ReplacePlaceholders replaces common placeholders on the inputText
// and returns the replaced content
func ReplacePlaceholders(inputText string) string {
	replacements := map[string]string{
		"ONEINFRA_VERSION": constants.BuildVersion,
		"CONSOLE_VERSION":  constants.ReleaseData.ConsoleVersion,
	}
	for placeHolder, value := range replacements {
		inputText = strings.ReplaceAll(
			inputText,
			fmt.Sprintf("{%s}", placeHolder),
			value,
		)
	}
	return inputText
}
