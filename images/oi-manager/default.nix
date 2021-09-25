let
  pkgs = import <nixpkgs.nix> {};
  oi-manager = import <oi-manager.nix> { pkgs = pkgs; };
in pkgs.dockerTools.buildImage {
  name = "oi-manager";
  tag = oi-manager.version;
  contents = [ oi-manager ];
  config = {
    Cmd = [ "/bin/oi-manager" ];
  };
  extraCommands = ''
    ln -s ${oi-manager}/bin/oi-manager /bin/oi-manager
  '';
}
