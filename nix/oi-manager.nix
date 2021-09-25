{ pkgs ? import <nixpkgs> {} }:
let
  name = "oi-manager";
in pkgs.buildGoModule rec {
  pname = name;
  version = "20.09.0-alpha21";
  ldflags = ''
    -X github.com/oneinfra/oneinfra/internal/pkg/constants.BuildVersion=${version}
  '';

  subPackages = [ "cmd/oi-manager" ];

  vendorSha256 = null;

  src = pkgs.fetchFromGitHub {
    owner = "oneinfra";
    repo = "oneinfra";
    rev = version;
    sha256 = "17ma6p1nnnpr4599bzwvw51xxiny6b867dg0xm5yy1x78af0dh95";
  };

  meta = {
    description = "oneinfra manager";
    homepage = "https://oneinfra.net/";
    license = pkgs.lib.licenses.asl20;
  };
}
