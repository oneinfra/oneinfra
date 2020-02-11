# oneinfra

[![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra)
[![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=master)](https://travis-ci.org/oneinfra/oneinfra)
[![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=svg)](https://circleci.com/gh/oneinfra/oneinfra)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-green.svg)](https://opensource.org/licenses/Apache-2.0)

`oneinfra` is a Kubernetes Masters as a Service platform, or KMaaS.

It features a declarative infrastructure declaration.

`oneinfra` relies on a number of hypervisors exposing a `CRI`
endpoint. `oneinfra` will only connect to this `CRI` endpoint in order
to create the Kubernetes Master nodes.

## Quick start with docker

You can create hypervisors in your machine with docker. Each docker
container will resemble a physical or virtual hypervisor in your
infrastructure.

```
$ oi-local-cluster cluster create | oi cluster inject --name test | oi node inject --name test --cluster test | tee cluster.txt | oi reconcile
$ cat cluster.txt | oi cluster kubeconfig --cluster test > ~/.kube/config
$ kubectl cluster-info
Kubernetes master is running at https://127.0.0.1:30100
```

And it's working. Let's degrain each step more in detail.

### Infrastructure creation

```
oi-local-cluster cluster create
```

Generates a number of hypervisors, each one running as a docker
container inside your machine. By default its size is three, and the
name of the hypervisor set is `test`, you can tweak these though.

> This command will output on `stdout` a versioned declaration of all
> the hypervisors.

### Cluster injection

```
oi cluster inject --name test
```

Injects a versioned cluster with name `test`. Also generates a number
of certificate and certificate keys tied to the cluster.

You can inject as many clusters as you want, by piping them, as long
as they have different names.

> This command will take previous definitions from `stdin`, append the
> cluster definition and print everything to `stdout`.

### Node injection

```
oi node inject --name test --cluster test
```

Each node represents a Kubernetes Master node.

Injects a versioned node with name `test`, assigned to the cluster
with name `test`, created on the previous step

You can inject as many nodes as you want, by piping them, as long as
they have different names.

> This command will take previous definitions from `stdin`, append the
> node definition and print everything to `stdout`.

### Teeing

By running `tee cluster.txt` we are saving all our hypervisor, cluster
and node versioned resources into a file `cluster.txt`, so we can use
it afterwards, since the next step will stop forwarding `stdin` to
`stdout`.

### Infrastructure reconciliation

```
oi reconcile
```

This step will reconcile clusters and nodes passed by `stdin`,
effectively initializing your Kubernetes Master nodes.

### KubeConfig generation

```
cat cluster.txt | oi cluster kubeconfig --cluster test > ~/.kube/config
```

Since the `cluster.txt` contains the authoritative information about
our cluster, we can generate as many administrator `kubeconfig` files
as desired, based on client certificate authentication.
