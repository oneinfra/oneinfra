let oneinfraVersion = (../dhall/versions.dhall).oneinfraVersion

let consoleVersion = (../dhall/versions.dhall).consoleVersion

in  ''
    # Quick start

    The quick start will drive you manually through the same process that
    the lightning-quick start does in an automated way.

    For the quick start we will leverage Kubernetes as a management
    cluster, but [you can also try `oneinfra` without the need of having
    Kubernetes as a management cluster if you
    prefer.](quick-start-without-kubernetes.md)


    ### With Kubernetes as a management cluster

    * Requirements
      * A Kubernetes cluster that will be the management cluster, where we
        will deploy the `oneinfra` controller manager
      * The `oneinfra` controller manager running in the management
        cluster needs to be able to reach the hypervisors you define
      * Docker, if you want to create fake local hypervisors using
        `oi-local-hypervisor-set`, or if you are going to use `kind`

    1. Install kind and create the management cluster. If you already have
       a Kubernetes cluster you can use, you can skip this step.

        ```console

        $ kind create cluster
        ```

    2. Deploy `cert-manager` and `oneinfra`.

        ```console
        $ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.14.1/cert-manager.yaml
        $ kubectl wait --for=condition=Available deployment --timeout=2m -n cert-manager --all
        $ kubectl apply -f https://raw.githubusercontent.com/oneinfra/oneinfra/${oneinfraVersion}/config/generated/all.yaml
        $ kubectl wait --for=condition=Available deployment --timeout=2m -n oneinfra-system --all
        ```

    3. Create a local set of fake hypervisors, so `oneinfra` can schedule
    managed control plane components. You can [also provision and define
    your own set of hypervisors](hypervisors.md) if you prefer.

        ```console
        $ oi-local-hypervisor-set create --tcp | kubectl apply -f -
        ```

        Note that `oi-local-hypervisor-set` **should not** be used to
        provision hypervisors for production environments -- this tool is
        just to easily test `oneinfra`. In production environments you
        will have to provision the hypervisors and define them as
        [described here](hypervisors.md).


    4. Create [as many managed clusters as you want using `oneinfra`
       API's](quick-start-creating-managed-clusters.md). You can
       [also deploy the Web console](#deploy-the-web-console-optional)
       that will allow you to create clusters through an intuitive web
       interface.


    ## Deploy the Web console (optional)

    `oneinfra` provides you a [simple web
    console](https://github.com/oneinfra/console).

    ### Generate a JWT key for the Web console

    You will have to create a JWT key that the console backend will use to
    generate your JWT tokens when authenticating users. Let's do that:

    ```console
    $ kubectl create secret generic -n oneinfra-system jwt-key --from-literal=jwt-key=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
    ```


    ### Kubernetes secrets as authentication mechanism

    This is an only for testing authentication mechanism. Secrets inside a
    namespace `oneinfra-users` resemble users.

    ```console
    $ kubectl apply -f https://raw.githubusercontent.com/oneinfra/console/${consoleVersion}/config/generated/all-kubernetes-secrets.yaml
    $ kubectl wait --for=condition=Available deployment --timeout=2m -n oneinfra-system --all
    ```

    A user named `sample-user` with password `sample-user` has been
    automatically created. Refer to the console inline help to learn how
    to manage users with this authentication mechanism.

    If you prefer to enable other authentication mechanisms that are
    production ready, please [read the instructions
    here](web-console-oauth.md).


    ### Access the web console service

    You can use any regular Kubernetes means to expose the web console
    service; for ease of testing you can access it by using a port
    forward:

    ```console
    $ kubectl port-forward -n oneinfra-system svc/oneinfra-console 8000:80
    ```

    You can now access the console by visiting `http://localhost:8000` in
    your browser.
    ''
