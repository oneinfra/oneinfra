let kubernetesVersions = ./kubernetesVersions.dhall

let containerdVersions = ./containerdVersions.dhall

let defaultKubernetesVersion = ./defaultKubernetesVersion.dhall

let oneinfraVersion = "20.09.0-alpha21"

let consoleVersion = "20.05.0-alpha4"

in  { kubernetesVersions
    , containerdVersions
    , defaultKubernetesVersion
    , oneinfraVersion
    , consoleVersion
    }
