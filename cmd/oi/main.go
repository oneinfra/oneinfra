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
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/app/oi/cluster"
	"github.com/oneinfra/oneinfra/internal/app/oi/component"
	jointoken "github.com/oneinfra/oneinfra/internal/app/oi/join-token"
	"github.com/oneinfra/oneinfra/internal/app/oi/node"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
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
								Name:  "kubernetes-version",
								Usage: "kubernetes version, latest if not provided",
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
							kubernetesVersion := c.String("kubernetes-version")
							if len(kubernetesVersion) == 0 || kubernetesVersion == "latest" {
								kubernetesVersion = constants.LatestKubernetesVersion
							}
							return cluster.Inject(c.String("name"), kubernetesVersion, c.String("vpn-cidr"), c.StringSlice("apiserver-extra-sans"))
						},
					},
					{
						Name:  "admin-kubeconfig",
						Usage: "generate an admin kubeconfig file for the cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.AdminKubeConfig(c.String("cluster"))
						},
					},
					{
						Name:  "apiserver-ca",
						Usage: "prints the apiserver CA certificate",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.APIServerCA(c.String("cluster"))
						},
					},
					{
						Name:  "join-token-public-key",
						Usage: "prints the join token public key",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.JoinTokenPublicKey(c.String("cluster"))
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
					&cli.IntFlag{
						Name:  "max-retries",
						Usage: "max reconcile retry loops",
						Value: 5,
					},
					&cli.DurationFlag{
						Name:  "retry-wait-time",
						Usage: "time to wait between retries",
						Value: 5 * time.Second,
					},
				},
				Action: func(c *cli.Context) error {
					flagSet := flag.FlagSet{}
					klog.InitFlags(&flagSet)
					flagSet.Set("v", strconv.Itoa(c.Int("verbosity")))
					return cluster.Reconcile(c.Int("max-retries"), c.Duration("retry-wait-time"))
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
				Name:  "join-token",
				Usage: "join token operations",
				Subcommands: []*cli.Command{
					{
						Name:  "inject",
						Usage: "inject a join token",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "cluster",
								Required: true,
								Usage:    "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return jointoken.Inject(c.String("cluster"))
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
								Name:     "container-runtime-endpoint",
								Required: true,
								Usage:    "container runtime endpoint of this node",
							},
							&cli.StringFlag{
								Name:     "image-service-endpoint",
								Required: true,
								Usage:    "image service endpoint of this node",
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
							&cli.StringFlag{
								Name:     "join-token-public-key-file",
								Required: true,
								Usage:    "join token public key",
							},
						},
						Action: func(c *cli.Context) error {
							return node.Join(
								c.String("nodename"),
								c.String("apiserver-endpoint"),
								c.String("apiserver-ca-cert-file"),
								c.String("join-token"),
								c.String("join-token-public-key-file"),
								c.String("container-runtime-endpoint"),
								c.String("image-service-endpoint"),
							)
						},
					},
				},
			},
			{
				Name:  "version",
				Usage: "version information",
				Action: func(c *cli.Context) error {
					fmt.Println(constants.ReleaseData.Version)
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:  "kubernetes",
						Usage: "supported Kubernetes versions",
						Action: func(c *cli.Context) error {
							for _, kubernetesVersion := range constants.ReleaseData.KubernetesVersions {
								fmt.Println(kubernetesVersion.KubernetesVersion)
							}
							return nil
						},
					},
					{
						Name: "component",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "kubernetes-version",
								Usage: "Kubernetes version to inspect, latest if not provided",
							},
							&cli.StringFlag{
								Name:     "component",
								Required: true,
								Usage:    "component to inspect",
							},
						},
						Usage: "specific component version for the given Kubernetes version",
						Action: func(c *cli.Context) error {
							kubernetesVersion := c.String("kubernetes-version")
							if len(kubernetesVersion) == 0 || kubernetesVersion == "latest" {
								kubernetesVersion = constants.LatestKubernetesVersion
							}
							componentVersion, err := constants.KubernetesComponentVersion(kubernetesVersion, constants.Component(c.String("component")))
							if err != nil {
								return err
							}
							fmt.Println(componentVersion)
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
