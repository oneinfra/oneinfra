{ nixpkgs_revision ? "c576998594b4b8790f291d17fa92d499d1dc5d42",
  nixpkgs_sha256 ? "1k1xpardbbpb23wdki2ws30b3f20nd6fpx6lm802s1c6k3xh2d4c" }:
let
  pkgs = import (builtins.fetchTarball {
    name = "nixpkgs-${nixpkgs_revision}";
    url = "https://github.com/nixos/nixpkgs/archive/${nixpkgs_revision}.tar.gz";
    sha256 = nixpkgs_sha256;
  }) {};
in
pkgs.mkShell {
  buildInputs = with pkgs; [
    (import ./nix/code-generator.nix { inherit pkgs; })
    (import ./nix/controller-tools.nix { inherit pkgs; })
    (import ./nix/kubebuilder.nix { inherit pkgs; })
    go_1_16
    golint
    kustomize
    jq
    yq-go
  ];
}
