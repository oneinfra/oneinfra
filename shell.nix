{
  code-generator ? {
    version = "0.21.0";
    sha256 = "1xbwibyrscxvffdya53p258dardib9q34gjhzyi7lgb4pq6465h5";
    vendorSha256 = "0lhm6q6pf22773idj9idjqz9cdj4i6y1ms9wb9hlb516pprgyid4";
  },
  controller-tools ? {
    version = "0.5.0";
    sha256 = "169nljr5718hlww7d21cq853qmjiszkfl9mxx1j1k7bdl4nps2vq";
    vendorSha256 = "0xgrs0b3z9pwa57bh24nhmm3rp2wi0ynhxjwyp6zsqas6snnxnzc";
  },
  kubebuilder ? {
    version = "2.3.2";
    sha256 = "10f48nmpkb3kx36x92a77mnrn48y6fvwq9dxlfw0r35hsrv1sm2g";
    vendorSha256 = "1v0ba2h1ld8l4kvsd0rajl2540v98pa8q19lwbxdll28rz83msh9";
  }
}:
let
  pkgs = import (./nix/nixpkgs.nix) {};
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    (import ./nix/code-generator.nix {
      inherit pkgs;
      inherit (code-generator) version sha256 vendorSha256;
    })
    (import ./nix/controller-tools.nix {
      inherit pkgs;
      inherit (controller-tools) version sha256 vendorSha256;
    })
    (import ./nix/kubebuilder.nix {
      inherit pkgs;
      inherit (kubebuilder) version sha256 vendorSha256;
    })
    cri-tools
    docker
    git
    go_1_16
    golint
    kubectl
    kustomize
    jq
    yq-go
  ];

  shellHook = ''
    export GOPATH=$HOME/go
    export PATH=$GOPATH/bin:$PATH
  '';
}
