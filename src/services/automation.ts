/**
 * Automation Orchestrator
 *
 * Coordinates detection, search triggering, and logging for complete
 * automation cycles. Handles partial failures gracefully.
 */

import { detectAll } from "./detector";
import { triggerSearches } from "./search-trigger";
import {
  logCycleStart,
  logCycleEnd,
  logSearches,
  logServerError,
  logSearchError,
} from "../lib/logger";

/** Result of a complete automation cycle */
export interface CycleResult {
  success: boolean;
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
  errors: string[];
}

/**
 * Execute a complete automation cycle
 *
 * @param isManual - Whether this is a manual trigger (vs scheduled)
 * @param dryRun - If true, preview what would be searched without triggering actual searches
 * @returns Cycle result summary
 */
export async function runAutomationCycle(
  isManual = false,
  dryRun = false
): Promise<CycleResult> {
  const errors: string[] = [];

  // Log cycle start (only if not dry-run)
  if (!dryRun) {
    logCycleStart(isManual);
  }

  // Phase 1: Detection
  const detectionResults = await detectAll();

  // Log detection failures (only if not dry-run)
  for (const result of detectionResults.results) {
    if (result.error) {
      if (!dryRun) {
        logServerError(result.serverName, result.serverType, result.error);
      }
      errors.push(`Detection failed for ${result.serverName}: ${result.error}`);
    }
  }

  // Phase 2: Search Triggering (or preview in dry-run mode)
  let searchResults;
  try {
    searchResults = await triggerSearches(detectionResults, dryRun);
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : String(error);
    errors.push(`Search triggering failed: ${errorMsg}`);

    // Log cycle end with failure (only if not dry-run)
    if (!dryRun) {
      logCycleEnd(0, errors.length, isManual);
    }

    return {
      success: false,
      detectionResults: {
        totalMissing: detectionResults.totalMissing,
        totalCutoff: detectionResults.totalCutoff,
        successCount: detectionResults.successCount,
        failureCount: detectionResults.failureCount,
      },
      searchResults: {
        missingTriggered: 0,
        cutoffTriggered: 0,
        successCount: 0,
        failureCount: 0,
      },
      totalSearches: 0,
      totalFailures: errors.length,
      errors,
    };
  }

  // Log successful searches (only if not dry-run)
  if (!dryRun) {
    const serverSearchCounts = new Map<
      string,
      { missing: number; cutoff: number }
    >();

    for (const result of searchResults.results) {
      const key = `${result.serverName}:${result.serverType}`;

      if (!serverSearchCounts.has(key)) {
        serverSearchCounts.set(key, { missing: 0, cutoff: 0 });
      }

      const counts = serverSearchCounts.get(key)!;

      if (result.success) {
        if (result.category === "missing") {
          counts.missing += result.itemIds.length;
        } else {
          counts.cutoff += result.itemIds.length;
        }
      } else {
        // Log search failures
        logSearchError(
          result.serverName,
          result.serverType,
          result.category,
          result.error ?? "Unknown error"
        );
        errors.push(
          `Search trigger failed for ${result.serverName} (${result.category}): ${result.error}`
        );
      }
    }

    // Log aggregated successful searches per server
    for (const [key, counts] of serverSearchCounts) {
      const [serverName, serverType] = key.split(":");

      if (counts.missing > 0) {
        logSearches(
          serverName,
          serverType as "radarr" | "sonarr",
          "missing",
          counts.missing,
          isManual
        );
      }

      if (counts.cutoff > 0) {
        logSearches(
          serverName,
          serverType as "radarr" | "sonarr",
          "cutoff",
          counts.cutoff,
          isManual
        );
      }
    }
  }

  const totalSearches =
    searchResults.missingTriggered + searchResults.cutoffTriggered;
  const totalFailures =
    detectionResults.failureCount + searchResults.failureCount;

  // Log cycle end (only if not dry-run)
  if (!dryRun) {
    logCycleEnd(totalSearches, totalFailures, isManual);
  }

  return {
    success: totalFailures === 0,
    detectionResults: {
      totalMissing: detectionResults.totalMissing,
      totalCutoff: detectionResults.totalCutoff,
      successCount: detectionResults.successCount,
      failureCount: detectionResults.failureCount,
    },
    searchResults: {
      missingTriggered: searchResults.missingTriggered,
      cutoffTriggered: searchResults.cutoffTriggered,
      successCount: searchResults.successCount,
      failureCount: searchResults.failureCount,
    },
    totalSearches,
    totalFailures,
    errors,
  };
}

/**
 * Get a human-readable summary of cycle results
 */
export function formatCycleResult(result: CycleResult): string {
  const lines: string[] = [];

  lines.push("=== Automation Cycle Summary ===");
  lines.push("");

  // Detection summary
  lines.push("Detection:");
  lines.push(
    `  Found: ${result.detectionResults.totalMissing} missing, ${result.detectionResults.totalCutoff} cutoff`
  );
  lines.push(
    `  Servers: ${result.detectionResults.successCount} successful, ${result.detectionResults.failureCount} failed`
  );
  lines.push("");

  // Search summary
  lines.push("Search Triggering:");
  lines.push(
    `  Triggered: ${result.searchResults.missingTriggered} missing, ${result.searchResults.cutoffTriggered} cutoff`
  );
  lines.push(
    `  Total: ${result.totalSearches} searches, ${result.searchResults.failureCount} failures`
  );
  lines.push("");

  // Overall status
  if (result.success) {
    lines.push("Status: ✓ Success");
  } else {
    lines.push(`Status: ✗ ${result.totalFailures} failures`);
    if (result.errors.length > 0) {
      lines.push("");
      lines.push("Errors:");
      for (const error of result.errors) {
        lines.push(`  - ${error}`);
      }
    }
  }

  return lines.join("\n");
}
