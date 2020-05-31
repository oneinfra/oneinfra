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

package azure

// If you happen to wonder about underscores, have a look at:
//
// https://developercommunity.visualstudio.com/content/problem/346122/job-parser-erroneously-treats-yaml-mapping-keys-or.html
//
// Azure Pipelines created their own YAML parser that doesn't strictly
// follow the YAML spec, and thus, requires some fields to be
// "first". These are the fields you see with underscore in the
// structs. The final textual YAML file will be post processed to
// remove the underscores.

// Pipeline represents an Azure pipeline
type Pipeline struct {
	Variables map[string]string `json:"variables,omitempty"`
	Trigger   *Trigger          `json:"trigger,omitempty"`
	PR        *PRTrigger        `json:"pr,omitempty"`
	Jobs      []Job             `json:"jobs,omitempty"`
}

// Trigger represents a pipeline trigger
type Trigger struct {
	Branches *BranchesTrigger `json:"branches,omitempty"`
	Tags     *TagsTrigger     `json:"tags,omitempty"`
	Paths    *PathsTrigger    `json:"paths,omitempty"`
}

// BranchesTrigger represents a branch trigger
type BranchesTrigger struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// TagsTrigger represents a tag trigger
type TagsTrigger struct {
	Include []string `json:"include,omitempty"`
}

// PathsTrigger represents a path trigger
type PathsTrigger struct {
	Include []string `json:"include,omitempty"`
}

// PRTrigger represents a PR trigger
type PRTrigger struct {
	Branches *BranchesTrigger `json:"branches,omitempty"`
}

// Job represents an Azure job
type Job struct {
	Job         string   `json:"_job,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Pool        Pool     `json:"pool,omitempty"`
	Steps       []Step   `json:"steps,omitempty"`
	DependsOn   []string `json:"dependsOn,omitempty"`
}

// Pool represents an Azure pool
type Pool struct {
	VMImage string `json:"vmImage,omitempty"`
}

// Step represents an Azure step
type Step struct {
	Bash        string            `json:"_bash,omitempty"`
	DisplayName string            `json:"displayName,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}
