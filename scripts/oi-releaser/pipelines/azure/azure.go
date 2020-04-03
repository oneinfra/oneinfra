/*
Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

type Pipeline struct {
	Variables map[string]string `json:"variables,omitempty"`
	Jobs      []Job             `json:"jobs,omitempty"`
}

type Job struct {
	Job         string `json:"job,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Pool        Pool   `json:"pool,omitempty"`
	Steps       []Step `json:"steps,omitempty"`
}

type Pool struct {
	VMImage string `json:"vmImage,omitempty"`
}

type Step struct {
	Bash        string            `json:"bash,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}
