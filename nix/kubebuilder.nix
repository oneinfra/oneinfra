{ pkgs }:
let
  name = "kubebuilder";
  kubebuilderVersion = "2.3.2";
  sha256 = "10f48nmpkb3kx36x92a77mnrn48y6fvwq9dxlfw0r35hsrv1sm2g";
  vendorSha256 = "1v0ba2h1ld8l4kvsd0rajl2540v98pa8q19lwbxdll28rz83msh9";
in pkgs.buildGoModule {
  pname = name;
  version = kubebuilderVersion;

  vendorSha256 = vendorSha256;
  subPackages = [ "cmd" ];

  doCheck = false;
  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = name;
    rev = "v${kubebuilderVersion}";
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
