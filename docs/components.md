# Components

* [Documentation](https://pkg.go.dev/github.com/oneinfra/oneinfra/apis/cluster/v1alpha1?tab=doc#Component)

Components are the schedulable unit that allows `oneinfra` to
reconcile the different parts of the system that conform a Kubernetes
control plane.

Components are automatically reconciled by `oneinfra` based [on the
cluster](clusters.md) number desired of control plane replicas.

You can read the [DESIGN.md](DESIGN.md) document for a broad overview.
