{ pkgs }:
let
  name = "code-generator";
  codeGeneratorVersion = "0.21.0";
  sha256 = "1xbwibyrscxvffdya53p258dardib9q34gjhzyi7lgb4pq6465h5";
  vendorSha256 = "0lhm6q6pf22773idj9idjqz9cdj4i6y1ms9wb9hlb516pprgyid4";
in pkgs.buildGoModule {
  pname = name;
  version = codeGeneratorVersion;

  vendorSha256 = vendorSha256;

  doCheck = false;
  runVend = true;

  src = pkgs.fetchFromGitHub {
    owner = "kubernetes";
    repo = name;
    rev = "v${codeGeneratorVersion}";
    sha256 = sha256;
  };

  meta = {
    description = "Generators for kube-like API types";
    homepage = "https://github.com/kubernetes/code-generator";
    license = pkgs.lib.licenses.asl20;
  };
}
