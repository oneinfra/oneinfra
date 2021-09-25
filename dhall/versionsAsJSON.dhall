let Prelude = ./prelude.dhall

let versions = ./versions.dhall

let kubernetesVersion/ToJSON
    : ./KubernetesBundle.dhall -> Prelude.JSON.Type
    = \(kubernetesBundle : ./KubernetesBundle.dhall) ->
        Prelude.JSON.object
          ( toMap
              { version = Prelude.JSON.string kubernetesBundle.kubernetesVersion
              , etcdVersion = Prelude.JSON.string kubernetesBundle.etcdVersion
              , coreDNSVersion =
                  Prelude.JSON.string kubernetesBundle.coreDNSVersion
              }
          )

let containerdTestBundle/ToJSON
    : ./ContainerdTestBundle.dhall -> Prelude.JSON.Type
    = \(containerdTestBundle : ./ContainerdTestBundle.dhall) ->
        Prelude.JSON.object
          ( toMap
              { version =
                  Prelude.JSON.string containerdTestBundle.containerdVersion
              , criToolsVersion =
                  Prelude.JSON.string containerdTestBundle.criToolsVersion
              , cniPluginsVersion =
                  Prelude.JSON.string containerdTestBundle.cniPluginsVersion
              }
          )

let kubernetesTestVersion/ToMap
    : ./KubernetesBundle.dhall ->
        { mapKey : Text, mapValue : Prelude.JSON.Type }
    = \(kubernetesBundle : ./KubernetesBundle.dhall) ->
        { mapKey = kubernetesBundle.kubernetesVersion
        , mapValue =
            Prelude.JSON.object
              ( toMap
                  { containerdVersion =
                      Prelude.JSON.string
                        kubernetesBundle.testBundle.containerdBundleVersion.containerdVersion
                  , pauseVersion =
                      Prelude.JSON.string
                        kubernetesBundle.testBundle.pauseVersion
                  }
              )
        }

let containerdVersions/ToJSON
    : List ./ContainerdTestBundle.dhall -> Prelude.JSON.Type
    = \(containerdTestBundles : List ./ContainerdTestBundle.dhall) ->
        Prelude.JSON.array
          ( Prelude.List.map
              ./ContainerdTestBundle.dhall
              Prelude.JSON.Type
              containerdTestBundle/ToJSON
              containerdTestBundles
          )

let kubernetesVersions/ToJSON
    : List ./KubernetesBundle.dhall -> Prelude.JSON.Type
    = \(kubernetesBundles : List ./KubernetesBundle.dhall) ->
        Prelude.JSON.array
          ( Prelude.List.map
              ./KubernetesBundle.dhall
              Prelude.JSON.Type
              kubernetesVersion/ToJSON
              kubernetesBundles
          )

let kubernetesTestVersions/ToJSON
    : List ./KubernetesBundle.dhall -> Prelude.JSON.Type
    = \(kubernetesBundles : List ./KubernetesBundle.dhall) ->
        Prelude.JSON.object
          ( Prelude.List.fold
              ./KubernetesBundle.dhall
              kubernetesBundles
              (List { mapKey : Text, mapValue : Prelude.JSON.Type })
              ( \(x : ./KubernetesBundle.dhall) ->
                \(y : List { mapKey : Text, mapValue : Prelude.JSON.Type }) ->
                  [ kubernetesTestVersion/ToMap x ] # y
              )
              ([] : List { mapKey : Text, mapValue : Prelude.JSON.Type })
          )

let releaseData/ToJSON =
      Prelude.JSON.object
        ( toMap
            { consoleVersion = Prelude.JSON.string versions.consoleVersion
            , defaultKubernetesVersion =
                Prelude.JSON.string
                  versions.defaultKubernetesVersion.kubernetesVersion
            , kubernetesVersions =
                kubernetesVersions/ToJSON versions.kubernetesVersions
            }
        )

let testData/ToJSON =
      Prelude.JSON.object
        ( toMap
            { containerdVersions =
                containerdVersions/ToJSON versions.containerdVersions
            , kubernetesVersions =
                kubernetesTestVersions/ToJSON versions.kubernetesVersions
            }
        )

in      versions
    //  { kubernetesVersions/ToJSON, releaseData/ToJSON, testData/ToJSON }
