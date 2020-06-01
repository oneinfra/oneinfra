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

package git

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	commitIdentifier             = regexp.MustCompile("^commit ([a-f0-9]+)")
	releaseNoteIdentifier        = regexp.MustCompile("^(\\s*)```release-note")
	closingReleaseNoteIdentifier = regexp.MustCompile("^\\s*```$")
)

// ReleaseNote represents a release note tied to a commit
type ReleaseNote struct {
	Commit      string
	ReleaseNote string
}

// ReleaseNotes returns the release notes from the textual list of
// commits provided in inputText
func ReleaseNotes(inputText string) []ReleaseNote {
	releaseNotes := []ReleaseNote{}
	inputBuffer := bytes.NewBufferString(inputText)
	scanner := bufio.NewScanner(inputBuffer)
	scanner.Split(bufio.ScanLines)
	var currentCommit string
	var currentReleaseNote []string
	var identifyingReleaseNote bool
	var leftMargin int
	for scanner.Scan() {
		line := scanner.Text()
		commit := commitIdentifier.FindStringSubmatch(line)
		if len(commit) > 0 {
			currentCommit = commit[1]
			currentReleaseNote = []string{}
			identifyingReleaseNote = false
			continue
		}
		releaseNote := releaseNoteIdentifier.FindStringSubmatch(line)
		if len(releaseNote) > 0 {
			leftMargin = len(releaseNote[1])
			identifyingReleaseNote = true
			continue
		}
		if closingReleaseNoteIdentifier.MatchString(line) {
			identifyingReleaseNote = false
			if len(currentReleaseNote) == 0 {
				continue
			}
			releaseNotes = append(
				releaseNotes,
				ReleaseNote{
					Commit:      currentCommit,
					ReleaseNote: strings.Join(currentReleaseNote, "\n"),
				},
			)
			continue
		}
		if identifyingReleaseNote {
			releaseNoteContentFetcher := regexp.MustCompile(fmt.Sprintf("^\\s{%d}(.*)$", leftMargin))
			releaseNoteLine := releaseNoteContentFetcher.FindStringSubmatch(line)
			if len(releaseNoteLine) == 0 {
				continue
			}
			currentReleaseNote = append(
				currentReleaseNote,
				releaseNoteLine[1],
			)
		}
	}
	return releaseNotes
}
