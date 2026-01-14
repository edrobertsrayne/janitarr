/**
 * Activity Logger
 *
 * High-level logging API for automation activities. Logs are persisted
 * to SQLite and automatically purged after 30 days.
 */

import type { ServerType, LogEntry, LogEntryType, SearchCategory } from "../types";
import { getDatabase } from "../storage/database";

/** Log a cycle start event */
export function logCycleStart(isManual = false): LogEntry {
  const db = getDatabase();
  return db.addLog({
    type: "cycle_start",
    message: isManual ? "Manual automation cycle started" : "Scheduled automation cycle started",
    isManual,
  });
}

/** Log a cycle end event with summary */
export function logCycleEnd(
  totalSearches: number,
  failures: number,
  isManual = false
): LogEntry {
  const db = getDatabase();
  const message =
    failures > 0
      ? `Automation cycle complete: ${totalSearches} searches triggered, ${failures} failures`
      : `Automation cycle complete: ${totalSearches} searches triggered`;

  return db.addLog({
    type: "cycle_end",
    message,
    count: totalSearches,
    isManual,
  });
}

/** Log triggered searches for a server */
export function logSearches(
  serverName: string,
  serverType: ServerType,
  category: SearchCategory,
  count: number,
  isManual = false
): LogEntry {
  const db = getDatabase();
  const categoryLabel = category === "missing" ? "missing" : "cutoff";

  return db.addLog({
    type: "search",
    serverName,
    serverType,
    category,
    count,
    message: `Triggered ${count} ${categoryLabel} searches on ${serverName}`,
    isManual,
  });
}

/** Log a server connection error */
export function logServerError(
  serverName: string,
  serverType: ServerType,
  reason: string
): LogEntry {
  const db = getDatabase();
  return db.addLog({
    type: "error",
    serverName,
    serverType,
    message: `Connection failed to ${serverName}: ${reason}`,
  });
}

/** Log a search trigger error */
export function logSearchError(
  serverName: string,
  serverType: ServerType,
  category: SearchCategory,
  reason: string
): LogEntry {
  const db = getDatabase();
  const categoryLabel = category === "missing" ? "missing" : "cutoff";

  return db.addLog({
    type: "error",
    serverName,
    serverType,
    category,
    message: `Failed to trigger ${categoryLabel} searches on ${serverName}: ${reason}`,
  });
}

/** Get recent log entries */
export function getRecentLogs(limit = 100, offset = 0): LogEntry[] {
  const db = getDatabase();
  return db.getLogs(limit, offset);
}

/** Get total log count */
export function getLogCount(): number {
  const db = getDatabase();
  return db.getLogCount();
}

/** Clear all logs (requires confirmation in UI) */
export function clearAllLogs(): number {
  const db = getDatabase();
  return db.clearLogs();
}

/** Purge logs older than 30 days */
export function purgeOldLogs(): number {
  const db = getDatabase();
  return db.purgeOldLogs();
}

/** Get last cycle summary */
export interface CycleSummary {
  lastCycleTime: Date | null;
  lastCycleSearches: number;
  lastCycleFailures: number;
  wasManual: boolean;
}

export function getLastCycleSummary(): CycleSummary {
  const db = getDatabase();
  const logs = db.getLogs(50); // Check recent logs

  // Find the most recent cycle_end
  const cycleEnd = logs.find((l) => l.type === "cycle_end");

  if (!cycleEnd) {
    return {
      lastCycleTime: null,
      lastCycleSearches: 0,
      lastCycleFailures: 0,
      wasManual: false,
    };
  }

  // Count failures between this cycle_end and the preceding cycle_start
  let failures = 0;

  for (const log of logs) {
    if (log.id === cycleEnd.id) continue;

    if (log.type === "cycle_start") {
      break;
    }

    if (log.type === "error") {
      failures++;
    }
  }

  return {
    lastCycleTime: cycleEnd.timestamp,
    lastCycleSearches: cycleEnd.count ?? 0,
    lastCycleFailures: failures,
    wasManual: cycleEnd.isManual ?? false,
  };
}

/**
 * Format a log entry for display
 */
export function formatLogEntry(entry: LogEntry): string {
  const time = entry.timestamp.toLocaleString();
  const prefix = entry.isManual ? "[Manual]" : "";

  switch (entry.type) {
    case "cycle_start":
      return `${time} ${prefix} Cycle started`;
    case "cycle_end":
      return `${time} ${prefix} ${entry.message}`;
    case "search":
      return `${time} ${prefix} ${entry.message}`;
    case "error":
      return `${time} [ERROR] ${entry.message}`;
    default:
      return `${time} ${entry.message}`;
  }
}

/**
 * Get log entry type display label
 */
export function getLogTypeLabel(type: LogEntryType): string {
  switch (type) {
    case "cycle_start":
      return "Cycle Start";
    case "cycle_end":
      return "Cycle End";
    case "search":
      return "Search";
    case "error":
      return "Error";
    default:
      return "Unknown";
  }
}
