{ pkgs, version, sha256, vendorSha256 }:
let
  name = "kubebuilder";
in pkgs.buildGoModule {
  pname = name;
  version = version;

  vendorSha256 = vendorSha256;
  subPackages = [ "cmd" ];

  doCheck = false;
  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = name;
    rev = "v${version}";
    sha256 = sha256;
  };

  postInstall = ''
    mv $out/bin/cmd $out/bin/kubebuilder
  '';

  meta = {
    description = "SDK for building Kubernetes APIs using CRDs";
    homepage = "https://kubebuilder.io/";
    license = pkgs.lib.licenses.asl20;
  };
}
