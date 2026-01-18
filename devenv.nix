{
  pkgs,
  lib,
  config,
  ...
}: {
  # Fix for devenv secretspec module
  _module.args.secretspec = null;

  # Packages - equivalent to buildInputs in flake.nix
  packages = [
    pkgs.chromium # for headless testing
    pkgs.golangci-lint # comprehensive Go linting
    pkgs.gomod2nix # for Nix packaging
  ];

  # Enable dotenv to load from .env file
  dotenv.enable = true;

  env.CHROMIUM_PATH = "${pkgs.chromium}/bin/chromium";

  # Enable Claude Code integration
  claude.code.enable = true;

  languages = {
    javascript = {
      enable = true;
      bun = {
        enable = true;
        install.enable = true;
      };
    };
    go = {
      enable = true;
      version = "1.25.5"; # Pin to match go.mod
    };
    nix.enable = true;
  };

  # MCP Servers for Claude Code integration
  claude.code.mcpServers = {
    context7 = {
      type = "stdio";
      command = "bunx";
      args = ["--bun" "@upstash/context7-mcp"];
      env = {
        CONTEXT7_API_KEY = config.dotenv.vars.CONTEXT7_API_KEY or "";
      };
    };
  };

  git-hooks.hooks = {
    # Nix/Shell formatting
    alejandra.enable = true;
    shellcheck.enable = true;

    # JavaScript/CSS formatting
    eslint.enable = true;
    prettier.enable = true;

    # Go formatting and linting
    gofmt.enable = true;
    govet = {
      enable = true;
      pass_filenames = false; # Analyze whole packages
    };
    gotest.enable = true; # Ensure tests pass before commit
    golangci-lint = {
      enable = true;
      pass_filenames = false; # Analyze whole packages
    };
  };

  # Nix package outputs
  outputs = {
    app = pkgs.callPackage ./default.nix {
      inherit pkgs;
      name = "janitarr";
      version = "0.1.0";
    };
  };

  # Optional: Helpful scripts for common workflows
  scripts = {
    plan = {
      description = "Run the planning agent against the codebase";
      exec = ''
        iterations="''${1:-1}"
        current_iteration=0

        while [ "$current_iteration" -lt "$iterations" ]; do
          echo "Running plan iteration $((current_iteration + 1)) of $iterations"
          cat PROMPT_plan.md | claude -p --model opus --output-format stream-json --verbose --dangerously-skip-permissions
          current_iteration=$((current_iteration + 1))
        done
      '';
    };
    build = {
      description = "Run the build agent against the codebase. Use --gemini to run with Gemini.";
      exec = ''
        iterations=''${1:-5}
        current_iteration=0

        while [ "$current_iteration" -lt "$iterations" ]; do
          echo "Running build iteration $((current_iteration + 1)) of $iterations"
          cat PROMPT_build.md | claude -p --model sonnet --dangerously-skip-permissions
          current_iteration=$((current_iteration + 1))
        done
      '';
    };
  };

  # Optional: Enter shell message
  enterShell = ''
    echo "Janitarr development environment loaded"
    echo ""
    echo "Available tools:"
    echo "  - Go $(go version | cut -d' ' -f3) (pinned to match go.mod)"
    echo "  - templ - HTML template generation"
    echo "  - golangci-lint - comprehensive Go linting"
    echo "  - Bun - JavaScript runtime for E2E tests"
    echo "  - Playwright - E2E testing framework"
    echo ""
    echo "MCP Servers (Claude Code integration):"
    echo "  - context7 - Codebase context provider"
    echo ""
    echo "Quick commands:"
    echo "  make generate  - Generate templ templates + Tailwind CSS"
    echo "  make build     - Generate and build binary"
    echo "  make test      - Run tests with race detection"
    echo ""
    echo "First time? Run: bun install"
  '';
}
