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

package binaries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/pkg/errors"
)

var (
	architectures = []string{"amd64", "arm", "arm64"}
)

// GithubRelease represents a Github release
type GithubRelease struct {
	ID int `json:"id"`
}

// BuildBinaries builds the provided binaries in a release matching
// the current tag
func BuildBinaries(binaries []string) error {
	if err := cleanBinDirectory(); err != nil {
		return err
	}
	return buildBinaries(binaries, tagName())
}

// PublishBinaries publishes the provided binaries in a release
// matching the current tag
func PublishBinaries(binaries []string) error {
	if err := BuildBinaries(binaries); err != nil {
		return err
	}
	releaseID, err := fetchReleaseID(tagName())
	if err != nil {
		return errors.Wrapf(err, "could not fetch release id for tag %q", tagName())
	}
	for _, architecture := range architectures {
		for _, binary := range binaries {
			binaryName := binaryName(binary, architecture, tagName())
			if err := publishBinary(binaryName, releaseID); err != nil {
				klog.Errorf("could not publish binary %q: %v", binaryName, err)
				continue
			}
			klog.Infof("binary %q successfully published in release %q", binaryName, tagName())
		}
	}
	return nil
}

func binaryName(binary, architecture, tagName string) string {
	return fmt.Sprintf("%s-linux-%s-%s", binary, architecture, tagName)
}

func fetchReleaseID(tagName string) (string, error) {
	requestURL := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   path.Join("/repos/oneinfra/oneinfra/releases/tags", tagName),
	}
	req, _ := http.NewRequest(http.MethodGet, requestURL.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	rawContent, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var githubRelease GithubRelease
	if err := json.Unmarshal(rawContent, &githubRelease); err != nil {
		return "", err
	}
	return strconv.Itoa(githubRelease.ID), nil
}

func buildBinaries(binaries []string, tagName string) error {
	for _, architecture := range architectures {
		for _, binary := range binaries {
			binaryName := binaryName(binary, architecture, tagName)
			cmd := exec.Command(
				"go",
				"build",
				"-ldflags", fmt.Sprintf("-X github.com/oneinfra/oneinfra/internal/pkg/constants.BuildVersion=%s", constants.BuildVersion),
				"-mod", "vendor",
				"-o", filepath.Join("bin", binaryName),
				fmt.Sprintf("github.com/oneinfra/oneinfra/cmd/%s", binary),
			)
			cmd.Env = os.Environ()
			cmd.Env = append(
				cmd.Env,
				"CGO_ENABLED=0",
				"GOOS=linux",
				fmt.Sprintf("GOARCH=%s", architecture),
				"GO111MODULE=on",
			)
			if err := cmd.Run(); err != nil {
				klog.Errorf("failed to build %q binary: %v", binaryName, err)
				continue
			}
			klog.Infof("binary %q successfully built", binaryName)
		}
	}
	return nil
}

func publishBinary(binaryName, releaseID string) error {
	binaryContents, err := os.Open(filepath.Join("bin", binaryName))
	if err != nil {
		return err
	}
	requestURL := url.URL{
		Scheme:   "https",
		Host:     "uploads.github.com",
		Path:     path.Join("/repos/oneinfra/oneinfra/releases", releaseID, "assets"),
		RawQuery: fmt.Sprintf("name=%s", binaryName),
	}
	req, err := http.NewRequest(http.MethodPost, requestURL.String(), binaryContents)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(filepath.Join("bin", binaryName))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")))
	req.Header.Set("Content-Type", "application/x-executable")
	req.ContentLength = fileInfo.Size()
	_, err = http.DefaultClient.Do(req)
	return err
}

func tagName() string {
	cmd := exec.Command("git", "describe", "--tags")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		klog.Fatalf("failed to execute command: %v", err)
	}
	return strings.Trim(stdout.String(), " \n")
}

func cleanBinDirectory() error {
	binDir, err := ioutil.ReadDir("bin")
	if err != nil {
		return err
	}
	for _, binDirFile := range binDir {
		if err := os.RemoveAll(filepath.Join("bin", binDirFile.Name())); err != nil {
			return err
		}
	}
	return nil
}
