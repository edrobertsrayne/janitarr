{ pkgs, ... }: {
  # Packages - equivalent to buildInputs in flake.nix
  packages = [
    pkgs.bun
  ];

  # Environment variables (if needed in future)
  # env.JANITARR_DB_PATH = "./data/janitarr.db";

  # Optional: Helpful scripts for common workflows
  # scripts.test-server.exec = "bun run src/index.ts server test";

  # Optional: Enter shell message
  # enterShell = ''
  #   echo "Janitarr development environment loaded"
  #   echo "Run 'bun install' to install dependencies"
  # '';
}
