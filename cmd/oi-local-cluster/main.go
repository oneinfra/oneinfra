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
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/oneinfra/oneinfra/internal/app/oi-local-cluster/cluster"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
)

func main() {
	app := &cli.App{
		Usage: "manage test clusters",
		Commands: []*cli.Command{
			{
				Name:  "cluster",
				Usage: "test cluster operations",
				Subcommands: []*cli.Command{
					{
						Name:  "create",
						Usage: "create a test cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Value: "test",
								Usage: "test cluster name",
							},
							&cli.StringFlag{
								Name:  "kubernetes-version",
								Value: "default",
								Usage: "expected Kubernetes version to be deployed (will contain all requirements already installed)",
							},
							&cli.IntFlag{
								Name:  "size-private",
								Value: 1,
								Usage: "private hypervisor test cluster size",
							},
							&cli.IntFlag{
								Name:  "size-public",
								Value: 2,
								Usage: "public hypervisor test cluster size",
							},
							&cli.BoolFlag{
								Name:  "remote",
								Value: false,
								Usage: "whether remote hypervisors (CRI exposed through TCP) will be used. If not, local hypervisors (CRI exposed through a UNIX socket) will be used",
							},
						},
						Action: func(c *cli.Context) error {
							kubernetesVersion := c.String("kubernetes-version")
							if len(kubernetesVersion) == 0 || kubernetesVersion == "default" {
								kubernetesVersion = constants.ReleaseData.DefaultKubernetesVersion
							}
							return cluster.Create(c.String("name"), kubernetesVersion, c.Int("size-private"), c.Int("size-public"), c.Bool("remote"))
						},
					},
					{
						Name:  "endpoint",
						Usage: "retrieves a test cluster endpoint",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "test cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.Endpoint(c.String("name"))
						},
					},
					{
						Name:  "destroy",
						Usage: "destroy a test cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Value: "test",
								Usage: "test cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.Destroy(c.String("name"))
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
