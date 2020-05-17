# Quick start

## Creating managed clusters

1. Create a [managed cluster](../config/samples/simple-cluster.yaml):

    ```console
    $ kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/20.05.0-alpha13/config/samples/simple-cluster.yaml
    $ kubectl wait --for=condition=ReconcileSucceeded --timeout=2m cluster simple-cluster
    ```

2. And access it:

    ```console
    $ kubectl get cluster simple-cluster -o yaml | oi cluster admin-kubeconfig > simple-cluster-kubeconfig.conf
    $ kubectl --kubeconfig=simple-cluster-kubeconfig.conf cluster-info
    Kubernetes master is running at https://172.17.0.4:30000
    CoreDNS is running at https://172.17.0.4:30000/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
    ```

3. (optional) You can then create a [second managed
   cluster](../config/samples/ha-cluster.yaml), this one comprised by
   three control plane instances:

    ```console
    $ kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/20.05.0-alpha13/config/samples/ha-cluster.yaml
    $ kubectl wait --for=condition=ReconcileSucceeded --timeout=2m cluster ha-cluster
    ```

    1. And access it:

        ```console
        $ kubectl get cluster ha-cluster -o yaml | oi cluster admin-kubeconfig > ha-cluster-kubeconfig.conf
        $ kubectl --kubeconfig=ha-cluster-kubeconfig.conf cluster-info
        Kubernetes master is running at https://172.17.0.4:30001
        CoreDNS is running at https://172.17.0.4:30001/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy
        ```
4. List clusters and components on the management cluster:

    ```console
    $ kubectl get clusters -A
    NAMESPACE   NAME             KUBERNETES VERSION   API SERVER ENDPOINT        VPN     VPN CIDR   AGE
    default     ha-cluster       1.18.2               https://172.17.0.4:30001   false              3m10s
    default     simple-cluster   1.18.2               https://172.17.0.4:30000   false              6m40s
    ```

    ```console
    $ kubectl get components -A
    NAMESPACE   NAME                                         CLUSTER          ROLE                    HYPERVISOR                  AGE
    default     ha-cluster-control-plane-4v5ft               ha-cluster       control-plane           test-private-hypervisor-0   3m32s
    default     ha-cluster-control-plane-9d9hq               ha-cluster       control-plane           test-private-hypervisor-0   3m32s
    default     ha-cluster-control-plane-ingress-vffm9       ha-cluster       control-plane-ingress   test-public-hypervisor-0    3m32s
    default     ha-cluster-control-plane-md6dv               ha-cluster       control-plane           test-private-hypervisor-0   3m32s
    default     simple-cluster-control-plane-ingress-28wwd   simple-cluster   control-plane-ingress   test-public-hypervisor-0    7m1s
    default     simple-cluster-control-plane-jqwtz           simple-cluster   control-plane           test-private-hypervisor-0   7m1s
    ```

Then play as much as you want by creating new clusters, deleting
existing ones, or anything you want to try. Have fun!


## Defining clusters

You can have a more detailed [read at the documentation on how to
define clusters](clusters.md).
