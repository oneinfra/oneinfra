# Release

This document describes what artifacts conform a `oneinfra` release
and how to create one.


## Considerations

A `oneinfra` release might include only changes on `oneinfra` itself,
or on some other components as well (e.g. the supported Kubernetes
versions).

In general, there are a number of artifacts that conform a `oneinfra`
release as a whole, let's go through them.


### The `RELEASE` file

The [`RELEASE` file](../RELEASE) contains all versioning information
for a given `oneinfra` release.

* `consoleVersion`: the recommended [`console`
  project](https://github.com/oneinfra/console) tag that should be
  used to manage `oneinfra`. The `console` is an optional project that
  enables a Web UI view to manage `oneinfra` managed clusters.

* `defaultKubernetesVersion`: the default Kubernetes version a managed
  cluster should be defaulted to if it doesn't include an explicit
  Kubernetes version.

* `kubernetesVersions`: the list of all suppoted Kubernetes versions
  for this `oneinfra` release. This list contains a specific `etcd`
  and `CoreDNS` version for that specific Kubernetes version.

The `RELEASE` file is used to construct all release artifacts when
tagging a new release.


#### Artifacts constructed from the `RELEASE` file

These artifacts are core for `oneinfra` and require to be built for
every release.

For the current `oneinfra` tag:

* `oi-local-hypervisor-set` binary in the GitHub release itself.
* `oi` binary in the GitHub release itself.
* `oneinfra/oi-manager:<oneinfra-tag>` container image that contains
  the `oneinfra` Kubernetes manager
* `oneinfra/oi:<oneinfra-tag>` container image that contains the
  `oneinfra` CLI tool

For each supported Kubernetes version tag:

* `oneinfra/kubelet-installer:<kubernetes-tag>` container image that
  contains the Kubelet installer to be pulled on worker nodes when
  they are joined to a managed cluster


### The `RELEASE_TEST` file

The [`RELEASE_TEST` file](../RELEASE_TEST) contains all versioning
information for ease of testing.

* `containerdVersions`: the list of all `containerd` versions used for
  testing `oneinfra`. This list contains a specific
  [`cri-tools`](https://github.com/kubernetes-sigs/cri-tools) and
  [`CNI plugins`](https://github.com/containernetworking/plugins)
  version.

* `kubernetesVersions`: the list of all Kubernetes versions supported
  in a Kubernetes release. This list contains a specific `containerd`
  version and `pause` image version.


#### Artifacts constructed from the `RELEASE_TEST` file

These artifacts are published for faster CI runs. They are not
required to exist in any case, but they improve the CI times if they
do.

For each `containerd` version:

* `oneinfra/containerd:<containerd-tag>`: container image that
  contains a `containerd` instance

For each Kubernetes version:

* `oneinfra/hypervisor:<kubernetes-tag>`: container image that is
  based on `oneinfra/containerd:<containerd-tag>` that contains all
  test dependencies for a given Kubernetes version already pulled
  (`haproxy:latest`, `tooling:latest`, `etcd:<etcd-version>`,
  `pause:<pause-version>`, `kube-apiserver:<kubernetes-version>`,
  `kube-controller-manager:<kubernetes-version>`,
  `kube-scheduler:<kubernetes-version>`...)


### Artifacts not managed by the release

Some artifacts are manually constructed and pushed, and are core for
`oneinfra` to operate correctly. They are not (yet) managed by the
release pipeline or process.

* `oneinfra/console`: this is the [`console`
  project](https://github.com/oneinfra/console) that is tagged
  independently to `oneinfra`.

* `oneinfra/builder`: the builder image used to build `oneinfra`
  binaries everywhere

* `oneinfra/tooling`: a tooling image used to perform changes on the
  host (like copying files, or calling to the D-Bus system bus on the
  host)

* `oneinfra/etcd`: [contains some changes to `etcd` so learner members
  can be properly
  promoted](https://github.com/etcd-io/etcd/pull/11640). It is
  expected to be dropped when an upstream version can be used that
  contains this changeset.

* `oneinfra/dqlite`, `oneinfra/kine`: used to experiment with `kine` +
  `dqlite`. Not being used in production nor in testing right now, but
  could be used in the future, and thus, included in the release
  pipeline.

* `oneinfra/haproxy`: used to create the HAProxy instances in the
  ingress hypervisors, that will serve as the managed control planes
  endpoint for external users.


## Creating a release

`oneinfra` follows the [calver versioning
scheme](https://calver.org/).

The release process is very automated. Let's assume that you want to
create a new release tagged `YY.MM.MINOR-META`. The `META` section
could be ommitted if it's not an `alpha`, `beta` or `rc` release.

* Make sure `RELEASE` and `RELEASE_TEST` matches and that they have
  the indended versions of all required components before releasing.
* Create a new branch for the release preparation
  * `git co -b prepare-YY.MM.MINOR-META`
* Ensure that all generated code is up to date
  * `ONEINFRA_VERSION=YY.MM.MINOR-META make generate-all`
* Commit and push all your changes
  * `git commit -am "Prepare for releasing YY.MM.MINOR-META"`
  * `git push origin prepare-YY.MM.MINOR-META`
* Wait until CI reports everything is green
* Sign tag the new version
  * `git tag -s -m YY.MM.MINOR-META YY.MM.MINOR-META`
* Push the new tag
  * `git push origin YY.MM.MINOR-META`
* Wait until CI reports everything is green
  * A new release should have been created on GitHub automatically
  * New images should have been published to the Docker Hub
  * Assets should have been published in the release in GitHub
* Merge changes into `master` and push them
  * `git checkout master`
  * `git merge prepare-YY.MM.MINOR-META`
  * `git push origin master`
* Remove the preparation branch locally and remotely
  * `git branch -d prepare-YY.MM.MINOR-META`
  * `git push origin :prepare-YY.MM.MINOR-META`

Release is completely finished now!
