let oneinfraVersion = (./dhall/versions.dhall).oneinfraVersion

let KubernetesBundle = ./dhall/KubernetesBundle.dhall

let versionTable =
      List/fold
        KubernetesBundle
        (./dhall/versions.dhall).kubernetesVersions
        Text
        ( \(kubernetesBundle : KubernetesBundle) ->
          \(text : Text) ->
            ''
            | ${kubernetesBundle.kubernetesVersion} |
            ${text}''
        )
        ""

in  ''
    | Go Report                                                                                                                                      | License                                                                                                                              |
    |------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
    | [![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0) |

    # oneinfra

    `oneinfra` is a Kubernetes as a Service platform. It empowers you to
    provide or consume Kubernetes clusters at scale, on any platform or
    service provider. You decide.

    |                                                                                                                                          |                                                                                                                                                   |
    |------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
    | [![Cluster list](screenshots/cluster-list.png)](https://raw.githubusercontent.com/oneinfra/oneinfra/main/screenshots/cluster-list.png) | [![Cluster details](screenshots/cluster-details.png)](https://raw.githubusercontent.com/oneinfra/oneinfra/main/screenshots/cluster-details.png) |

    [Read more about its design here](docs/DESIGN.md).


    ## Managed Kubernetes versions

    | Kubernetes version|
    |-------------------|
    ${versionTable}


    ## Lightning-quick start

    * Requirements
      * Docker
      * `kind`
      * `kubectl`

    On a Linux environment, execute:

    ```console
    $ curl https://raw.githubusercontent.com/oneinfra/oneinfra/${oneinfraVersion}/scripts/demo.sh | sh
    ```

    After the script is done, you will be able to access your `oneinfra`
    demo environment in `http://localhost:8000` and log in with username
    `sample-user` with password `sample-user`.


    ## Quick start

    If you prefer to run the quick start yourself instead of the lightning
    quick start, [follow the instructions here](docs/quick-start.md).


    ## Joining worker nodes to a managed cluster

    You can read more details about the [worker joining process
    here](docs/joining-worker-nodes.md).


    ## License

    `oneinfra` is licensed under the terms of the Apache 2.0 license.

    ```
    Copyright (C) 2021 Rafael Fernández López <ereslibre@ereslibre.es>

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
    ```
    ''
