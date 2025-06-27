{
  description = "A Go client for the Omlox Hub™";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
      ...
    }:
    let
      supportedSystems = [
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
      nixpkgsFor = forAllSystems (
        system:
        import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
        }
      );

      version = self.shortRev or self.dirtyShortRev;
      commitHash = self.rev or self.dirtyRev;
    in
    {
      overlays.default = final: _: {
        omlox-cli = final.callPackage ./package.nix {
          inherit version commitHash;
        };
      };

      formatter = forAllSystems (system: (nixpkgsFor.${system}).nixfmt-tree);

      packages = forAllSystems (system: {
        default = (nixpkgsFor.${system}).omlox-cli;
        omlox-cli = (nixpkgsFor.${system}).omlox-cli;
      });

      devShells = forAllSystems (
        system: with nixpkgsFor.${system}; {
          default = mkShell {
            inputsFrom = [ omlox-cli ];
            packages = [
              git
              gnumake
              easyjson
              goreleaser
              copywrite
            ];
          };
        }
      );
    };
}
