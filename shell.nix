{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    (import ./nix/kubebuilder.nix { inherit pkgs; })
    (import ./nix/controller-tools.nix { inherit pkgs; })
    pkgs.kustomize
  ];
}
