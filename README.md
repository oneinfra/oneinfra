| Go Report                                                                                                                                      | Travis                                                                                                             | CircleCI                                                                                                             | Azure Test                                                                                                                                                                             | Azure Release                                                                                                                                                                                | License                                                                                                                              |
|------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| [![Go Report Card](https://goreportcard.com/badge/github.com/oneinfra/oneinfra)](https://goreportcard.com/report/github.com/oneinfra/oneinfra) | [![Travis CI](https://travis-ci.org/oneinfra/oneinfra.svg?branch=main)](https://travis-ci.org/oneinfra/oneinfra) | [![CircleCI](https://circleci.com/gh/oneinfra/oneinfra.svg?style=shield)](https://circleci.com/gh/oneinfra/oneinfra) | [![Test Pipeline](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=3&_a=summary) | [![Release Pipeline](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/release?branchName=main)](https://dev.azure.com/oneinfra/oneinfra/_build/latest?definitionId=4&_a=summary) | [![License: Apache 2.0](https://img.shields.io/badge/License-Apache2.0-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0) |


# oneinfra

`oneinfra` is a Kubernetes as a Service platform. It empowers you to
provide or consume Kubernetes clusters at scale, on any platform or
service provider. You decide.

|                                                                                                                                          |                                                                                                                                                   |
|------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| [![Cluster list](screenshots/cluster-list.png)](https://raw.githubusercontent.com/oneinfra/oneinfra/main/screenshots/cluster-list.png) | [![Cluster details](screenshots/cluster-details.png)](https://raw.githubusercontent.com/oneinfra/oneinfra/main/screenshots/cluster-details.png) |

[Read more about its design here](docs/DESIGN.md).


## Managed Kubernetes versions

| Kubernetes version | Deployable with      | Default in           |                                                                                                                                                                                                     |
|--------------------|----------------------|----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `1.15.12`          | `20.09.0-alpha21` |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.15.12)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3) |
| `1.16.15`          | `20.09.0-alpha21` |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.16.15)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3) |
| `1.17.17`          | `20.09.0-alpha21` |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.17.17)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3) |
| `1.18.18`          | `20.09.0-alpha21` |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.18.18)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3) |
| `1.19.10`          | `20.09.0-alpha21` | `20.09.0-alpha21` | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.19.10)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3) |
| `1.20.6`           | Next release         |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.20.6)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3)  |
| `1.21.0`           | Next release         |                      | [![Build Status](https://dev.azure.com/oneinfra/oneinfra/_apis/build/status/test?branchName=main&jobName=e2e%20tests%20-%201.21.0)](https://dev.azure.com/oneinfra/oneinfra/_build?definitionId=3)  |


## Lightning-quick start

* Requirements
  * Docker
  * `kind`
  * `kubectl`

On a Linux environment, execute:

```console
$ curl https://raw.githubusercontent.com/oneinfra/oneinfra/20.09.0-alpha21/scripts/demo.sh | sh
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
