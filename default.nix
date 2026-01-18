{
  pkgs,
  name ? "janitarr",
  version ? "0.1.0",
  ...
}:
pkgs.buildGoApplication {
  pname = name;
  inherit version;

  src = builtins.path {
    path = ./.;
    name = "source";
  };

  # Reference the gomod2nix dependency specifications
  modules = ./gomod2nix.toml;

  # Build from src/ directory
  subPackages = ["src"];

  # Build flags matching Makefile
  ldflags = ["-s" "-w"];
}
