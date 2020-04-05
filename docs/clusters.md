# Clusters

Clusters are the main component that will be used to define new
clusters in your infrastructure.

A minimal example is:

```yaml
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
```

There are many things you can configure on a cluster, but by default
if any of the fields are empty, `oneinfra` will default them to sane
defaults. For completeness, this is everything you can configure at
the moment:

```go
// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	// +optional
	CertificateAuthorities *CertificateAuthorities `json:"certificateAuthorities,omitempty"`
	// +optional
	EtcdServer *EtcdServer `json:"etcdServer,omitempty"`
	// +optional
	APIServer *KubeAPIServer `json:"apiServer,omitempty"`
	// +optional
	VPNCIDR string `json:"vpnCIDR,omitempty"`
	// +optional
	JoinKey *commonv1alpha1.KeyPair `json:"joinKey,omitempty"`
	// +optional
	JoinTokens []string `json:"joinTokens,omitempty"`
}
```

Creating a cluster by itself won't do anything. You need to define
components attached to this cluster.


# Components

Depending on how many components you define, you will be creating a
single control plane instance, or an HA control plane. For example, if
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

If you add two more resources:

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

Then, `oneinfra` will reconcile a 3 control plane instances HA
cluster, with a single control plane ingress.

---

**Note**: at the moment **only one ingress component** is allowed per
cluster, what means that the system has a single point of
failure. This will be ammended soon.

---
