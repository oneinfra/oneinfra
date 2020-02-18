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
    oi node inject --name controlplane --cluster cluster --role controlplane | \
    oi node inject --name loadbalancer --cluster cluster --role controlplane-ingress | \
    oi reconcile | \
    tee cluster.conf | \
    oi cluster kubeconfig --cluster cluster > ~/.kube/config
```

And access it:

```
$ kubectl cluster-info
Kubernetes master is running at https://127.0.0.1:30000
```

It's working! Let's degrain each step more in detail.

### Infrastructure creation

```
oi-local-cluster cluster create
```

Generates a number of hypervisors, each one running as a docker
container inside your machine. By default its size is three, and the
name of the hypervisor set is `test`. By default, one hypervisor is
public, two hypervisors are private.

> This command will output on `stdout` a versioned declaration of all
> hypervisors.

### Cluster injection

```
oi cluster inject --name cluster
```

Injects a versioned cluster with name `cluster`. Also generates a number
of certificate and certificate keys tied to the cluster.

You can inject as many clusters as you want, by piping them, as long
as they have different names.

> This command will take previous definitions from `stdin`, append the
> cluster definition and print everything to `stdout`.

### Node injection (control plane node)

```
oi node inject --name controlplane --cluster cluster --role controlplane
```

Each node represents a Kubernetes Master node.

Injects a versioned node with name `controlplane`, assigned to the
cluster with name `cluster`, created on the previous step.

You can inject as many nodes as you want, by piping them, as long as
they have different names for the same cluster.

> This command will take previous definitions from `stdin`, append the
> node definition and print everything to `stdout`.

### Node injection (control plane ingress node)

```
oi node inject --name loadbalancer --cluster cluster --role controlplane-ingress
```

Injects a Control Plane ingress (haproxy) instance, that will
automatically point to all the Kubernetes Masters linked to the
provided cluster.

### Infrastructure reconciliation

```
oi reconcile
```

This step will reconcile clusters and nodes passed by `stdin`,
effectively initializing your Kubernetes Master nodes.

> This command will take previous definitions from `stdin`, and print
> the updated infrastructure definition to `stdout`.

### Teeing

By running `tee cluster.conf` we are saving all our hypervisor, cluster
and node versioned resources into a file `cluster.conf`, so we can
inspect it for further reference, since we are relying on the CLI on
this example.

### KubeConfig generation

```
oi cluster kubeconfig --cluster cluster > ~/.kube/config
```

Since the `cluster.conf` contains the authoritative information about
our cluster, we can generate as many administrator `kubeconfig` files
as desired, based on client certificate authentication.

The KubeConfig file automatically points to the ingress endpoint.

> This command will print to `stdout` a `kubeconfig` file that is able
> to access to the cluster with name `cluster`.

### (Bonus track) Join workers

For development purposes, you might want to join worker nodes. Note
that `oneinfra` is focused on creating Control Plane instances as a
service, but in any case it can be handy to test joining some worker
nodes, specially for end to end and acceptance testing.

```
CLUSTER_CONF=cluster.conf CLUSTER_NAME=cluster scripts/create-fake-worker.sh
```

You can run this command as many times as you want. Every time, a new
worker node will join within seconds!
