# Joining worker nodes

`oneinfra` allows you to easily join worker nodes to any managed
cluster.

Requirements on the worker node to be joined:

* `systemd`
* `CRI` endpoint available
  * e.g. `containerd` or `cri-o` already running, or any other service
    that implements the `CRI` interface.

Running the command `oi node join` allows you to join this node to a
cluster, these are the arguments you will need:

| Argument                       | Description                                                                                                            |
|--------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `--nodename`                   | The name that this node will use to join the Kubernetes cluster                                                        |
| `--container-runtime-endpoint` | The CRI endpoint for the container runtime service in this worker node (e.g. `unix:///run/containerd/containerd.sock`) |
| `--image-service-endpoint`     | The CRI endpoint for the image service in this worker node (usually the same as `--container-runtime-endpoint`)        |
| `--apiserver-endpoint`         | The API server endpoint of the cluster that this node will join                                                        |
| `--apiserver-ca-cert-file`     | The file containing the CA certificate to validate the API server endpoint connection                                  |
| `--join-token`                 | The join token that will be used for the joining process                                                               |

When you execute the `oi node join` command, this is the sequence of
actions that will take place:

* `oi` creates a Kubernetes client pointing to the
  `--apiserver-endpoint`, validating its identity with
  `--apiserver-ca-cert-file`. This client will authenticate against
  the API server using the provided join token, which has a very
  locked down set of permissions.

* `oi` downloads the join public key from the cluster present in the
  `oneinfra-join` ConfigMap, inside the `oneinfra-system` namespace.

* `oi` generates a symmetric key, ciphering it with the join public
  key.

* `oi` creates a [`NodeJoinRequest`
  resource](../apis/node/v1alpha1/nodejoinrequest_types.go) with the
  nodename, the ciphered symmetric key, the container runtime endpoint
  and the image service endpoint.

* `oi` performs an active wait on the created `NodeJoinRequest`
  resource, waiting for the `Issued` `condition` to become `True` in
  its `status` field.

* During this time, `oneinfra` on the management cluster will perform
  the following actions:

  * Discover the `NodeJoinRequest` resource in the target cluster.

  * Decipher the symmetric key with the join private key, known to
    `oneinfra` in the management cluster.

  * Fill the following information on the `NodeJoinRequest` `status`
    object, ciphering all fields with the `NodeJoinRequest` provided
    symmetric key in the spec:

    * Kubernetes version
    * VPN address and peers
    * Kubelet kubeconfig contents
    * Kubelet config contents
    * Kubelet server certificate and private key

  * Set the `Issued` condition to `True` on the `NodeJoinRequest`
    `status` field.

* `oi` on the worker node detects the `Issued` condition became
  `True`, and so it performs the following actions:

  * Decipher the kubeconfig file, and write it to disk.
  * Decipher the kubelet config file, and write it to disk.
  * Install the `kubelet` binary that matches the Kubernetes version
    in the `NodeJoinRequest` `status` object.
  * Set up a `systemd` service to enable and start the `kubelet`.
  * (TODO) Set up wireguard, if needed.
  * Exit successfully

From this point on, the `kubelet` in the worker node will
automatically register against the API server, resulting in a fully
functional worker node.
