{
  nixpkgs ? {
    revision = "7e9b0dff974c89e070da1ad85713ff3c20b0ca97";
    sha256 = "1ckzhh24mgz6jd1xhfgx0i9mijk6xjqxwsshnvq789xsavrmsc36";
  }
}:
import (builtins.fetchTarball {
  name = "nixpkgs-${nixpkgs.revision}";
  url = "https://github.com/nixos/nixpkgs/archive/${nixpkgs.revision}.tar.gz";
  sha256 = nixpkgs.sha256;
}) {}
