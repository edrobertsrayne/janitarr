{ pkgs, lib, config, ... }:
let
  # Read .env file and extract CONTEXT7_API_KEY
  envFile = builtins.readFile ./.env;
  envLines = lib.splitString "\n" envFile;
  context7ApiKey = lib.findFirst
    (line: lib.hasPrefix "CONTEXT7_API_KEY=" line)
    ""
    envLines;
  apiKeyValue = if context7ApiKey != "" then
    lib.removePrefix "CONTEXT7_API_KEY=" context7ApiKey
  else
    "";
in
{
  # Fix for devenv secretspec module
  _module.args.secretspec = null;

  # Packages - equivalent to buildInputs in flake.nix
  packages = [
    pkgs.bun
    pkgs.chromium # for headless testing
  ];

  # Environment variables
  env.CONTEXT7_API_KEY = apiKeyValue;

  # Disable dotenv hint (we handle .env manually in this config)
  dotenv.disableHint = true;

  # Enable Claude Code integration
  claude.code.enable = true;

  # MCP Servers for Claude Code integration
  claude.code.mcpServers = {
    context7 = {
      type = "stdio";
      command = "bunx";
      args = [ "--bun" "@upstash/context7-mcp" ];
      env = {
        CONTEXT7_API_KEY = apiKeyValue;
      };
    };
  };

  # Optional: Helpful scripts for common workflows
  # scripts.test-server.exec = "bun run src/index.ts server test";

  # Optional: Enter shell message
  enterShell = ''
    echo "Janitarr development environment loaded"
    echo "Run 'bun install' to install dependencies"
    echo ""
    echo "MCP Servers configured for Claude Code:"
    echo "  - context7: Codebase context provider"
  '';
}
