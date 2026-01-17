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
                iterations_arg=""
                gemini_mode=false

                # Parse arguments
                while (( "$#" )); do
                  case "$1" in
                    --gemini)
                      gemini_mode=true
                      shift
                      ;;
                    *)
                      iterations_arg="$1"
                      shift
                      ;;
                  esac
                done

                iterations=''${iterations_arg:-5}
                current_iteration=0

                while [ "$current_iteration" -lt "$iterations" ]; do
                  echo "Running build iteration $((current_iteration + 1)) of $iterations"

                  if [ "$gemini_mode" = true ]; then
                    echo "Using gemini --yolo"
                    cat PROMPT_build.md | bunx gemini --output-format stream-json --yolo
                  else
                    echo "Using claude"
                    cat PROMPT_build.md | claude -p --output-format stream-json --verbose --dangerously-skip-permissions
                  fi

                  current_iteration=$((current_iteration + 1))
                done
      '';
    };
  };

  # Optional: Enter shell message
  enterShell = ''
    echo "Janitarr development environment loaded"
    echo "Run 'bun install' to install dependencies"
    echo ""
    echo "MCP Servers configured for Claude Code:"
    echo "  - context7: Codebase context provider"
  '';
}
