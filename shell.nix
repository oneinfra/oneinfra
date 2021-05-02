{ nixpkgs_revision ? "c576998594b4b8790f291d17fa92d499d1dc5d42" }:
let
  pkgs = import (builtins.fetchTarball {
    name = "nixpkgs-${nixpkgs_revision}";
    url = "https://github.com/nixos/nixpkgs/archive/${nixpkgs_revision}.tar.gz";
    sha256 = "1k1xpardbbpb23wdki2ws30b3f20nd6fpx6lm802s1c6k3xh2d4c";
  }) {};
in
pkgs.mkShell {
  buildInputs = [
    (import ./nix/code-generator.nix { inherit pkgs; })
    (import ./nix/controller-tools.nix { inherit pkgs; })
    (import ./nix/kubebuilder.nix { inherit pkgs; })
    pkgs.go_1_16
    pkgs.golint
    pkgs.kustomize
    pkgs.jq
    pkgs.yq-go
  ];
}
