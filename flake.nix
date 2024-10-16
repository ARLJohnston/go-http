{
  description = "Flake for go-http server with various instrumentations";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
  let
    supportedSystems =
      [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      nixpkgsFor = forAllSystems (system:
      import nixpkgs {
        inherit system;
      });
  in {
    devShell = forAllSystems (system:
    let pkgs = nixpkgsFor.${system};
    in with pkgs;
    mkShell {
      buildInputs = [
        go_1_23
        gotools
        go-tools
        gopls
        nixpkgs-fmt
        yaml-language-server
        opentofu
        terraform-ls
        nixd
        delve
        dockerfile-language-server-nodejs
      ];

      CGO_ENABLED="0";
    });
  };
}
