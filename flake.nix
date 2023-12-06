{
  description = "A Go client for the Omlox Hub™";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        buildDeps = with pkgs; [ git go_1_21 gnumake ];
        devDeps = with pkgs;
          buildDeps ++ [
            easyjson
            goreleaser
          ];
      in
      { devShell = pkgs.mkShell { buildInputs = devDeps; }; }
    );

}
