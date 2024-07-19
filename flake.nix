{
  description = "Hello World";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      version = builtins.substring 0 1 self.lastModifiedDate;

      supportedSystems =
        [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (system:
        import nixpkgs {
          inherit system;
          overlays = [ (final: prev: { go = prev.go_1_22; }) ];
        });
    in {
      devShell = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in with pkgs;
        mkShell {
          buildInputs = [ go_1_21 gotools go-tools gopls nixpkgs-fmt ];
        });

      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in rec {
          bin = pkgs.buildGoModule {
            pname = "hello";
            inherit version;
            src = ./.;
            vendorHash = null;
          };

          docker = pkgs.dockerTools.buildLayeredImage {
            name = "hello";
            tag = "latest";
            config.Cmd =
              "${bin}/bin/learnGoWithTests"; # Depends on name in go.mod
          };
        });

    };
}
