{ pkgs, lib, config, ... }:
{
  # Fix for devenv secretspec module
  _module.args.secretspec = null;

  # Packages - equivalent to buildInputs in flake.nix
  packages = [
    pkgs.bun
    pkgs.chromium # for headless testing
  ];

  # Enable dotenv to load from .env file
  dotenv.enable = true;

  # Enable Claude Code integration
  claude.code.enable = true;

  # MCP Servers for Claude Code integration
  claude.code.mcpServers = {
    context7 = {
      type = "stdio";
      command = "bunx";
      args = [ "--bun" "@upstash/context7-mcp" ];
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
          cat << 'EOF' | claude -p --model opus --output-format stream-json --verbose
You are a software planning agent. Your job is to analyze specifications against existing code and create a prioritized task list.

Study the specs/ folder to understand requirements.
Study the src/ folder (or equivalent) to understand what exists.
Study the current IMPLEMENTATION_PLAN.md (if it exists).

Compare specs against code. What's missing? What needs fixing?

Create or update IMPLEMENTATION_PLAN.md with a prioritized bullet-point list:
- [ ] Task 1: description (dependency notes if any)
- [ ] Task 2: description
- ... (sort by priority: highest impact / lowest risk first)

Important: Plan only. Do NOT implement anything.
Important: Don't assume functionality is missing—search the codebase first to confirm.

When the plan is complete and prioritized, output: <promise>PLAN_COMPLETE</promise>
EOF
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

          prompt_content=$(cat << 'EOF'
You are a software engineer building features from a plan.

Your job:
1. Read IMPLEMENTATION_PLAN.md
2. Choose the most important task not yet complete
3. Before changing anything, search the codebase—don't assume it's not implemented
4. Implement the feature according to specs/
5. Run tests and fix failures
6. Update IMPLEMENTATION_PLAN.md, commit, and exit

Study the specs/ folder to understand requirements.
Study the src/ folder (or equivalent) to understand current code.
Study IMPLEMENTATION_PLAN.md to pick the next task.

Step 1: Choose the most important incomplete task.
Step 2: Search the codebase to confirm it's not already implemented.
Step 3: Implement the feature. If tests exist, run them often.
Step 4: When tests pass, update IMPLEMENTATION_PLAN.md (mark task done or note blockers).
Step 5: Commit and push, following the convential commits specification

Important: Implement ONE task only. Don't try to do everything at once.
Important: Run tests frequently to catch issues early.
Important: If you find bugs unrelated to your task, fix them too—single source of truth.

When the task is complete, tests pass, and you've committed:

<promise>BUILD_COMPLETE</promise>
EOF
)

          if [ "$gemini_mode" = true ]; then
            echo "Using gemini --yolo"
            echo "$prompt_content" | bunx gemini -p --output-format stream-json --verbose --yolo
          else
            echo "Using claude"
            echo "$prompt_content" | claude -p --output-format stream-json --verbose --dangerously-skip-permissions
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
