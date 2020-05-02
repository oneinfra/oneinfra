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

package main

import (
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/scripts/oi-releaser/binaries"
	"github.com/oneinfra/oneinfra/scripts/oi-releaser/images"
	"github.com/oneinfra/oneinfra/scripts/oi-releaser/pipelines"
)

func main() {
	app := &cli.App{
		Usage: "oneinfra releaser CLI tool",
		Commands: []*cli.Command{
			{
				Name:  "binaries",
				Usage: "binaries operations",
				Subcommands: []*cli.Command{
					{
						Name:  "build",
						Usage: "build binaries",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "binary",
								Usage: "binaries to build; can be provided several times, all if not provided",
							},
						},
						Action: func(c *cli.Context) error {
							return binaries.BuildBinaries(c.StringSlice("binary"))
						},
					},
					{
						Name:  "publish",
						Usage: "publish binaries",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "binary",
								Usage: "binaries to publish; can be provided several times, all if not provided",
							},
						},
						Action: func(c *cli.Context) error {
							return binaries.PublishBinaries(c.StringSlice("binary"))
						},
					},
				},
			},
			{
				Name:  "container-images",
				Usage: "container images operations",
				Subcommands: []*cli.Command{
					{
						Name:  "build",
						Usage: "build container image artifacts",
						Flags: []cli.Flag{
							&cli.StringSliceFlag{
								Name:  "image",
								Usage: "images to build; can be provided several times in the form of image:version, all if not provided",
							},
						},
						Action: func(c *cli.Context) error {
							images.BuildContainerImages(
								chosenContainerImages(c.StringSlice("image")),
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
								Usage: "images to publish; can be provided several times in the form of image:version, all if not provided",
							},
						},
						Action: func(c *cli.Context) error {
							images.PublishContainerImages(
								chosenContainerImages(c.StringSlice("image")),
							)
							return nil
						},
					},
				},
			},
			{
				Name:  "pipelines",
				Usage: "pipeline operations",
				Subcommands: []*cli.Command{
					{
						Name:  "test",
						Usage: "test pipeline operations",
						Subcommands: []*cli.Command{
							{
								Name:  "dump",
								Usage: "dump the test pipeline to stdout",
								Action: func(c *cli.Context) error {
									return pipelines.AzureTest()
								},
							},
						},
					},
					{
						Name:  "release",
						Usage: "release pipeline operations",
						Subcommands: []*cli.Command{
							{
								Name:  "dump",
								Usage: "dump the release pipeline to stdout",
								Action: func(c *cli.Context) error {
									return pipelines.AzureRelease()
								},
							},
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

func chosenContainerImages(containerImages []string) images.ContainerImageMapWithTags {
	chosenContainerImages := images.ContainerImageMapWithTags{}
	// Used to avoid image:version duplicates
	chosenContainerImageMap := map[string]map[string]struct{}{}
	for _, chosenContainerImage := range containerImages {
		imageSplit := strings.Split(chosenContainerImage, ":")
		if len(imageSplit) != 2 {
			klog.Fatalf("could not parse %q as image:tag", chosenContainerImage)
		}
		imageName, imageVersion := imageSplit[0], imageSplit[1]
		if chosenContainerImageMap[imageName] == nil {
			chosenContainerImages[images.ContainerImage(imageName)] = []string{}
			chosenContainerImageMap[imageName] = map[string]struct{}{}
		}
		if _, exists := chosenContainerImageMap[imageName][imageVersion]; exists {
			continue
		}
		chosenContainerImages[images.ContainerImage(imageName)] = append(
			chosenContainerImages[images.ContainerImage(imageName)],
			imageVersion,
		)
		chosenContainerImageMap[imageName][imageVersion] = struct{}{}
	}
	return chosenContainerImages
}
