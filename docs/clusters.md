# Clusters

Clusters are the main resource kind that will be used to glue all
components that conform a cluster. The cluster resource contains the
authoritative information that every component that belongs to this
cluster will use (e.g. certificate authorities, certificates, join
tokens...).

* [Documentation](https://pkg.go.dev/github.com/oneinfra/oneinfra/apis/cluster/v1alpha1?tab=doc#Cluster)

A minimal example of a cluster resource is:

```yaml
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Cluster
metadata:
  name: my-cluster
```

There are many things you can configure on a cluster, but by default
if any of the fields are empty, `oneinfra` will default them to sane
defaults.

The specification of a cluster looks as follows:

```go
type ClusterSpec struct {
	// +optional
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`

	// The number of control plane replicas this cluster will
	// manage. One control plane replica if not provided.
	//
	// +optional
	ControlPlaneReplicas int `json:"controlPlaneReplicas,omitempty"`

	// +optional
	CertificateAuthorities *CertificateAuthorities `json:"certificateAuthorities,omitempty"`

	// +optional
	EtcdServer *EtcdServer `json:"etcdServer,omitempty"`

	// +optional
	APIServer *KubeAPIServer `json:"apiServer,omitempty"`

	// +optional
	VPN *VPN `json:"vpn,omitempty"`

	// +optional
	JoinKey *commonv1alpha1.KeyPair `json:"joinKey,omitempty"`

	// +optional
	JoinTokens []string `json:"joinTokens,omitempty"`

	// +optional
	Networking *ClusterNetworking `json:"networking,omitempty"`
}
```

So, for example, we can create two different clusters with two
different supported managed versions, by defining two different
resources:

```yaml
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Cluster
metadata:
  name: my-1-17-cluster
spec:
  kubernetesVersion: 1.17.4
---
apiVersion: cluster.oneinfra.ereslibre.es/v1alpha1
kind: Cluster
metadata:
  name: my-1-18-cluster
spec:
  kubernetesVersion: 1.18.1
```

When the `Cluster` resource is created, `oneinfra` will automatically
create the desired number of control plane replicas -- [as component
resources](components.md), and a single control plane ingress
instance.

---

**Note**: at the moment **only one ingress component** is allowed per
cluster, what means that the system has a single point of
failure. This will be ammended soon.

---

You can read the [DESIGN.md](DESIGN.md) document for a broad overview.
