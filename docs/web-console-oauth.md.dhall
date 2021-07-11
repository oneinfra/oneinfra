let consoleVersion = (../dhall/versions.dhall).consoleVersion

in  ''
    # Web console: OAuth backends

    This is the list of the OAuth backends supported by the web console.

    ## Github

    Deploy the console:

    ```console
    $ kubectl apply -f https://raw.githubusercontent.com/oneinfra/console/${consoleVersion}/config/generated/all-github-oauth.yaml
    ```

    Create a new OAuth app in the Github settings page. Make sure that the
    authorization callback URL points to
    `https://<host>:<port>/api/auth/github`.

    Populate your Github OAuth client ID and token in a `github-oauth`
    secret inside the `oneinfra-system` namespace:

    ```console
    $ kubectl create secret generic -n oneinfra-system github-oauth --from-literal=client-id=<Github OAuth Client ID> --from-literal=client-secret=<Github OAuth Secret>
    ```
    ''
