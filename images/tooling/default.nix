let
  pkgs = import <nixpkgs.nix> {};
in pkgs.dockerTools.buildImage {
  name = "tooling";
  tag = "latest"; # FIXME (ereslibre)
  contents = [
    pkgs.dbus
    (pkgs.writeScriptBin "write-base64-file.sh" ''
      #!${pkgs.runtimeShell}
      ${pkgs.coreutils}/bin/echo "$1" | ${pkgs.coreutils}/bin/base64 -d - > "$2"
    '')
  ];
}
