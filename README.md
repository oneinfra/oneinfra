# oneinfra

[![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra)
[![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=master)](https://travis-ci.org/oneinfra/oneinfra)
[![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=shield)](https://circleci.com/gh/oneinfra/oneinfra)
[![Azure Pipelines](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/oneinfra.oneinfra?branchName=master)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=1&branchName=master)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

![oneinfra logo](logos/oneinfra.png)

`oneinfra` is a Kubernetes Masters as a Service platform, or KMaaS.

It features a declarative infrastructure definition.

You can read more about its [design here](docs/DESIGN.md).

## Quick start with docker

You can create hypervisors in your machine with docker. Each docker
container will resemble a physical or virtual hypervisor in your
infrastructure.

```
$ oi-local-cluster cluster create | \
    oi cluster inject --name cluster | \
    oi component inject --name controlplane --cluster cluster --role controlplane | \
    oi component inject --name loadbalancer --cluster cluster --role controlplane-ingress | \
    oi reconcile | \
    tee cluster.conf | \
    oi cluster kubeconfig --cluster cluster > ~/.kube/config
```

And access it:

```
$ kubectl cluster-info
Kubernetes master is running at https://127.0.0.1:30000
```
