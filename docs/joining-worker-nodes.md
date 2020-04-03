# Joining worker nodes

`oneinfra` allows you to easily join worker nodes to any managed
cluster.

Requirements on the worker node to be joined:

* `systemd`
* `CRI` endpoint available

Running the command `oi node join` allows you to join this node to a
cluster, these are the arguments you will need:

| Argument                     | Description                                                                               |
|------------------------------|-------------------------------------------------------------------------------------------|
| --nodename                   | The name that this node will use to join the Kubernetes cluster                           |
| --container-runtime-endpoint | The CRI endpoint for the container runtime service (containerd, cri-o...)                 |
| --image-service-endpoint     | The CRI endpoint for the image service (usually the same as --container-runtime-endpoint) |
| --apiserver-endpoint         | The API server endpoint of the cluster that this node will join                           |
| --apiserver-ca-cert-file     | The file containing the CA certificate to validate the API server endpoint                |
| --join-token                 | The join token that will be used for joining the existing cluster                         |
| --join-token-public-key-file | The public key of the cluster to be joining                                               |

When you execute this command, this is what will happen:

* Create a client pointing to the `--apiserver-endpoint`, validating
  its identity with `--apiserver-ca-cert-file`.

* Generate a symmetric key, and will cypher it with the
  `--join-token-public-key-file` argument.

* Create a [node join request
  resource](https://github.com/oneinfra/oneinfra/blob/master/apis/node/v1alpha1/nodejoinrequest_types.go)
  with the `--nodename`, the ciphered symmetric key, the
  `--container-runtime-endpoint`, and the
  `--image-service-endpoint`.

* Wait for `oneinfra` to reconcile this `NodeJoinRequest`, watching
  the created `NodeJoinRequest` resource for the `issued` `condition`
  in its `status` field.

* During this time, `oneinfra` will perform the following actions:

  * Discover the `NodeJoinRequest` in the target cluster.

  * Decypher the symmetric key with the private key matching the
    `--join-token-public-key-file`.

  * Fill the following information on the `NodeJoinRequest` status
    object, ciphered with the deciphered symmetric key:

    * Kubernetes version
    * VPN address and peers
    * Kubelet kubeconfig contents
    * Kubelet config contents
    * Kubelet server certificate

  * Set the `issued` condition on the `NodeJoinRequest`.

* `oi` on the worker node detects the `issued` condition, so it will
  perform the following actions:

  * Decipher the kubeconfig file, and write it to disk.
  * Decipher the kubelet config file, and write it to disk.
  * Install the `kubelet` binary, by copying it from a
    `kubelet-installer` container image, using the provided Kubernetes
    version tag.
  * Set up a `systemd` service to enable and start the `kubelet`.
  * (TODO) Set up wireguard, if needed.

From this point on, the `kubelet` will automatically register against
the API server, resulting in a fully functional worker node.
