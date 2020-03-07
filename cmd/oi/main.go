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
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/app/oi/cluster"
	"github.com/oneinfra/oneinfra/internal/app/oi/component"
	"github.com/oneinfra/oneinfra/internal/app/oi/node"
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
							&cli.StringFlag{
								Name:  "vpn-cidr",
								Usage: "CIDR used for the internal VPN",
								Value: "10.0.0.0/8",
							},
							&cli.StringSliceFlag{
								Name:  "apiserver-extra-sans",
								Usage: "API server extra SANs",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.Inject(c.String("name"), c.String("vpn-cidr"), c.StringSlice("apiserver-extra-sans"))
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
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "verbosity",
						Aliases: []string{"v"},
						Usage:   "logging verbosity",
						Value:   1,
					},
				},
				Action: func(c *cli.Context) error {
					flagSet := flag.FlagSet{}
					klog.InitFlags(&flagSet)
					flagSet.Set("v", strconv.Itoa(c.Int("verbosity")))
					return cluster.Reconcile()
				},
			},
			{
				Name:  "component",
				Usage: "component operations",
				Subcommands: []*cli.Command{
					{
						Name:  "inject",
						Usage: "inject a component",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: true,
								Usage:    "component name",
							},
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
							&cli.StringFlag{
								Name:     "role",
								Required: true,
								Usage:    "role of the component (controlplane, controlplane-ingress)",
							},
						},
						Action: func(c *cli.Context) error {
							return component.Inject(c.String("name"), c.String("cluster"), c.String("role"))
						},
					},
				},
			},
			{
				Name:  "node",
				Usage: "node operations",
				Subcommands: []*cli.Command{
					{
						Name:  "join",
						Usage: "joins a node to an existing cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "nodename",
								Required: true,
								Usage:    "node name of this node when joining",
							},
							&cli.StringFlag{
								Name:     "apiserver-endpoint",
								Required: true,
								Usage:    "endpoint of the apiserver to join to",
							},
							&cli.StringFlag{
								Name:     "apiserver-ca-cert-file",
								Required: true,
								Usage:    "apiserver CA certificate to check the apiserver identity",
							},
							&cli.StringFlag{
								Name:     "join-token",
								Required: true,
								Usage:    "token to use for joining",
							},
						},
						Action: func(c *cli.Context) error {
							return node.Join(c.String("nodename"), c.String("apiserver-endpoint"), c.String("apiserver-ca-cert-file"), c.String("join-token"))
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
