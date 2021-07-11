let ContainerdBundleVersion : Type =
    { containerdVersion: Text
    , criToolsVersion: Text
    , cniPluginsVersion: Text
    }

let TestBundle : Type =
    { containerdBundleVersion: ContainerdBundleVersion
    , pauseVersion: Text
    }

let containerd134TestBundle: TestBundle = {
    containerdBundleVersion =
    {
    , containerdVersion = "1.3.4"
    , criToolsVersion = "1.18.0"
    , cniPluginsVersion = "0.8.6"
    },
    pauseVersion = "3.1"
}

let KubernetesBundle : Type =
    { kubernetesVersion: Text
    , etcdVersion: Text
    , coreDNSVersion: Text
    , testBundle : TestBundle }

let defaultKubernetesVersion: KubernetesBundle =
    { kubernetesVersion = "1.21.0"
    , etcdVersion = "3.4.3"
    , coreDNSVersion = "1.6.7"
    , testBundle = containerd134TestBundle
    }

let kubernetesVersions : List KubernetesBundle = [
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