{ pkgs }:
pkgs.buildGoModule rec {
  pname = "kubebuilder";
  version = "2.3.2";

  vendorSha256 = "1v0ba2h1ld8l4kvsd0rajl2540v98pa8q19lwbxdll28rz83msh9";
  subPackages = [ "cmd" ];

  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes-sigs";
    repo = "kubebuilder";
    rev = "v${version}";
    sha256 = "10f48nmpkb3kx36x92a77mnrn48y6fvwq9dxlfw0r35hsrv1sm2g";
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
