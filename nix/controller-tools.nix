{ pkgs, version, sha256, vendorSha256 }:
let
  name = "controller-tools";
in pkgs.buildGoModule {
  pname = name;
  version = version;

  vendorSha256 = vendorSha256;
  subPackages = [ "cmd/controller-gen" ];

  doCheck = false;
  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = name;
    rev = "v${version}";
    sha256 = sha256;
  };

  meta = {
    description = "Tools to use with the controller-runtime libraries";
    homepage = "https://github.com/kubernetes-sigs/controller-tools";
    license = pkgs.lib.licenses.asl20;
  };
}
