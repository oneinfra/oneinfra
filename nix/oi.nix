{ pkgs ? import <nixpkgs> {} }:
let
  name = "oi";
  inherit (pkgs) dhallToNix;
in pkgs.buildGoModule rec {
  pname = name;
  version = (dhallToNix "(${../dhall}/versions.dhall).oneinfraVersion");
  ldflags = ''
    -X github.com/oneinfra/oneinfra/internal/pkg/constants.BuildVersion=${version}
  '';

  subPackages = [ "cmd/oi" ];

  vendorSha256 = null;

  src = pkgs.fetchFromGitHub {
    owner = "oneinfra";
    repo = "oneinfra";
    rev = version;
    sha256 = "17ma6p1nnnpr4599bzwvw51xxiny6b867dg0xm5yy1x78af0dh95";
  };

  meta = {
    description = "oneinfra CLI tool";
    homepage = "https://oneinfra.net/";
    license = pkgs.lib.licenses.asl20;
  };
}
