{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    (import ./nix/kubebuilder.nix { inherit pkgs; })
    (import ./nix/controller-tools.nix { inherit pkgs; })
    pkgs.go_1_16
    pkgs.golint
    pkgs.kustomize
    pkgs.jq
    pkgs.yq-go
  ];
}
