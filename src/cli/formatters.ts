/**
 * CLI Output Formatters
 *
 * Provides colored output, table formatting, and various display utilities
 * for the Janitarr CLI.
 */

import chalk from "chalk";
import type { ServerInfo } from "../services/server-manager";
import type { LogEntry, AppConfig } from "../types";

/** Format a success message */
export function success(message: string): string {
  return chalk.green(`✓ ${message}`);
}

/** Format an error message */
export function error(message: string): string {
  return chalk.red(`✗ ${message}`);
}

/** Format a warning message */
export function warning(message: string): string {
  return chalk.yellow(`⚠ ${message}`);
}

/** Format an info message */
export function info(message: string): string {
  return chalk.blue(`ℹ ${message}`);
}

/** Format a header */
export function header(message: string): string {
  return chalk.bold.cyan(message);
}

/** Format a key-value pair */
export function keyValue(key: string, value: string): string {
  return `${chalk.gray(key + ":")} ${value}`;
}

/**
 * Format a list of servers as a table
 */
export function formatServerTable(servers: ServerInfo[]): string {
  if (servers.length === 0) {
    return chalk.gray("No servers configured");
  }

  const lines: string[] = [
    header("Configured Servers"),
    "",
  ];

  // Calculate column widths
  const nameWidth = Math.max(
    ...servers.map((s) => s.name.length),
    4
  );
  const typeWidth = 6; // "radarr" or "sonarr"
  const urlWidth = Math.max(
    ...servers.map((s) => s.url.length),
    3
  );

  // Header row
  const headerRow = [
    "NAME".padEnd(nameWidth),
    "TYPE".padEnd(typeWidth),
    "URL".padEnd(urlWidth),
    "API KEY",
  ].join("  ");
  lines.push(chalk.bold(headerRow));
  lines.push("-".repeat(headerRow.length));

  // Data rows
  for (const server of servers) {
    const typeColor = server.type === "radarr" ? chalk.magenta : chalk.cyan;
    lines.push(
      [
        chalk.white(server.name.padEnd(nameWidth)),
        typeColor(server.type.padEnd(typeWidth)),
        chalk.gray(server.url.padEnd(urlWidth)),
        chalk.dim(server.maskedApiKey),
      ].join("  ")
    );
  }

  return lines.join("\n");
}

/**
 * Format server info as JSON
 */
export function formatServerJson(servers: ServerInfo[]): string {
  return JSON.stringify(servers, null, 2);
}

/**
 * Format log entries as a table
 */
export function formatLogTable(logs: LogEntry[]): string {
  if (logs.length === 0) {
    return chalk.gray("No log entries");
  }

  const lines: string[] = [
    header("Activity Logs"),
    "",
  ];

  for (const log of logs) {
    const timestamp = formatTimestamp(log.timestamp);
    const prefix = getLogPrefix(log.type);
    const manual = log.isManual ? chalk.gray(" [manual]") : "";

    lines.push(`${chalk.gray(timestamp)} ${prefix} ${log.message}${manual}`);

    // Add details for search logs
    if (log.type === "search" && log.serverName) {
      const serverType = log.serverType === "radarr" ? chalk.magenta("Radarr") : chalk.cyan("Sonarr");
      lines.push(chalk.gray(`  └─ ${serverType} - ${log.serverName}${log.count ? ` (${log.count} items)` : ""}`));
    }
  }

  return lines.join("\n");
}

/**
 * Format log entries as JSON
 */
export function formatLogJson(logs: LogEntry[]): string {
  return JSON.stringify(logs, null, 2);
}

/**
 * Get colored prefix for log type
 */
function getLogPrefix(type: string): string {
  switch (type) {
    case "cycle_start":
      return chalk.blue("▶");
    case "cycle_end":
      return chalk.green("■");
    case "search":
      return chalk.cyan("→");
    case "error":
      return chalk.red("✗");
    default:
      return chalk.gray("•");
  }
}

/**
 * Format timestamp for display
 */
function formatTimestamp(date: Date): string {
  const d = new Date(date);
  const now = new Date();
  const isToday = d.toDateString() === now.toDateString();

  if (isToday) {
    // Show time only for today's entries
    return d.toLocaleTimeString("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } else {
    // Show date and time for older entries
    return d.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
}

/**
 * Format configuration as key-value pairs
 */
export function formatConfig(config: AppConfig): string {
  const lines: string[] = [
    header("Configuration"),
    "",
  ];

  lines.push(chalk.bold("Schedule:"));
  lines.push(keyValue("  Enabled", config.schedule.enabled ? chalk.green("Yes") : chalk.red("No")));
  lines.push(keyValue("  Interval", `${config.schedule.intervalHours} hours`));
  lines.push("");

  lines.push(chalk.bold("Search Limits:"));
  lines.push(
    keyValue(
      "  Missing content",
      config.searchLimits.missingLimit === 0
        ? chalk.gray("Disabled")
        : `${config.searchLimits.missingLimit} items`
    )
  );
  lines.push(
    keyValue(
      "  Quality cutoff",
      config.searchLimits.cutoffLimit === 0
        ? chalk.gray("Disabled")
        : `${config.searchLimits.cutoffLimit} items`
    )
  );

  return lines.join("\n");
}

/**
 * Format configuration as JSON
 */
export function formatConfigJson(config: AppConfig): string {
  return JSON.stringify(config, null, 2);
}

/**
 * Format cycle result summary
 */
export function formatCycleSummary(result: {
  detectionResults: {
    totalMissing: number;
    totalCutoff: number;
    successCount: number;
    failureCount: number;
  };
  searchResults: {
    missingTriggered: number;
    cutoffTriggered: number;
    successCount: number;
    failureCount: number;
  };
  totalSearches: number;
  totalFailures: number;
}): string {
  const lines: string[] = [
    header("Automation Cycle Complete"),
    "",
  ];

  lines.push(chalk.bold("Detection:"));
  lines.push(
    keyValue("  Missing items", chalk.yellow(result.detectionResults.totalMissing.toString()))
  );
  lines.push(
    keyValue("  Cutoff items", chalk.yellow(result.detectionResults.totalCutoff.toString()))
  );
  lines.push(
    keyValue("  Servers checked", chalk.cyan(result.detectionResults.successCount.toString()))
  );
  if (result.detectionResults.failureCount > 0) {
    lines.push(
      keyValue("  Failures", chalk.red(result.detectionResults.failureCount.toString()))
    );
  }
  lines.push("");

  lines.push(chalk.bold("Searches:"));
  lines.push(
    keyValue("  Missing searches", chalk.cyan(result.searchResults.missingTriggered.toString()))
  );
  lines.push(
    keyValue("  Cutoff searches", chalk.cyan(result.searchResults.cutoffTriggered.toString()))
  );
  lines.push(
    keyValue("  Total triggered", chalk.green(result.totalSearches.toString()))
  );
  if (result.totalFailures > 0) {
    lines.push(
      keyValue("  Failures", chalk.red(result.totalFailures.toString()))
    );
  }

  return lines.join("\n");
}

/**
 * Format detection-only results
 */
export function formatDetectionSummary(results: {
  serverId: string;
  serverName: string;
  serverType: string;
  missingCount: number;
  cutoffCount: number;
  error?: string;
}[]): string {
  const lines: string[] = [
    header("Detection Scan Results"),
    "",
  ];

  let totalMissing = 0;
  let totalCutoff = 0;

  for (const result of results) {
    const serverType = result.serverType === "radarr" ? chalk.magenta("Radarr") : chalk.cyan("Sonarr");

    if (result.error) {
      lines.push(`${serverType} - ${chalk.white(result.serverName)}: ${chalk.red("Failed")}`);
      lines.push(chalk.gray(`  Error: ${result.error}`));
    } else {
      lines.push(`${serverType} - ${chalk.white(result.serverName)}:`);
      lines.push(
        chalk.gray(`  Missing: ${chalk.yellow(result.missingCount.toString())} | Cutoff: ${chalk.yellow(result.cutoffCount.toString())}`)
      );
      totalMissing += result.missingCount;
      totalCutoff += result.cutoffCount;
    }
  }

  lines.push("");
  lines.push(chalk.bold("Total:"));
  lines.push(
    keyValue("  Missing items", chalk.yellow(totalMissing.toString()))
  );
  lines.push(
    keyValue("  Cutoff items", chalk.yellow(totalCutoff.toString()))
  );

  return lines.join("\n");
}

/**
 * Show a spinner/progress indicator
 */
export function showProgress(message: string): void {
  process.stdout.write(chalk.gray(`${message}...`));
}

/**
 * Clear the current line (for progress indicators)
 */
export function clearLine(): void {
  process.stdout.write("\r\x1b[K");
}

/**
 * Format a prompt for user input
 */
export function prompt(message: string): string {
  return chalk.cyan(`? ${message}`);
}
