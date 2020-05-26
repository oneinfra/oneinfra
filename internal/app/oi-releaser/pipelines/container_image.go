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

package pipelines

import (
	"fmt"
	"strings"

	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/pipelines/azure"
)

func publishContainerJob(container string) azure.Job {
	return azure.Job{
		Job:         jobName(container),
		DisplayName: fmt.Sprintf("Publish %s container image", container),
		Pool:        azure.DefaultPool,
		Steps: []azure.Step{
			{
				Bash:        "make publish-container-image-ci",
				DisplayName: "Publish container image",
				Env: map[string]string{
					"CONTAINER_IMAGE":  container,
					"DOCKER_HUB_TOKEN": "$(DOCKER_HUB_TOKEN)",
				},
			},
		},
	}
}

func jobName(container string) string {
	underscoredVersion := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(container, ".", "_"),
			"-", "_",
		),
		":", "_",
	)
	return fmt.Sprintf("publish_%s_container_image", underscoredVersion)
}
