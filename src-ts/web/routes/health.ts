/**
 * Health check endpoint
 *
 * Provides comprehensive status of all services for monitoring and deployment systems.
 */

import type { DatabaseManager } from "../../storage/database";
import { getStatus as getSchedulerStatus } from "../../lib/scheduler";
import { HttpStatus } from "../types";

/** Health check response format */
export interface HealthResponse {
  status: "ok" | "degraded" | "error";
  timestamp: string;
  services: {
    webServer: { status: "ok" };
    scheduler: {
      status: "ok" | "disabled" | "error";
      isRunning: boolean;
      isCycleActive: boolean;
      nextRun: string | null;
    };
  };
  database: { status: "ok" | "error" };
}

/**
 * Handle GET /api/health request
 *
 * Returns comprehensive health status of all services.
 * - Overall status: "ok" when all services healthy
 * - Overall status: "degraded" when scheduler disabled but web server running
 * - Overall status: "error" when critical component failing
 * - HTTP 200 for "ok" and "degraded"
 * - HTTP 503 for "error"
 */
export async function handleHealthCheck(
  db: DatabaseManager,
): Promise<Response> {
  const timestamp = new Date().toISOString();

  // Check database connectivity with lightweight query
  let dbStatus: "ok" | "error" = "ok";
  try {
    // Simple query to verify database connection
    db.getAppConfig();
  } catch (error) {
    console.error("Database health check failed:", error);
    dbStatus = "error";
  }

  // Get scheduler status
  const schedulerState = getSchedulerStatus();
  // Get scheduler config from the provided database instance
  const appConfig = db.getAppConfig();
  const scheduleConfig = appConfig.schedule;

  // Determine scheduler status
  let schedulerStatus: "ok" | "disabled" | "error";
  if (!scheduleConfig.enabled) {
    schedulerStatus = "disabled";
  } else if (schedulerState.isRunning) {
    schedulerStatus = "ok";
  } else {
    // Scheduler enabled but not running is an error state
    schedulerStatus = "error";
  }

  // Determine overall health status
  let overallStatus: "ok" | "degraded" | "error";
  if (dbStatus === "error" || schedulerStatus === "error") {
    overallStatus = "error";
  } else if (schedulerStatus === "disabled") {
    overallStatus = "degraded";
  } else {
    overallStatus = "ok";
  }

  // Build response
  const response: HealthResponse = {
    status: overallStatus,
    timestamp,
    services: {
      webServer: { status: "ok" },
      scheduler: {
        status: schedulerStatus,
        isRunning: schedulerState.isRunning,
        isCycleActive: schedulerState.isCycleActive,
        nextRun: schedulerState.nextRunTime
          ? schedulerState.nextRunTime.toISOString()
          : null,
      },
    },
    database: { status: dbStatus },
  };

  // Return appropriate HTTP status code
  const httpStatus =
    overallStatus === "error" ? HttpStatus.SERVICE_UNAVAILABLE : HttpStatus.OK;

  return new Response(JSON.stringify(response), {
    status: httpStatus,
    headers: { "Content-Type": "application/json" },
  });
}
