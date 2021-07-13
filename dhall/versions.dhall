let containerd134TestBundle = ./containerd134TestBundle.dhall
let defaultKubernetesVersion = ./defaultKubernetesVersion.dhall

let kubernetesVersions : List ./kubernetesBundle.dhall = [
    { kubernetesVersion = "1.15.12"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.3.1"
    , testBundle = containerd134TestBundle
    },
    { kubernetesVersion = "1.16.15"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.2"
    , testBundle = containerd134TestBundle
    },
    { kubernetesVersion = "1.17.17"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.7"
    , testBundle = containerd134TestBundle
    },
    { kubernetesVersion = "1.18.18"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.7"
    , testBundle = containerd134TestBundle
    },
    { kubernetesVersion = "1.19.10"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.7"
    , testBundle = containerd134TestBundle
    },
    { kubernetesVersion = "1.20.6"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.7"
    , testBundle = containerd134TestBundle
    },
    defaultKubernetesVersion
] in {
    kubernetesVersions,
    defaultKubernetesVersion
}