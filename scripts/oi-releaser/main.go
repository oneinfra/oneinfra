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

	"github.com/urfave/cli/v2"

	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	"github.com/oneinfra/oneinfra/scripts/oi-releaser/images"
	"github.com/oneinfra/oneinfra/scripts/oi-releaser/pipelines"
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
							images.BuildContainerImages(
								kubernetesVersions(),
								chosenContainerImages(c.StringSlice("images")),
							)
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
							images.PublishContainerImages(
								kubernetesVersions(),
								chosenContainerImages(c.StringSlice("images")),
							)
							return nil
						},
					},
				},
			},
			{
				Name:  "test-pipeline",
				Usage: "test pipeline operations",
				Subcommands: []*cli.Command{
					{
						Name:  "dump",
						Usage: "dump the test pipeline to stdout",
						Action: func(c *cli.Context) error {
							return pipelines.AzureTest(
								kubernetesVersions(),
							)
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
