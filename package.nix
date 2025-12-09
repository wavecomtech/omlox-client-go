{
  lib,
  buildGoModule,
  version ? 0.0.0,
  commitHash ? "unknown",
}:

buildGoModule (finalAttrs: {
  pname = "omlox-cli";
  inherit version;

  src = ./.;

  vendorHash = "sha256-8pNpJ9brmL88CVbh9vlm/Cd4QSstZTIWYn5nFqfe2Xw=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
    "-X main.commitHash=${commitHash}"
  ];
})
