# Goal

`oneinfra`'s goal is to provide a Kubernetes Master as a Service
platform. It allows you to manage a number of Kubernetes Control
Planes in a set of hypervisors sitting in your infrastructure, on
premises or in the public cloud.

* Easily create isolated Kubernetes control planes
* Manage control planes with a well defined API
* Edge ready (TODO)
  * Allow to safely join worker nodes, regardless of where they are
    with respect to the control plane nodes (e.g. workers behind several
    NAT levels)
  * CNI overlay will be set on top of the VPN tunnel
* Operating System agnostic
* Provide both a CLI and a set of controllers (TODO) to drive the
  reconciliation process
  * This makes it easier to either use the CLI directly, or deploy
    `oneinfra` controllers and custom resources on top of an existing
    Kubernetes cluster (management cluster)
  * Both CLI and these set of controllers should rely on the same
    API's (custom resources) to drive the reconciliation process

# Components

## Hypervisor

An hypervisor is a physical or virtual machine where `oneinfra` will
create the control plane nodes.

Hypervisors are user provided in the form of a CRI endpoint to connect
to.

### Public and private hypervisors

`oneinfra` makes the differentiation between public and private
hypervisors.

#### Public hypervisors

Public hypervisors is where the ingress for the control plane nodes
will live. This is, a load balancer (or set of load balancers), and a
VPN server.

Public hypervisors are therefore exposed to the outer world, and
should have an interface inside a public network. They should also
have an interface inside a private network, so they can route traffic
to private hypervisors.

#### Private hypervisors

Private hypervisors is where the control plane components will
run. These private hypervisors should have an interface inside a
private network, only reachable by public hypervisors.

## Cluster

A cluster is the abstraction of a whole Kubernetes cluster. It has
all the certificate authorities and shared information required by the
different nodes that are part of the same cluster.

## Node

A node represents a schedulable unit. A node belongs both to a
cluster, and to an hypervisor. A cluster can (and will) have any
number of nodes associated with it.

Two types of nodes exist right now:

* Control Plane
  * Control Plane nodes are formed by:
    * An `etcd` instance
    * An `API server` instance
    * A `Controller Manager` instance
    * A `Scheduler` instance

* Control Plane Ingress
  * Control Plane ingress nodes are formed by:
    * An `haproxy` instance, pointing to all control plane instances
      that are part of the same cluster.
    * (Soon) by a VPN server.

In the future, it will be possible to scale Control Plane components
in a fine grained way; you will be able to create more `API server`
instances without having to create more `Controller Manager` or
`Scheduler` instances.

# Non-goals

* Create or manage infrastructure in any way
* Upgrade clusters
* Perform any kind of worker node management
  * The only exception is the joining process for setting up the VPN
    client
