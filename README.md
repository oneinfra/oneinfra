# oneinfra

[![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra)
[![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=master)](https://travis-ci.org/oneinfra/oneinfra)
[![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=shield)](https://circleci.com/gh/oneinfra/oneinfra)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

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
$ oi-local-cluster cluster create | \
    oi cluster inject --name test | \
    oi node inject --name test --cluster test --role controlplane | \
    oi node inject --name loadbalancer --cluster test --role controlplane-ingress | \
    tee cluster.txt | \
    oi reconcile
```

Generate a kubeconfig file for your cluster:

```
$ cat cluster.txt | oi cluster kubeconfig --cluster test --endpoint-host-override 127.0.0.1 > ~/.kube/config
```

And access it:

```
$ kubectl cluster-info
Kubernetes master is running at https://127.0.0.1:30100
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
oi node inject --name test --cluster test --role controlplane
```

Each node represents a Kubernetes Master node.

Injects a versioned node with name `test`, assigned to the cluster
with name `test`, created on the previous step.

You can inject as many nodes as you want, by piping them, as long as
they have different names for the same cluster.

> This command will take previous definitions from `stdin`, append the
> node definition and print everything to `stdout`.

### Node injection (take two)

```
oi node inject --name loadbalancer --cluster test --role controlplane-ingress
```

Injects a Control Plane ingress (haproxy) instance, that will
automatically point to all the Kubernetes Masters linked to the
provided cluster.

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
cat cluster.txt | oi cluster kubeconfig --cluster test --endpoint-host-override 127.0.0.1 > ~/.kube/config
```

Since the `cluster.txt` contains the authoritative information about
our cluster, we can generate as many administrator `kubeconfig` files
as desired, based on client certificate authentication.

The KubeConfig file automatically points to the ingress endpoint. In
this case, we are also overriding the API endpoint host. This is to
mimic that we are accessing to the public hypervisor through the
internet facing interface (as opposed to the interface connected to
the private network and hypervisors -- on the docker lab, the IP
address of the container).

Had we omitted this argument, we could connect normally as well:

```
cat cluster.txt | oi cluster kubeconfig --cluster test > ~/.kube/config
```

But it does not reflect the suggested networking schema.

> This command will print to `stdout` a `kubeconfig` file that is able
> to access to the cluster with name `test`.

### (Bonus track) Join workers

For development purposes, you might want to join worker nodes. Note
that `oneinfra` is focused on creating Control Plane instances as a
service, but in any case it can be handy to test joining some worker
nodes, specially for end to end and acceptance testing.

```
CLUSTER_CONF=cluster.txt CLUSTER_NAME=test scripts/create-fake-worker.sh
```

You can run this command as many times as you want. Every time, a new
worker node will join within seconds!
