{ pkgs, version, sha256, vendorSha256 }:
let
  name = "code-generator";
in pkgs.buildGoModule {
  pname = name;
  version = version;

  allowGoReference = true;
  doCheck = false;
  runVend = true;
  vendorSha256 = vendorSha256;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes";
    repo = name;
    rev = "v${version}";
    sha256 = sha256;
  };

  meta = {
    description = "Generators for kube-like API types";
    homepage = "https://github.com/kubernetes/code-generator";
    license = pkgs.lib.licenses.asl20;
  };
}
