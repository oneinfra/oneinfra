let Prelude = ../../dhall/prelude.dhall

let GithubActions =
      https://raw.githubusercontent.com/regadas/github-actions-dhall/master/package.dhall sha256:66b276bb67cca4cfcfd1027da45857cc8d53e75ea98433b15dade1e1e1ec22c8

let KubernetesBundle = ../../dhall/KubernetesBundle.dhall

let kubernetesVersions = (../../dhall/versions.dhall).kubernetesVersions

let setupSteps =
      [ GithubActions.Step::{ uses = Some "actions/checkout@v2" }
      , GithubActions.Step::{
        , run = Some "git fetch --prune --unshallow --tags"
        }
      , GithubActions.Step::{ uses = Some "cachix/install-nix-action@v14" }
      , GithubActions.Step::{
        , uses = Some "cachix/cachix-action@v10"
        , `with` = Some
            ( toMap
                { name = "oneinfra"
                , authToken = "\${{ secrets.CACHIX_AUTH_TOKEN }}"
                }
            )
        }
      ]

let e2eTests =
      Prelude.List.map
        KubernetesBundle
        { mapKey : Text, mapValue : GithubActions.types.Job }
        ( \(bundle : KubernetesBundle) ->
            { mapKey =
                "e2e-${Prelude.Text.replace "." "_" bundle.kubernetesVersion}"
            , mapValue = GithubActions.Job::{
              , name = Some "e2e (${bundle.kubernetesVersion})"
              , runs-on = GithubActions.RunsOn.Type.ubuntu-latest
              , steps =
                    setupSteps
                  # [ GithubActions.Step::{
                      , run = Some
                          "nix-shell --pure --run 'KUBERNETES_VERSION=${bundle.kubernetesVersion} make e2e'"
                      }
                    ]
              }
            }
        )
        kubernetesVersions

in  GithubActions.Workflow::{
    , name = "Build, test and publish"
    , on = GithubActions.On::{ push = Some GithubActions.Push::{=} }
    , jobs =
          toMap
            { build = GithubActions.Job::{
              , name = Some "Build"
              , runs-on = GithubActions.RunsOn.Type.ubuntu-latest
              , steps =
                    setupSteps
                  # [ GithubActions.steps.run
                        { run = "nix-shell --pure --run 'make'" }
                    ]
              }
            , test = GithubActions.Job::{
              , name = Some "Test"
              , runs-on = GithubActions.RunsOn.Type.ubuntu-latest
              , steps =
                    setupSteps
                  # [ GithubActions.steps.run
                        { run = "nix-shell --pure --run 'make test'" }
                    ]
              }
            , e2e-default = GithubActions.Job::{
              , name = Some "e2e (default)"
              , runs-on = GithubActions.RunsOn.Type.ubuntu-latest
              , steps =
                    setupSteps
                  # [ GithubActions.steps.run
                        { run = "nix-shell --pure --run 'make e2e'" }
                    ]
              }
            }
        # e2eTests
    }
