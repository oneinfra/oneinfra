let
  pkgs = import <nixpkgs.nix> {};
  oi = import <oi.nix> { pkgs = pkgs; };
in pkgs.dockerTools.buildImage {
  name = "oi";
  tag = oi.version;
  contents = [ oi ];
  config = {
    Cmd = [ "/bin/oi" ];
  };
  extraCommands = ''
    ln -sf ${oi}/bin/oi /bin/oi
  '';
}
