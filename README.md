# oneinfra

[![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra)
[![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=master)](https://travis-ci.org/oneinfra/oneinfra)
[![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=shield)](https://circleci.com/gh/oneinfra/oneinfra)
[![Azure Pipelines](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/oneinfra.oneinfra?branchName=master)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=1&branchName=master)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

![oneinfra logo](logos/oneinfra.png)

`oneinfra` is a Kubernetes as a Service platform, or KaaS.

It features a declarative infrastructure definition.

You can read more about its [design here](docs/DESIGN.md).

## Install

```
$ GO111MODULE=on go get github.com/oneinfra/oneinfra/...@master
```

This should have installed the following binaries:

* `oi-local-cluster`: allows you to test `oneinfra` locally in your
  machine, with docker containers as hypervisors.

* `oi`: CLI tool that allows you to discover `oneinfra`.

* `oi-manager`: Kubernetes set of controllers that reconcile your
  `oneinfra` defined clusters.

## Quick start with docker

You can create hypervisors in your machine with docker. Each docker
container will resemble a physical or virtual hypervisor in your
infrastructure.

```
$ oi-local-cluster cluster create | \
    oi cluster inject --name cluster | \
    oi component inject --name controlplane1 --cluster cluster --role controlplane | \
    oi component inject --name controlplane2 --cluster cluster --role controlplane | \
    oi component inject --name controlplane3 --cluster cluster --role controlplane | \
    oi component inject --name loadbalancer --cluster cluster --role controlplane-ingress | \
    oi reconcile | \
    tee cluster.conf | \
    oi cluster admin-kubeconfig --cluster cluster > ~/.kube/config
```

And access it:

```
$ kubectl cluster-info
Kubernetes master is running at https://127.0.0.1:30000
```

## Features missing for the first release cut

`oneinfra` is a heavy Work In Progress, and not intended for
production yet.

- [X] Implement initial worker join logic
- [X] Extend worker join to wait for `issued` condition
- [X] Finish worker node join logic
- [ ] Provide hypervisor VM images ([WIP](https://github.com/oneinfra/oneinfra/compare/master...vm-images))
- [ ] Provide worker VM images ([WIP](https://github.com/oneinfra/oneinfra/compare/master...vm-images))

## License

`oneinfra` is licensed under the terms of the Apache 2.0 license.

```
Copyright (C) 2020 Rafael Fernández López <ereslibre@ereslibre.es>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
