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
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/oneinfra/oneinfra/internal/app/oi/cluster"
	jointoken "github.com/oneinfra/oneinfra/internal/app/oi/join-token"
	"github.com/oneinfra/oneinfra/internal/app/oi/node"
	"github.com/oneinfra/oneinfra/internal/pkg/constants"
	releasecomponents "github.com/oneinfra/oneinfra/internal/pkg/release-components"
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
								Name:  "name",
								Usage: "cluster name",
								Value: "cluster",
							},
							&cli.StringFlag{
								Name:  "kubernetes-version",
								Usage: "kubernetes version",
								Value: "default",
							},
							&cli.IntFlag{
								Name:  "control-plane-replicas",
								Usage: "Control plane number of replicas",
								Value: 1,
							},
							&cli.BoolFlag{
								Name:  "vpn-enabled",
								Usage: "CIDR used for the internal VPN",
								Value: false,
							},
							&cli.StringFlag{
								Name:  "vpn-cidr",
								Usage: "CIDR used for the internal VPN",
								Value: "10.0.0.0/16",
							},
							&cli.StringSliceFlag{
								Name:  "apiserver-extra-sans",
								Usage: "API server extra SANs",
							},
						},
						Action: func(c *cli.Context) error {
							kubernetesVersion := c.String("kubernetes-version")
							if kubernetesVersion == "default" {
								kubernetesVersion = constants.ReleaseData.DefaultKubernetesVersion
							}
							return cluster.Inject(c.String("name"), kubernetesVersion, c.Int("control-plane-replicas"), c.Bool("vpn-enabled"), c.String("vpn-cidr"), c.StringSlice("apiserver-extra-sans"))
						},
					},
					{
						Name:  "admin-kubeconfig",
						Usage: "generate an admin kubeconfig file for the cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "cluster",
								Usage: "cluster name (can be omitted if stdin has only one cluster resource)",
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
								Name:  "cluster",
								Usage: "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return cluster.APIServerCA(c.String("cluster"))
						},
					},
					{
						Name:  "version",
						Usage: "prints versioning information for the given cluster",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "cluster",
								Usage: "cluster name",
							},
						},
						Subcommands: []*cli.Command{
							{
								Name:  "kubernetes",
								Usage: "print the Kubernetes version for the given cluster",
								Action: func(c *cli.Context) error {
									kubernetesVersion, err := cluster.KubernetesVersion(c.String("cluster"))
									if err != nil {
										return err
									}
									fmt.Println(kubernetesVersion)
									return nil
								},
							},
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
				Name:  "join-token",
				Usage: "join token operations",
				Subcommands: []*cli.Command{
					{
						Name:  "inject",
						Usage: "inject a join token; prints resulting manifests in stdout, and the created join token in stderr",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "cluster",
								Usage: "cluster name",
							},
						},
						Action: func(c *cli.Context) error {
							return jointoken.Inject(c.String("cluster"))
						},
					},
					{
						Name:  "generate",
						Usage: "generates a random join token and prints it to stdout",
						Action: func(c *cli.Context) error {
							return jointoken.Generate()
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
							&cli.IntFlag{
								Name:    "verbosity",
								Aliases: []string{"v"},
								Usage:   "logging verbosity",
								Value:   1,
							},
							&cli.StringFlag{
								Name:     "nodename",
								Required: true,
								Usage:    "node name of this node when joining",
							},
							&cli.StringFlag{
								Name:  "container-runtime-endpoint",
								Usage: "container runtime endpoint of this node",
								Value: "unix:///run/containerd/containerd.sock",
							},
							&cli.StringFlag{
								Name:  "image-service-endpoint",
								Usage: "image service endpoint of this node",
								Value: "unix:///run/containerd/containerd.sock",
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
							&cli.StringSliceFlag{
								Name:  "extra-san",
								Usage: "extra Subject Alternative Names (SAN's) for the Kubelet server certificate. You can provide this argument many times.",
							},
						},
						Action: func(c *cli.Context) error {
							flagSet := flag.FlagSet{}
							klog.InitFlags(&flagSet)
							flagSet.Set("v", strconv.Itoa(c.Int("verbosity")))
							return node.Join(
								c.String("nodename"),
								c.String("apiserver-endpoint"),
								c.String("apiserver-ca-cert-file"),
								c.String("join-token"),
								c.String("container-runtime-endpoint"),
								c.String("image-service-endpoint"),
								c.StringSlice("extra-san"),
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
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "default",
								Usage: "print the default kubernetes version",
							},
						},
						Action: func(c *cli.Context) error {
							if c.Bool("default") {
								fmt.Println(constants.ReleaseData.DefaultKubernetesVersion)
							} else {
								for _, kubernetesVersion := range constants.ReleaseData.KubernetesVersions {
									fmt.Println(kubernetesVersion.Version)
								}
							}
							return nil
						},
					},
					{
						Name: "component",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "kubernetes-version",
								Usage: "Kubernetes version to inspect",
								Value: "default",
							},
							&cli.StringFlag{
								Name:     "component",
								Required: true,
								Usage:    fmt.Sprintf("component to inspect (components: %s) (test components: %s)", releasecomponents.KubernetesComponents, releasecomponents.KubernetesTestComponents),
							},
						},
						Usage: "specific component version for the given Kubernetes version",
						Action: func(c *cli.Context) error {
							kubernetesVersion := c.String("kubernetes-version")
							if kubernetesVersion == "default" {
								kubernetesVersion = constants.ReleaseData.DefaultKubernetesVersion
							}
							if componentVersion, err := constants.KubernetesComponentVersion(kubernetesVersion, releasecomponents.KubernetesComponent(c.String("component"))); err == nil {
								fmt.Println(componentVersion)
								return nil
							}
							testComponentVersion, err := constants.KubernetesTestComponentVersion(kubernetesVersion, releasecomponents.KubernetesTestComponent(c.String("component")))
							if err != nil {
								return err
							}
							fmt.Println(testComponentVersion)
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
