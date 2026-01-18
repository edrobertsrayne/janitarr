/**
 * Statistics API routes
 */

import type { DatabaseManager } from "../../storage/database";
import { jsonSuccess, jsonError, extractPathParam, HttpStatus } from "../types";
import type { StatsSummaryResponse, ServerStatsResponse } from "../types";

/**
 * Handle GET /api/stats/summary
 */
export async function handleGetStatsSummary(
  db: DatabaseManager,
): Promise<Response> {
  try {
    const systemStats = db.getSystemStats();

    // Calculate next scheduled time
    const { getStatus } = await import("../../lib/scheduler");
    const schedulerStatus = getStatus();

    // Count active servers (all servers for now since we don't have enabled/disabled flag in DB yet)
    const servers = await db.getAllServers();
    const totalServers = servers.length;
    const activeServers = servers.length; // All servers are considered active for now

    const response: StatsSummaryResponse = {
      totalServers,
      activeServers,
      lastCycleTime: systemStats.lastCycleTime,
      nextScheduledTime: schedulerStatus.nextRunTime?.toISOString() || null,
      searchesLast24h: systemStats.searchesLast24h,
      errorsLast24h: systemStats.errorsLast24h,
    };

    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to retrieve summary statistics: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR,
    );
  }
}

/**
 * Handle GET /api/stats/servers/:id
 */
export async function handleGetServerStats(
  path: string,
  db: DatabaseManager,
): Promise<Response> {
  try {
    const serverId = extractPathParam(path, /^\/api\/stats\/servers\/([^/]+)$/);
    if (!serverId) {
      return jsonError("Invalid server ID", HttpStatus.BAD_REQUEST);
    }

    // Check server exists
    const server = await db.getServer(serverId);
    if (!server) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }

    const stats = db.getServerStats(serverId);

    // Calculate success rate
    const totalAttempts = stats.totalSearches + stats.errorCount;
    const successRate =
      totalAttempts > 0 ? (stats.totalSearches / totalAttempts) * 100 : 100;

    const response: ServerStatsResponse = {
      totalSearches: stats.totalSearches,
      successRate: Math.round(successRate * 100) / 100, // Round to 2 decimal places
      lastCheckTime: stats.lastCheckTime,
      errorCount: stats.errorCount,
    };

    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to retrieve server statistics: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR,
    );
  }
}
