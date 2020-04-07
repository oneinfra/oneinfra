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
  kubernetesVersion: 1.18.0
```

Creating a cluster resource by itself won't do anything. You have to
define
[components](https://github.com/oneinfra/oneinfra/blob/master/docs/components.md)
attached to a given cluster, so `oneinfra` will reconcile these
components taking into account the cluster they belong to.

You can read the
[DESIGN.md](https://github.com/oneinfra/oneinfra/blob/master/docs/DESIGN.md)
document for a broad overview.
