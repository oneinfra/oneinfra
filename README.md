# oneinfra

`oneinfra` is a Kubernetes as a Service platform.

You can [read more about its design here](docs/DESIGN.md).

| Go Report                                                                                                                                      | Travis                                                                                                             | CircleCI                                                                                                             | Azure Test                                                                                                                                                                                    | Azure Release                                                                                                                                                                                       | License                                                                                                                              |
|------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra) | [![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=master)](https://travis-ci.org/oneinfra/oneinfra) | [![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=shield)](https://circleci.com/gh/oneinfra/oneinfra) | [![Test Pipeline](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=3&branchName=master) | [![Release Pipeline](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/release?branchName=master)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=4&branchName=master) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)|


## Managed Kubernetes versions

| Kubernetes version | Deployable with  | Default in       |                                                                                                                                                                            |                                                                                                                                                                             |
|--------------------|------------------|------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `1.15.11`          | `20.04.0-alpha1` |                  | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.15.11)%20with%20local%20CRI%20endpoints)        | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.15.11)%20with%20remote%20CRI%20endpoints)        |
| `1.16.8`           | `20.04.0-alpha1` |                  | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.16.8)%20with%20local%20CRI%20endpoints)         | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.16.8)%20with%20remote%20CRI%20endpoints)         |
| `1.17.4`           | `20.04.0-alpha1` |                  | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.17.4)%20with%20local%20CRI%20endpoints)         | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.17.4)%20with%20remote%20CRI%20endpoints)         |
| `1.18.0`           | `20.04.0-alpha1` | `20.04.0-alpha1` | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.18.0)%20with%20local%20CRI%20endpoints)         | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.18.0)%20with%20remote%20CRI%20endpoints)         |
| `1.19.0-alpha.1`   | `20.04.0-alpha1` |                  | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.19.0-alpha.1)%20with%20local%20CRI%20endpoints) | ![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=master&jobName=e2e%20tests%20(1.19.0-alpha.1)%20with%20remote%20CRI%20endpoints) |


## Go install

Build has been tested with go versions 1.13 and 1.14.

```
$ GO111MODULE=on go get github.com/oneinfra/oneinfra/...@master
```

This should have installed the following binaries:

* `oi-local-cluster`: allows you to test `oneinfra` locally in your
  machine, creating Docker containers as hypervisors.

* `oi`: CLI tool that allows you to test `oneinfra` locally in a
  standalone way, without requiring Kubernetes to store manifests.

* `oi-manager`: Kubernetes set of controllers that reconcile defined
  clusters.


## Quick start

### With Kubernetes as a management cluster

* Requirements
  * A Kubernetes cluster (will be our management cluster)
  * Docker (for creating fake local hypervisors)

[Install
`kind`](https://github.com/kubernetes-sigs/kind#installation-and-usage)
in order to try `oneinfra` in the easiest way. If you already have a
Kubernetes cluster you can use, you can skip this step.

```
$ kind create cluster
```

Deploy `cert-manager` and `oneinfra`.

```
$ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.14.1/cert-manager.yaml
$ kubectl wait --for=condition=Available deployment -l app.kubernetes.io/instance=cert-manager -n cert-manager
$ kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/master/config/generated/all.yaml
```

Create a fake local set of hypervisors, so we can schedule clusters
control plane components somewhere.

```
$ oi-local-cluster cluster create --remote | kubectl apply -f -
```

Now, create a managed cluster:

```
$ kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/master/config/samples/simple-cluster.yaml
$ kubectl wait --for=condition=ReconcileSucceeded cluster simple-cluster
$ kubectl get cluster simple-cluster -o yaml | oi cluster admin-kubeconfig > simple-cluster.conf
```

And access it:

```
$ kubectl --kubeconfig=simple-cluster.conf cluster-info
Kubernetes master is running at https://172.17.0.5:30000
```


### Without Kubernetes (for testing purposes only)

* Requirements
  * Docker

If you don't want to deploy Kubernetes to test `oneinfra`, you can try
the `oi` CLI tool that will allow you to test the reconciliation
processes of `oneinfra` without the need of a Kubernetes cluster.

```
$ mkdir ~/.kube
$ oi-local-cluster cluster create | \
    oi cluster inject --name simple-cluster | \
    oi component inject --name controlplane1 --role control-plane | \
    oi component inject --name controlplane2 --role control-plane | \
    oi component inject --name controlplane3 --role control-plane | \
    oi component inject --name loadbalancer --role control-plane-ingress | \
    oi reconcile | \
    tee simple-cluster.conf |  # so you can inspect the simple-cluster.conf afterwards :-) \
    oi cluster admin-kubeconfig > ~/.kube/config
```

And access it:

```
$ kubectl cluster-info
Kubernetes master is running at https://172.17.0.4:30000
```


## Joining worker nodes to a cluster

You can read more details about the [worker joining process
here](https://github.com/oneinfra/oneinfra/blob/master/docs/joining-worker-nodes.md).


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
