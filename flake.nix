{
  description = "Janitarr - Automation tool for Radarr and Sonarr";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachSystem ["x86_64-linux" "aarch64-linux"] (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
        janitarr = pkgs.callPackage ./nix/package.nix {};
      in {
        packages = {
          janitarr = janitarr;
          default = janitarr;
        };
      }
    )
    // {
      nixosModules = {
        janitarr = import ./nix/module.nix;
        default = self.nixosModules.janitarr;
      };
    };
}
