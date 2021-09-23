let oneinfraVersion = (../dhall/versions.dhall).oneinfraVersion

in  ''
    # Install from source

    If you prefer to build `oneinfra` from source, you can either clone
    this repository and run `make`, or use `go get` directly.

    ```console
    $ GO111MODULE=on go get github.com/oneinfra/oneinfra/...@${oneinfraVersion}
    ```
    ''
