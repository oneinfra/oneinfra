# Hypervisors

Hypervisors is where `oneinfra` will run the control plane components
for all managed clusters.

* [Documentation](https://pkg.go.dev/github.com/oneinfra/oneinfra/apis/infra/v1alpha1?tab=doc#Hypervisor)

An Hypervisor has different attributes that you can specify:

* A way for `oneinfra` to run processes on hypervisors
  * `LocalCRIEndpoint`: A CRI socket is exposed in the local
    filesystem.
    * When using a local CRI endpoint, requests are unauthenticated
      and use the UNIX socket directly.
  * `RemoteCRIEndpoint`: A CRI socket is exposed in the network,
    reachable by `oneinfra`.
    * When using a remote CRI endpoint, requests are authenticated
      with a client certificate and private key. They also require a
      CA certificate to validate the server presented certificate in
      the client side.

* Public and private hypervisors
  * `Public` hypervisors is where the control plane ingress components
    will be placed (e.g. `haproxy` and a VPN terminator). Managed
    cluster users will connect to public hypervisors.

* `IPAddress` is the IP address of the hypervisor
  * For `Public` hypervisors, this must be a reachable IP address for
    managed clusters users.
  * For `Private` hypervisors, this must be a reachable IP address for
    other `Private` hypervisors, and `Public` hypervisors, so
    components scheduled on different `Private` hypervisors can talk
    to each other, and `Public` hypervisors can route traffic to
    `Private` ones.

* Port ranges
  * Every service, either public or private is going to be allocated a
    port on the scheduled hypervisor, so `oneinfra` needs to know what
    port range is safe to use.


## Setting up an hypervisor

For setting an hypervisor up, you will need a service that implements
the Container Runtime Interface set up (e.g. containerd, cri-o...).

The hypervisor spec looks as follows:

```go
type HypervisorSpec struct {
	// LocalCRIEndpoint is the unix socket where this hypervisor is
	// reachable. This is only intended for development and testing
	// purposes. On production environments RemoteCRIEndpoint should be
	// used. Either a LocalCRIEndpoint or a RemoteCRIEndpoint has to be
	// provided.
	//
	// +optional
	LocalCRIEndpoint *LocalHypervisorCRIEndpoint `json:"localCRIEndpoint,omitempty"`

	// RemoteCRIEndpoint is the TCP address where this hypervisor is
	// reachable. Either a LocalCRIEndpoint or a RemoteCRIEndpoint has
	// to be provided.
	//
	// +optional
	RemoteCRIEndpoint *RemoteHypervisorCRIEndpoint `json:"remoteCRIEndpoint,omitempty"`

	// Public hypervisors will be scheduled cluster ingress components,
	// whereas private hypervisors will be scheduled the control plane
	// components themselves.
	Public bool `json:"public"`

	// IPAddress of this hypervisor. Public hypervisors must have a
	// publicly reachable IP address.
	IPAddress string `json:"ipAddress,omitempty"`

	// PortRange is the port range to be used for allocating exposed
	// components.
	PortRange HypervisorPortRange `json:"portRange,omitempty"`
}
```

Ideally, for production targeted hypervisors, you will use the
`RemoteCRIEndpoint`, whose specification looks like the following:

```go
type RemoteHypervisorCRIEndpoint struct {
	// CRIEndpoint is the address where this CRI endpoint is listening
	CRIEndpoint string `json:"criEndpointURI,omitempty"`

	// CACertificate is the CA certificate to validate the connection
	// against
	CACertificate string `json:"caCertificate,omitempty"`

	// ClientCertificate is the client certificate that will be used to
	// authenticate requests
	ClientCertificate *commonv1alpha1.Certificate `json:"clientCertificate,omitempty"`
}
```

And so, when `oneinfra` connects to this hypervisor using a remote CRI
endpoint, it will validate the server presented certificate with the
`CACertificate` and will present the server the client certificate and
key provided in `ClientCertificate`.

---

**Note**: in the near future `oneinfra` will allow you to set up
hypervisors in an easier way.

---

You will need to set up an authentication proxy on the hypervisor. The
test environment uses `haproxy`, so it is listening in a TCP port,
performing client certificate authentication, and forwarding those
requests to a local UNIX socket (where `containerd`, or `cri-o` are
listening).

An example of an `haproxy` configuration is as follows:

```
global
  chroot /var/lib/haproxy
  daemon
defaults
  log global
  mode tcp
  timeout connect 10s
  timeout client  60s
  timeout server  60s
frontend cri_frontend
  bind *:<HAProxy port> ssl crt <HAProxy cert bundle path> ca-file <HAProxy CA client cert path> verify required
  default_backend cri_backend
backend cri_backend
  server cri unix@containerd.sock
```

In this example, the `containerd.sock` is placed inside the chrooted
environment `/var/lib/haproxy`.

You can inspect the testing environment `oneinfra` creates in this
setup by executing `oi-local-hypervisor-set create --tcp`.

Also, since `oneinfra` will authenticate against the hypervisor using
a client certificate, you will need several certificates:

* HAProxy cert bundle path: refers to the bundle of the endpoint
  certificate and private key, this is the server certificate of this
  CRI endpoint. It will be used by `oneinfra` client to validate the
  server connection. The CA used to create this certificate will be
  placed in the `RemoteHypervisorCRIEndpoint` `CACertificate` field,
  PEM encoded.

* HAProxy CA client cert path: is the CA used by `haproxy` to validate
  the `oneinfra` client certificate, authenticating requests. The
  client certificate used by `oneinfra` to authenticate against the
  CRI endpoint will be placed in the `RemoteHypervisorCRIEndpoint`
  `ClientCertificate` field, which consists in a `Certificate` and
  `PrivateKey`, both PEM encoded.

You can read the [DESIGN.md](DESIGN.md) document for a broad
overview.
