{ pkgs ? import <nixpkgs> {} }:
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
