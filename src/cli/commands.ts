/**
 * CLI Command Implementations
 *
 * Defines all Janitarr CLI commands using Commander.js and connects them
 * to the backend services.
 */

import { Command } from "commander";
import * as readline from "node:readline";
import type { ServerType } from "../types";
import {
  addServer,
  listServers,
  updateServer, // Changed from editServer
  removeServer,
  testServerConnection,
  getServer,
} from "../services/server-manager";
import { detectAll } from "../services/detector";
import { runAutomationCycle } from "../services/automation";
import { getDatabase } from "../storage/database";
import {
  getScheduleConfig,
  setScheduleConfig,
  start as startScheduler,
  stop as stopScheduler,
  getStatus,
  getTimeUntilNextRun,
  registerCycleCallback,
  isRunning as isSchedulerRunning,
} from "../lib/scheduler";
import * as fmt from "./formatters";

/**
 * Create readline interface for user input
 */
function createInterface() {
  return readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });
}

/**
 * Prompt user for input
 */
function question(rl: readline.Interface, prompt: string): Promise<string> {
  return new Promise((resolve) => {
    rl.question(fmt.prompt(prompt) + " ", resolve);
  });
}

/**
 * Confirm action with user
 */
async function confirm(rl: readline.Interface, message: string): Promise<boolean> {
  const answer = await question(rl, `${message} (y/N)`);
  return answer.toLowerCase() === "y" || answer.toLowerCase() === "yes";
}

/**
 * Create the main CLI program
 */
export function createProgram(): Command {
  const program = new Command();

  program
    .name("janitarr")
    .description("Automation tool for Radarr and Sonarr media servers")
    .version("0.1.0");

  // Server management commands
  const serverCmd = program
    .command("server")
    .description("Manage Radarr/Sonarr server configurations");

  serverCmd
    .command("add")
    .description("Add a new media server")
    .action(async () => {
      const rl = createInterface();

      try {
        console.log(fmt.header("Add New Server"));
        console.log();

        const name = await question(rl, "Server name");
        if (!name.trim()) {
          console.log(fmt.error("Server name is required"));
          rl.close();
          return;
        }

        const typeAnswer = await question(rl, "Server type (radarr/sonarr)");
        const type = typeAnswer.toLowerCase().trim() as ServerType;
        if (type !== "radarr" && type !== "sonarr") {
          console.log(fmt.error("Invalid server type. Must be 'radarr' or 'sonarr'"));
          rl.close();
          return;
        }

        const url = await question(rl, "Server URL (e.g., http://localhost:7878)");
        if (!url.trim()) {
          console.log(fmt.error("Server URL is required"));
          rl.close();
          return;
        }

        const apiKey = await question(rl, "API Key");
        if (!apiKey.trim()) {
          console.log(fmt.error("API key is required"));
          rl.close();
          return;
        }

        console.log();
        fmt.showProgress("Testing connection");

        const result = await addServer(name, url, apiKey, type);

        fmt.clearLine();

        if (result.success) {
          console.log(fmt.success(`Server '${name}' added successfully`));
        } else {
          console.log(fmt.error(`Failed to add server: ${result.error}`));
        }
      } finally {
        rl.close();
      }
    });

  serverCmd
    .command("list")
    .description("List all configured servers")
    .option("--json", "Output as JSON")
    .action(async (options) => {
      const servers = await listServers();

      if (options.json) {
        console.log(fmt.formatServerJson(servers));
      } else {
        console.log(fmt.formatServerTable(servers));
      }
    });

  serverCmd
    .command("edit <id-or-name>")
    .description("Edit an existing server")
    .action(async (idOrName: string) => {
      const rl = createInterface();

      try {
        const serverResult = await getServer(idOrName);
        if (!serverResult.success) {
          console.log(fmt.error(serverResult.error));
          rl.close();
          return;
        }

        const server = serverResult.data;

        console.log(fmt.header(`Edit Server: ${server.name}`));
        console.log();
        console.log(fmt.info("Press Enter to keep current value"));
        console.log();

        const url = await question(rl, `URL [${server.url}]`);
        const apiKey = await question(rl, `API Key [${server.apiKey.slice(0, 8)}...]`);

        const newUrl = url.trim() || server.url;
        const newApiKey = apiKey.trim() || server.apiKey;

        if (newUrl === server.url && newApiKey === server.apiKey) {
          console.log(fmt.warning("No changes made"));
          rl.close();
          return;
        }

        console.log();
        fmt.showProgress("Testing connection");

        const result = await updateServer(getDatabase(), server.id, { url: newUrl, apiKey: newApiKey });

        fmt.clearLine();

        if (result.success) {
          console.log(fmt.success(`Server '${server.name}' updated successfully`));
        } else {
          console.log(fmt.error(`Failed to update server: ${result.error}`));
        }
      } finally {
        rl.close();
      }
    });

  serverCmd
    .command("remove <id-or-name>")
    .description("Remove a server")
    .action(async (idOrName: string) => {
      const rl = createInterface();

      try {
        const serverResult = await getServer(idOrName);
        if (!serverResult.success) {
          console.log(fmt.error(serverResult.error));
          rl.close();
          return;
        }

        const server = serverResult.data;

        const confirmed = await confirm(
          rl,
          `Remove server '${server.name}'?`
        );

        if (!confirmed) {
          console.log(fmt.info("Cancelled"));
          rl.close();
          return;
        }

        const result = await removeServer(server.id);

        if (result.success) {
          console.log(fmt.success(`Server '${server.name}' removed`));
        } else {
          console.log(fmt.error(`Failed to remove server: ${result.error}`));
        }
      } finally {
        rl.close();
      }
    });

  serverCmd
    .command("test <id-or-name>")
    .description("Test connection to a server")
    .action(async (idOrName: string) => {
      const serverResult = await getServer(idOrName);
      if (!serverResult.success) {
        console.log(fmt.error(serverResult.error));
        return;
      }

      const server = serverResult.data;

      fmt.showProgress(`Testing connection to ${server.name}`);

      const result = await testServerConnection(server.id);

      fmt.clearLine();

      if (result.success) {
        console.log(fmt.success(`Connection to '${server.name}' successful`));
        console.log(fmt.keyValue("Version", result.data.version));
      } else {
        console.log(fmt.error(`Connection failed: ${result.error}`));
      }
    });

  // Status and detection commands
  program
    .command("status")
    .description("Show scheduler status and configuration")
    .option("--json", "Output as JSON")
    .action(async (options) => {
      const status = getStatus();
      const servers = await listServers();
      const db = getDatabase();
      const config = db.getAppConfig();

      if (options.json) {
        console.log(
          JSON.stringify(
            {
              scheduler: status,
              servers: servers.length,
              config,
            },
            null,
            2
          )
        );
      } else {
        console.log(fmt.header("Janitarr Status"));
        console.log();

        console.log(fmt.keyValue("Servers configured", servers.length.toString()));
        console.log();

        console.log(fmt.keyValue("Scheduler", status.isRunning ? fmt.success("Running") : fmt.warning("Stopped")));

        if (status.isRunning) {
          const timeUntilNext = getTimeUntilNextRun();
          if (timeUntilNext !== null) {
            const minutes = Math.ceil(timeUntilNext / 60000);
            console.log(fmt.keyValue("Next run", `${minutes} minutes`));
          }
        }

        if (status.isCycleActive) {
          console.log(fmt.info("Automation cycle in progress"));
        }

        console.log();
        console.log(fmt.formatConfig(config));
      }
    });

  program
    .command("scan")
    .description("Scan servers for missing/cutoff content (no searches)")
    .option("--json", "Output as JSON")
    .action(async (options) => {
      fmt.showProgress("Scanning servers");

      const results = await detectAll();

      fmt.clearLine();

      if (options.json) {
        console.log(JSON.stringify(results, null, 2));
      } else {
        console.log(fmt.formatDetectionSummary(results.results));
      }
    });

  // Automation commands
  program
    .command("run")
    .description("Execute automation cycle immediately")
    .option("--dry-run", "Preview what would be searched without triggering searches")
    .option("--json", "Output as JSON")
    .action(async (options) => {
      const dryRun = options.dryRun === true;

      if (dryRun) {
        fmt.showProgress("Running automation cycle (DRY RUN - no searches will be triggered)");
      } else {
        fmt.showProgress("Running automation cycle");
      }

      const result = await runAutomationCycle(true, dryRun);

      fmt.clearLine();

      if (options.json) {
        console.log(JSON.stringify(result, null, 2));
      } else {
        if (dryRun) {
          console.log(fmt.info("=== DRY RUN MODE - Preview Only ==="));
          console.log();
        }

        console.log(fmt.formatCycleSummary(result));

        if (result.errors.length > 0) {
          console.log();
          console.log(fmt.error("Errors occurred:"));
          for (const err of result.errors) {
            console.log(fmt.error(`  ${err}`));
          }
        }

        if (dryRun) {
          console.log();
          console.log(fmt.info("No searches were triggered. Run without --dry-run to execute."));
        }
      }
    });

  program
    .command("start")
    .description("Start scheduler and web server (production mode)")
    .option("-p, --port <number>", "Port to listen on", "3434")
    .option("-h, --host <string>", "Host to bind to", "localhost")
    .action(async (options) => {
      const port = parseInt(options.port, 10);
      const host = options.host;

      // Validate port number
      if (isNaN(port) || port < 1 || port > 65535) {
        console.log(fmt.error("Invalid port number. Must be between 1 and 65535"));
        process.exit(1);
      }

      console.log(fmt.header("Starting Janitarr"));
      console.log();

      const db = getDatabase();
      const { createWebServer } = await import("../web/server");

      let webServer: ReturnType<typeof createWebServer> | null = null;

      // Start web server
      try {
        webServer = createWebServer({ port, host, db, silent: true });
        console.log(fmt.success("✓ Web server started"));
        console.log(fmt.info(`  Web UI: http://${host}:${port}`));
        console.log(fmt.info(`  API: http://${host}:${port}/api`));
        console.log(fmt.info(`  Health: http://${host}:${port}/api/health`));
        console.log(fmt.info(`  Metrics: http://${host}:${port}/metrics`));
        console.log();
      } catch (error) {
        console.log(fmt.error(`Failed to start web server: ${error instanceof Error ? error.message : String(error)}`));
        process.exit(1);
      }

      // Start scheduler if enabled
      const config = getScheduleConfig();
      if (!config.enabled) {
        console.log(fmt.warning("⚠ Scheduler is disabled in configuration"));
        console.log(fmt.info("  Web server running, but automation cycles will not run"));
        console.log(fmt.info("  Enable with: janitarr config set schedule.enabled true"));
        console.log();
      } else {
        if (isSchedulerRunning()) {
          console.log(fmt.warning("⚠ Scheduler is already running"));
          console.log();
        } else {
          // Register automation cycle callback
          registerCycleCallback(async (isManual: boolean) => {
            await runAutomationCycle(isManual);
          });

          await startScheduler();
          console.log(fmt.success("✓ Scheduler started"));
          console.log(fmt.info(`  Interval: ${config.intervalHours} hours`));
          console.log();
        }
      }

      console.log(fmt.info("Press Ctrl+C to stop"));

      // Graceful shutdown on SIGINT
      process.on("SIGINT", () => {
        console.log();
        console.log(fmt.info("Shutting down gracefully..."));

        // Stop scheduler
        if (isSchedulerRunning()) {
          stopScheduler();
          console.log(fmt.success("✓ Scheduler stopped"));
        }

        // Stop web server
        if (webServer) {
          webServer.stop();
          console.log(fmt.success("✓ Web server stopped"));
        }

        console.log(fmt.success("Shutdown complete"));
        process.exit(0);
      });

      // Keep the process running
      await new Promise(() => {}); // Never resolves
    });

  program
    .command("stop")
    .description("Stop running scheduler daemon")
    .action(() => {
      if (!isSchedulerRunning()) {
        console.log(fmt.warning("Scheduler is not running"));
        return;
      }

      stopScheduler();
      console.log(fmt.success("Scheduler stopped"));
    });

  // Configuration commands
  const configCmd = program
    .command("config")
    .description("Manage configuration");

  configCmd
    .command("show")
    .description("Show current configuration")
    .option("--json", "Output as JSON")
    .action((options) => {
      const db = getDatabase();
      const config = db.getAppConfig();

      if (options.json) {
        console.log(fmt.formatConfigJson(config));
      } else {
        console.log(fmt.formatConfig(config));
      }
    });

  configCmd
    .command("set <key> <value>")
    .description("Set configuration value")
    .action((key: string, value: string) => {
      const db = getDatabase();

      try {
        switch (key) {
          case "schedule.interval": {
            const hours = parseInt(value, 10);
            if (isNaN(hours) || hours < 1) {
              console.log(fmt.error("Interval must be a number >= 1"));
              return;
            }
            setScheduleConfig(hours, undefined);
            console.log(fmt.success(`Schedule interval set to ${hours} hours`));
            break;
          }

          case "schedule.enabled": {
            const enabled = value.toLowerCase() === "true" || value === "1";
            setScheduleConfig(undefined, enabled);
            console.log(
              fmt.success(`Schedule ${enabled ? "enabled" : "disabled"}`)
            );
            break;
          }

          case "limits.missing.movies": {
            const limit = parseInt(value, 10);
            if (isNaN(limit) || limit < 0) {
              console.log(fmt.error("Limit must be a number >= 0"));
              return;
            }
            db.setAppConfig({ searchLimits: { missingMoviesLimit: limit } });
            console.log(
              fmt.success(
                limit === 0
                  ? "Missing movies searches disabled"
                  : `Missing movies limit set to ${limit}`
              )
            );
            break;
          }

          case "limits.missing.episodes": {
            const limit = parseInt(value, 10);
            if (isNaN(limit) || limit < 0) {
              console.log(fmt.error("Limit must be a number >= 0"));
              return;
            }
            db.setAppConfig({ searchLimits: { missingEpisodesLimit: limit } });
            console.log(
              fmt.success(
                limit === 0
                  ? "Missing episodes searches disabled"
                  : `Missing episodes limit set to ${limit}`
              )
            );
            break;
          }

          case "limits.cutoff.movies": {
            const limit = parseInt(value, 10);
            if (isNaN(limit) || limit < 0) {
              console.log(fmt.error("Limit must be a number >= 0"));
              return;
            }
            db.setAppConfig({ searchLimits: { cutoffMoviesLimit: limit } });
            console.log(
              fmt.success(
                limit === 0
                  ? "Cutoff movies searches disabled"
                  : `Cutoff movies limit set to ${limit}`
              )
            );
            break;
          }

          case "limits.cutoff.episodes": {
            const limit = parseInt(value, 10);
            if (isNaN(limit) || limit < 0) {
              console.log(fmt.error("Limit must be a number >= 0"));
              return;
            }
            db.setAppConfig({ searchLimits: { cutoffEpisodesLimit: limit } });
            console.log(
              fmt.success(
                limit === 0
                  ? "Cutoff episodes searches disabled"
                  : `Cutoff episodes limit set to ${limit}`
              )
            );
            break;
          }

          default:
            console.log(fmt.error(`Unknown configuration key: ${key}`));
            console.log(
              fmt.info(
                "Valid keys: schedule.interval, schedule.enabled, limits.missing.movies, limits.missing.episodes, limits.cutoff.movies, limits.cutoff.episodes"
              )
            );
        }
      } catch (err) {
        console.log(
          fmt.error(`Failed to set configuration: ${err instanceof Error ? err.message : String(err)}`)
        );
      }
    });

  // Log commands
  program
    .command("logs")
    .description("Display activity logs")
    .option("-n, --limit <number>", "Number of entries to show", "50")
    .option("--all", "Show all logs")
    .option("--json", "Output as JSON")
    .option("--clear", "Clear all logs (requires confirmation)")
    .action(async (options) => {
      const db = getDatabase();

      if (options.clear) {
        const rl = createInterface();

        try {
          const confirmed = await confirm(rl, "Clear all logs?");

          if (!confirmed) {
            console.log(fmt.info("Cancelled"));
            rl.close();
            return;
          }

          db.clearLogs();
          console.log(fmt.success("All logs cleared"));
        } finally {
          rl.close();
        }

        return;
      }

      const limit = options.all ? undefined : parseInt(options.limit, 10);
      const logs = db.getLogs(limit);

      if (options.json) {
        console.log(fmt.formatLogJson(logs));
      } else {
        console.log(fmt.formatLogTable(logs));
      }
    });

  // Web server command
  program
    .command("serve")
    .description("Start the web server")
    .option("-p, --port <number>", "Port to listen on", "3000")
    .option("-h, --host <string>", "Host to bind to", "localhost")
    .action(async (options) => {
      const port = parseInt(options.port, 10);
      const host = options.host;

      if (isNaN(port) || port < 1 || port > 65535) {
        console.log(fmt.error("Invalid port number. Must be between 1 and 65535"));
        return;
      }

      console.log(fmt.header("Starting Janitarr Web Server"));
      console.log();

      const db = getDatabase();
      const { createWebServer } = await import("../web/server");

      try {
        createWebServer({ port, host, db });

        console.log();
        console.log(fmt.success("Web server started successfully"));
        console.log(fmt.info(`Access the web UI at: http://${host}:${port}`));
        console.log(fmt.info("Press Ctrl+C to stop the server"));

        // Keep the process running
        await new Promise(() => {});
      } catch (error) {
        console.log(fmt.error(`Failed to start web server: ${error instanceof Error ? error.message : String(error)}`));
        process.exit(1);
      }
    });

  return program;
}
