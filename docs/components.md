# Components

* [Documentation](https://pkg.go.dev/github.com/oneinfra/oneinfra/apis/cluster/v1alpha1?tab=doc#Component)

Depending on how many components you define, you will be creating a
single instance control plane, or an HA control plane. For example, if
you define a `control-plane` role component along with a
`control-plane-ingress` component:

```yaml
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Component
metadata:
  name: my-control-plane-1
spec:
  cluster: my-cluster
  role: control-plane
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Component
metadata:
  name: my-control-plane-ingress
spec:
  cluster: my-cluster
  role: control-plane-ingress
```

Then `oneinfra` will start reconciliating the control plane instance
and the control plane ingress. This will be a single control plane
instance cluster.

If you create two more resources:

```yaml
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Component
metadata:
  name: my-control-plane-2
spec:
  cluster: my-cluster
  role: control-plane
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Component
metadata:
  name: my-control-plane-3
spec:
  cluster: my-cluster
  role: control-plane
---
```

Then, `oneinfra` will reconcile a 3 control plane instance HA cluster,
with a single control plane ingress.

---

**Note**: at the moment **only one ingress component** is allowed per
cluster, what means that the system has a single point of
failure. This will be ammended soon.

---

You can read the [DESIGN.md](DESIGN.md) document for a broad overview.
