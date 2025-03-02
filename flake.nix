{
  description = "mines server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs, ... }: 
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        devShell = pkgs.mkShell {
          packages = [ 
            pkgs.go
            pkgs.air
            pkgs.fish 
          ];

          # shellHook = ''
          #   source .env
          # '';
        };
      }
    );
}
