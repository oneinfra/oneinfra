{ pkgs }:
pkgs.buildGoModule rec {
  pname = "controller-tools";
  version = "0.5.0";

  vendorSha256 = "0xgrs0b3z9pwa57bh24nhmm3rp2wi0ynhxjwyp6zsqas6snnxnzc";
  subPackages = [ "cmd/controller-gen" ];

  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "controller-tools";
    rev = "v${version}";
    sha256 = "169nljr5718hlww7d21cq853qmjiszkfl9mxx1j1k7bdl4nps2vq";
  };

  meta = {
    description = "Tools to use with the controller-runtime libraries";
    homepage = "https://github.com/kubernetes-sigs/controller-tools";
    license = pkgs.lib.licenses.asl20;
  };
}
