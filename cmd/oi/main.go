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

	"oneinfra.ereslibre.es/m/internal/app/oi/cluster"
	"oneinfra.ereslibre.es/m/internal/app/oi/node"
)

func main() {
	app := &cli.App{
		Usage: "oneinfra CLI tool",
		Commands: []*cli.Command{
			{
				Name:  "cluster",
				Usage: "cluster operations",
				Subcommands: []*cli.Command{
					{
						Name:  "inject",
						Usage: "inject a cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.Inject(c.String("name"))
						},
					},
					{
						Name:  "kubeconfig",
						Usage: "generate a kubeconfig file for the cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.KubeConfig(c.String("cluster"))
						},
					},
				},
			},
			{
				Name:  "reconcile",
				Usage: "reconcile all clusters",
				Action: func(c *cli.Context) error {
					return cluster.Reconcile()
				},
			},
			{
				Name:  "node",
				Usage: "node operations",
				Subcommands: []*cli.Command{
					{
						Name:  "inject",
						Usage: "inject a node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
								Usage:    "node name",
							},
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
							&cli.StringFlag{
								Name:     "role",
								Required: true,
								Usage:    "role of the node (controlplane, gater)",
							},
						},
						Action: func(c *cli.Context) error {
							return node.Inject(c.String("name"), c.String("cluster"), c.String("role"))
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
