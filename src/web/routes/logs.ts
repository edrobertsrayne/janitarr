/**
 * Logs API routes
 */

import type { DatabaseManager } from "../../storage/database";
import { jsonSuccess, jsonError, HttpStatus } from "../types";
import type { LogQueryParams, LogDeleteResponse } from "../types";
import type { LogEntryType } from "../../types";

/**
 * Handle GET /api/logs
 */
export async function handleGetLogs(url: URL, db: DatabaseManager): Promise<Response> {
  try {
    const params: LogQueryParams = {
      limit: parseInt(url.searchParams.get("limit") || "100", 10),
      offset: parseInt(url.searchParams.get("offset") || "0", 10),
      type: url.searchParams.get("type") as LogEntryType | undefined,
      server: url.searchParams.get("server") || undefined,
      startDate: url.searchParams.get("startDate") || undefined,
      endDate: url.searchParams.get("endDate") || undefined,
      search: url.searchParams.get("search") || undefined,
    };

    // Validate limit
    if (params.limit && (params.limit < 1 || params.limit > 1000)) {
      return jsonError("Limit must be between 1 and 1000", HttpStatus.BAD_REQUEST);
    }

    // Validate offset
    if (params.offset && params.offset < 0) {
      return jsonError("Offset must be non-negative", HttpStatus.BAD_REQUEST);
    }

    const logs = db.getLogsPaginated(
      {
        type: params.type,
        server: params.server,
        startDate: params.startDate,
        endDate: params.endDate,
        search: params.search,
      },
      params.limit || 100,
      params.offset || 0
    );

    return jsonSuccess(logs);
  } catch (error) {
    return jsonError(
      `Failed to retrieve logs: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle DELETE /api/logs
 */
export async function handleDeleteLogs(db: DatabaseManager): Promise<Response> {
  try {
    const deletedCount = db.clearLogs();
    const response: LogDeleteResponse = { deletedCount };
    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to delete logs: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle GET /api/logs/export
 */
export async function handleExportLogs(url: URL, db: DatabaseManager): Promise<Response> {
  try {
    const format = url.searchParams.get("format") || "json";

    const params: LogQueryParams = {
      limit: parseInt(url.searchParams.get("limit") || "1000", 10),
      offset: parseInt(url.searchParams.get("offset") || "0", 10),
      type: url.searchParams.get("type") as LogEntryType | undefined,
      server: url.searchParams.get("server") || undefined,
      startDate: url.searchParams.get("startDate") || undefined,
      endDate: url.searchParams.get("endDate") || undefined,
      search: url.searchParams.get("search") || undefined,
    };

    const logs = db.getLogsPaginated(
      {
        type: params.type,
        server: params.server,
        startDate: params.startDate,
        endDate: params.endDate,
        search: params.search,
      },
      params.limit || 1000,
      params.offset || 0
    );

    if (format === "csv") {
      // Convert to CSV format
      const headers = ["Timestamp", "Type", "Server", "Category", "Count", "Message", "Is Manual"];
      const rows = logs.map(log => [
        log.timestamp.toISOString(),
        log.type,
        log.serverName || "",
        log.category || "",
        log.count?.toString() || "",
        log.message,
        log.isManual ? "Yes" : "No",
      ]);

      const csv = [
        headers.join(","),
        ...rows.map(row => row.map(field => `"${field.replace(/"/g, '""')}"`).join(",")),
      ].join("\n");

      return new Response(csv, {
        status: HttpStatus.OK,
        headers: {
          "Content-Type": "text/csv",
          "Content-Disposition": `attachment; filename="janitarr-logs-${new Date().toISOString().split("T")[0]}.csv"`,
        },
      });
    }

    // Default to JSON format
    return new Response(JSON.stringify(logs, null, 2), {
      status: HttpStatus.OK,
      headers: {
        "Content-Type": "application/json",
        "Content-Disposition": `attachment; filename="janitarr-logs-${new Date().toISOString().split("T")[0]}.json"`,
      },
    });
  } catch (error) {
    return jsonError(
      `Failed to export logs: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}
