# Quick start

## Without Kubernetes as a management cluster (for testing purposes only)

* Requirements
  * Docker

If you don't want to deploy Kubernetes to test `oneinfra`, you can use
the `oi` CLI tool that will allow you to test the reconciliation
processes of `oneinfra` without the need of a Kubernetes cluster.

```console
$ oi-local-hypervisor-set create | oi cluster inject | oi reconcile > cluster-manifests.conf
```

And access it:

```console
$ cat cluster-manifests.conf | oi cluster admin-kubeconfig > cluster-kubeconfig.conf
$ kubectl --kubeconfig=cluster-kubeconfig.conf cluster-info
Kubernetes master is running at https://172.17.0.3:30000
CoreDNS is running at https://172.17.0.3:30000/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
```

In this mode it's very important to understand that `oi` will read
manifests from `stdin` and output them into `stdout`, make sure you
keep a file up to date with the latest reconciled resources -- this is
why this model is not suitable for production.
