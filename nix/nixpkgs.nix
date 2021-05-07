{
  nixpkgs ? {
    revision = "c576998594b4b8790f291d17fa92d499d1dc5d42";
    sha256 = "1k1xpardbbpb23wdki2ws30b3f20nd6fpx6lm802s1c6k3xh2d4c";
  }
}:
import (builtins.fetchTarball {
  name = "nixpkgs-${nixpkgs.revision}";
  url = "https://github.com/nixos/nixpkgs/archive/${nixpkgs.revision}.tar.gz";
  sha256 = nixpkgs.sha256;
}) {}
