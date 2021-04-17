{ pkgs }:
let
  controllerToolsVersion = "0.5.0";
  sha256 = "169nljr5718hlww7d21cq853qmjiszkfl9mxx1j1k7bdl4nps2vq";
  vendorSha256 = "0xgrs0b3z9pwa57bh24nhmm3rp2wi0ynhxjwyp6zsqas6snnxnzc";
in pkgs.buildGoModule {
  pname = "controller-tools";
  version = controllerToolsVersion;

  vendorSha256 = vendorSha256;
  subPackages = [ "cmd/controller-gen" ];

  doCheck = false;
  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "controller-tools";
    rev = "v${controllerToolsVersion}";
    sha256 = sha256;
  };

  meta = {
    description = "Tools to use with the controller-runtime libraries";
    homepage = "https://github.com/kubernetes-sigs/controller-tools";
    license = pkgs.lib.licenses.asl20;
  };
}
