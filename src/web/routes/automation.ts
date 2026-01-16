/**
 * Automation control API routes
 */

import { jsonSuccess, jsonError, parseJsonBody, HttpStatus } from "../types";
import type { TriggerAutomationRequest, TriggerAutomationResponse, AutomationStatusResponse } from "../types";
import type { DatabaseManager } from "../../storage/database";

/**
 * Handle POST /api/automation/trigger
 */
export async function handleTriggerAutomation(req: Request): Promise<Response> {
  try {
    const body = await parseJsonBody<TriggerAutomationRequest>(req);
    const type = body?.type || "full";

    // Validate type
    if (type !== "full" && type !== "missing" && type !== "cutoff") {
      return jsonError("Invalid automation type. Must be 'full', 'missing', or 'cutoff'", HttpStatus.BAD_REQUEST);
    }

    // Import and run automation cycle
    const { runAutomationCycle } = await import("../../services/automation");

    // Generate a job ID
    const jobId = crypto.randomUUID();

    // Run cycle asynchronously (don't await - return immediately)
    runAutomationCycle(true).catch(error => {
      console.error("Automation cycle failed:", error);
    });

    const response: TriggerAutomationResponse = {
      jobId,
      message: "Automation cycle started",
    };

    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to trigger automation: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle GET /api/automation/status
 */
export async function handleGetAutomationStatus(db: DatabaseManager): Promise<Response> {
  try {
    // Get scheduler status
    const { getStatus } = await import("../../lib/scheduler");
    const schedulerStatus = getStatus();

    // Get last cycle end log to extract results
    const lastCycleLogs = db.getLogsPaginated({ type: "cycle_end" }, 1, 0);
    const lastCycleLog = lastCycleLogs[0];

    let lastRunResults: { searchesTriggered: number; failedServers: number } | undefined;
    if (lastCycleLog) {
      // Parse results from message (format: "Cycle complete: X searches triggered across Y servers (Z failures)")
      const match = lastCycleLog.message.match(/(\d+) searches triggered.*\((\d+) failures\)/);
      if (match) {
        lastRunResults = {
          searchesTriggered: parseInt(match[1], 10),
          failedServers: parseInt(match[2], 10),
        };
      }
    }

    const response: AutomationStatusResponse = {
      running: schedulerStatus.isCycleActive,
      nextScheduledRun: schedulerStatus.nextRunTime?.toISOString() || null,
      lastRunTime: lastCycleLog?.timestamp.toISOString() || null,
      lastRunResults,
    };

    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to get automation status: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}
