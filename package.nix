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

  vendorHash = "sha256-Gwkse7dl/EYxHSibgc6PMGwtDhKiMJYFajwAICpluzw=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
    "-X main.commitHash=${commitHash}"
  ];
})