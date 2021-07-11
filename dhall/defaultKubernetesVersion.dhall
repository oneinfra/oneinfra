  { kubernetesVersion = "1.21.0"
  , etcdVersion = "3.4.3"
  , coreDNSVersion = "1.6.7"
  , testBundle = ./containerd134TestBundle.dhall
  }
: ./KubernetesBundle.dhall
