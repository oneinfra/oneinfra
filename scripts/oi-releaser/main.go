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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

var (
	allContainerImages = []string{
		"containerd",
		"hypervisor",
		"kubelet-installer",
		"oi",
		"oi-manager",
	}
)

func main() {
	app := &cli.App{
		Usage: "oneinfra releaser CLI tool",
		Commands: []*cli.Command{
			{
				Name:  "container-images",
				Usage: "container images operations",
				Subcommands: []*cli.Command{
					{
						Name:  "build",
						Usage: "build all container image artifacts",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "image",
								Usage: fmt.Sprintf("images to build %v; can be provided several times, all if not provided", allContainerImages),
							},
						},
						Action: func(c *cli.Context) error {
							buildContainerImages(chosenContainerImages(c.StringSlice("images")))
							return nil
						},
					},
					{
						Name:  "publish",
						Usage: "publish all container image artifacts",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "image",
								Usage: fmt.Sprintf("images to publish %v; can be provided several times, all if not provided", allContainerImages),
							},
						},
						Action: func(c *cli.Context) error {
							publishContainerImages(chosenContainerImages(c.StringSlice("images")))
							return nil
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func buildContainerImages(containerImages []string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not read current working directory: %v", err)
	}
	kubernetesVersions := kubernetesVersions()
	for _, containerImage := range containerImages {
		if err := os.Chdir(filepath.Join(cwd, "images", containerImage)); err != nil {
			log.Fatalf("could not change directory: %v", err)
		}
		for _, kubernetesVersion := range kubernetesVersions {
			cmd := exec.Command("make", "image")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			setCmdEnv(cmd, kubernetesVersion)
			if err := cmd.Run(); err != nil {
				log.Printf("failed to build %q image: %v", containerImage, err)
			}
		}
	}
}

func publishContainerImages(containerImages []string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not read current working directory: %v", err)
	}
	kubernetesVersions := kubernetesVersions()
	for _, containerImage := range containerImages {
		if err := os.Chdir(filepath.Join(cwd, "images", containerImage)); err != nil {
			log.Fatalf("could not change directory: %v", err)
		}
		for _, kubernetesVersion := range kubernetesVersions {
			cmd := exec.Command("make", "publish")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			setCmdEnv(cmd, kubernetesVersion)
			if err := cmd.Run(); err != nil {
				log.Printf("failed to publish %q image: %v", containerImage, err)
			}
		}
	}
}

func setCmdEnv(cmd *exec.Cmd, kubernetesVersion constants.KubernetesVersion) {
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("ONEINFRA_VERSION=%s", constants.ReleaseData.Version),
		fmt.Sprintf("KUBERNETES_VERSION=%s", kubernetesVersion.KubernetesVersion),
		fmt.Sprintf("CRI_TOOLS_VERSION=%s", kubernetesVersion.CRIToolsVersion),
		fmt.Sprintf("CONTAINERD_VERSION=%s", kubernetesVersion.ContainerdVersion),
		fmt.Sprintf("CNI_PLUGINS_VERSION=%s", kubernetesVersion.CNIPluginsVersion),
		fmt.Sprintf("ETCD_VERSION=%s", kubernetesVersion.EtcdVersion),
		fmt.Sprintf("PAUSE_VERSION=%s", kubernetesVersion.PauseVersion),
	}...)
}

func kubernetesVersions() []constants.KubernetesVersion {
	return constants.ReleaseData.KubernetesVersions
}

func chosenContainerImages(containerImages []string) []string {
	if len(containerImages) == 0 {
		return allContainerImages
	}
	chosenContainerImages := map[string]struct{}{}
	for _, chosenContainerImage := range containerImages {
		chosenContainerImages[chosenContainerImage] = struct{}{}
	}
	res := []string{}
	for _, containerImage := range allContainerImages {
		if _, exists := chosenContainerImages[containerImage]; exists {
			res = append(res, containerImage)
		}
	}
	return res
}
