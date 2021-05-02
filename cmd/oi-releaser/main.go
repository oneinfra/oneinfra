/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
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
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"k8s.io/klog/v2"

	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/binaries"
	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/git"
	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/images"
	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/pipelines"
	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/text"
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
								Usage: "binaries to build; can be provided several times",
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
								Usage: "binaries to publish; can be provided several times",
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
								Usage: "images to build; can be provided several times in the form of image:version, all if not provided or empty",
							},
							&cli.BoolFlag{
								Name:  "force",
								Value: false,
								Usage: "force building images, even if they exist",
							},
						},
						Action: func(c *cli.Context) error {
							images.BuildContainerImages(
								chosenContainerImages(c.StringSlice("image")),
								c.Bool("force"),
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
								Usage: "images to publish; can be provided several times in the form of image:version, all if not provided or empty",
							},
							&cli.BoolFlag{
								Name:  "force",
								Value: false,
								Usage: "force building images, even if they exist",
							},
						},
						Action: func(c *cli.Context) error {
							images.PublishContainerImages(
								chosenContainerImages(c.StringSlice("image")),
								c.Bool("force"),
							)
							return nil
						},
					},
				},
			},
			{
				Name:  "text",
				Usage: "text operations",
				Subcommands: []*cli.Command{
					{
						Name:  "replace-placeholders",
						Usage: "replace placeholders on text provided through stdin and print result to stdout",
						Action: func(c *cli.Context) error {
							stdin, err := ioutil.ReadAll(os.Stdin)
							if err != nil {
								return err
							}
							fmt.Print(text.ReplacePlaceholders(string(stdin)))
							return nil
						},
					},
				},
			},
			{
				Name:  "git",
				Usage: "git operations",
				Subcommands: []*cli.Command{
					{
						Name:  "release-notes",
						Usage: "given a log of commits, extract the release notes marked with release-note block",
						Action: func(c *cli.Context) error {
							stdin, err := ioutil.ReadAll(os.Stdin)
							if err != nil {
								return err
							}
							for _, releaseNote := range git.ReleaseNotes(string(stdin)) {
								fmt.Printf("- :heavy_check_mark: %s\n", releaseNote.Commit)
								fmt.Println("    ```")
								inputBuffer := bytes.NewBufferString(releaseNote.ReleaseNote)
								scanner := bufio.NewScanner(inputBuffer)
								scanner.Split(bufio.ScanLines)
								for scanner.Scan() {
									fmt.Printf("    %s\n", scanner.Text())
								}
								fmt.Println("    ```")
								fmt.Println()
							}
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
					{
						Name:  "publish-tooling-images",
						Usage: "publish tooling images pipeline operations",
						Subcommands: []*cli.Command{
							{
								Name:  "dump",
								Usage: "dump the publish tooling images pipeline to stdout",
								Action: func(c *cli.Context) error {
									return pipelines.AzurePublishToolingImages()
								},
							},
						},
					},
					{
						Name:  "publish-nightly-images",
						Usage: "publish nightly images pipeline operations",
						Subcommands: []*cli.Command{
							{
								Name:  "dump",
								Usage: "dump the publish nightly images pipeline to stdout",
								Action: func(c *cli.Context) error {
									return pipelines.AzurePublishNightlyImages()
								},
							},
						},
					},
					{
						Name:  "publish-testing-images",
						Usage: "publish testing images pipeline operations",
						Subcommands: []*cli.Command{
							{
								Name:  "dump",
								Usage: "dump the publish testing images pipeline to stdout",
								Action: func(c *cli.Context) error {
									return pipelines.AzurePublishTestingImages()
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

func chosenContainerImages(containerImages []string) images.ContainerImageList {
	if len(containerImages) == 1 && containerImages[0] == "" {
		return images.ContainerImageList{}
	}
	// Used to avoid image:version duplicates
	chosenContainerImageMap := map[string]map[string]struct{}{}
	for _, chosenContainerImage := range containerImages {
		imageSplit := strings.Split(chosenContainerImage, ":")
		if len(imageSplit) != 2 {
			klog.Fatalf("could not parse %q as image:tag", chosenContainerImage)
		}
		imageName, imageVersion := imageSplit[0], imageSplit[1]
		if chosenContainerImageMap[imageName] == nil {
			chosenContainerImageMap[imageName] = map[string]struct{}{}
		}
		if _, exists := chosenContainerImageMap[imageName][imageVersion]; exists {
			continue
		}
		chosenContainerImageMap[imageName][imageVersion] = struct{}{}
	}
	chosenContainerImages := images.ContainerImageList{}
	for containerImage, containerVersions := range chosenContainerImageMap {
		containerImageWithTags := images.ContainerImageWithTags{
			Image: images.ContainerImage(containerImage),
			Tags:  []string{},
		}
		for containerVersion := range containerVersions {
			containerImageWithTags.Tags = append(
				containerImageWithTags.Tags,
				containerVersion,
			)
		}
		chosenContainerImages = append(
			chosenContainerImages,
			containerImageWithTags,
		)
	}
	return chosenContainerImages
}
