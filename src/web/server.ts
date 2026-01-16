/**
 * Web server for Janitarr
 *
 * Provides REST API and WebSocket endpoints for the web frontend.
 */

import type { DatabaseManager } from "../storage/database";
import { websocketHandlers } from "./websocket";
import { jsonError, HttpStatus } from "./types";

// Import route handlers
import { handleGetConfig, handlePatchConfig, handleResetConfig } from "./routes/config";
import {
  handleGetServers,
  handleGetServer,
  handleCreateServer,
  handleUpdateServer,
  handleDeleteServer,
  handleTestServer,
} from "./routes/servers";
import { handleGetLogs, handleDeleteLogs, handleExportLogs } from "./routes/logs";
import { handleTriggerAutomation, handleGetAutomationStatus } from "./routes/automation";
import { handleGetStatsSummary, handleGetServerStats } from "./routes/stats";

/** Web server options */
export interface WebServerOptions {
  port?: number;
  host?: string;
  db: DatabaseManager;
}

/**
 * Create and start the web server
 */
export function createWebServer(options: WebServerOptions) {
  const { port = 3000, host = "localhost", db } = options;

  const server = Bun.serve({
    port,
    hostname: host,

    /**
     * HTTP request handler
     */
    async fetch(req, server) {
      const url = new URL(req.url);
      const path = url.pathname;
      const method = req.method;

      // Handle WebSocket upgrade
      if (path === "/ws/logs") {
        const upgraded = server.upgrade(req, { data: {} });
        if (!upgraded) {
          return jsonError("WebSocket upgrade failed", HttpStatus.INTERNAL_SERVER_ERROR);
        }
        return undefined; // WebSocket connection established
      }

      // Add CORS headers
      const headers = new Headers({
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type, Authorization",
      });

      // Handle preflight OPTIONS requests
      if (method === "OPTIONS") {
        return new Response(null, { status: HttpStatus.NO_CONTENT, headers });
      }

      try {
        let response: Response;

        // Route handling
        if (path === "/api/config" && method === "GET") {
          response = await handleGetConfig(db);
        } else if (path === "/api/config" && method === "PATCH") {
          response = await handlePatchConfig(req, db);
        } else if (path === "/api/config/reset" && method === "PUT") {
          response = await handleResetConfig(db);
        } else if (path === "/api/servers" && method === "GET") {
          response = await handleGetServers(url, db);
        } else if (path.match(/^\/api\/servers\/[^/]+$/) && method === "GET") {
          response = await handleGetServer(path, db);
        } else if (path === "/api/servers" && method === "POST") {
          response = await handleCreateServer(req, db);
        } else if (path.match(/^\/api\/servers\/[^/]+$/) && method === "PUT") {
          response = await handleUpdateServer(req, path, db);
        } else if (path.match(/^\/api\/servers\/[^/]+$/) && method === "DELETE") {
          response = await handleDeleteServer(path, db);
        } else if (path.match(/^\/api\/servers\/[^/]+\/test$/) && method === "POST") {
          response = await handleTestServer(path, db);
        } else if (path === "/api/logs" && method === "GET") {
          response = await handleGetLogs(url, db);
        } else if (path === "/api/logs" && method === "DELETE") {
          response = await handleDeleteLogs(db);
        } else if (path === "/api/logs/export" && method === "GET") {
          response = await handleExportLogs(url, db);
        } else if (path === "/api/automation/trigger" && method === "POST") {
          response = await handleTriggerAutomation(req);
        } else if (path === "/api/automation/status" && method === "GET") {
          response = await handleGetAutomationStatus(db);
        } else if (path === "/api/stats/summary" && method === "GET") {
          response = await handleGetStatsSummary(db);
        } else if (path.match(/^\/api\/stats\/servers\/[^/]+$/) && method === "GET") {
          response = await handleGetServerStats(path, db);
        } else if (path === "/api/health" && method === "GET") {
          // Health check endpoint
          response = new Response(JSON.stringify({ status: "ok", timestamp: new Date().toISOString() }), {
            status: HttpStatus.OK,
            headers: { "Content-Type": "application/json" },
          });
        } else {
          // 404 for unknown routes
          response = jsonError("Not found", HttpStatus.NOT_FOUND);
        }

        // Add CORS headers to response
        for (const [key, value] of headers.entries()) {
          response.headers.set(key, value);
        }

        return response;
      } catch (error) {
        console.error("Request handler error:", error);
        const errorResponse = jsonError(
          "Internal server error",
          HttpStatus.INTERNAL_SERVER_ERROR
        );

        // Add CORS headers to error response
        for (const [key, value] of headers.entries()) {
          errorResponse.headers.set(key, value);
        }

        return errorResponse;
      }
    },

    /**
     * WebSocket handlers
     */
    websocket: websocketHandlers,

    /**
     * Error handler
     */
    error(error) {
      console.error("Server error:", error);
      return jsonError("Internal server error", HttpStatus.INTERNAL_SERVER_ERROR);
    },
  });

  console.log(`Web server listening on http://${host}:${port}`);
  console.log(`WebSocket endpoint: ws://${host}:${port}/ws/logs`);
  console.log(`API base URL: http://${host}:${port}/api`);

  return server;
}

/**
 * Stop the web server
 */
export function stopWebServer(server: ReturnType<typeof Bun.serve>): void {
  server.stop();
  console.log("Web server stopped");
}
