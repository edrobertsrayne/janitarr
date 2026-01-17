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
  handleTestNewServer,
} from "./routes/servers";
import { handleGetLogs, handleDeleteLogs, handleExportLogs } from "./routes/logs";
import { handleTriggerAutomation, handleGetAutomationStatus } from "./routes/automation";
import { handleGetStatsSummary, handleGetServerStats } from "./routes/stats";
import { handleHealthCheck } from "./routes/health";
import { handleMetrics } from "./routes/metrics";
import { recordHttpRequest } from "../lib/metrics";

/** Web server options */
export interface WebServerOptions {
  port?: number;
  host?: string;
  db: DatabaseManager;
  /** Skip startup console output (default: false) */
  silent?: boolean;
}

/**
 * Serve static files from dist/public directory
 */
async function serveStaticFile(urlPath: string): Promise<Response> {
  // Default to index.html for root path and SPA routes
  const filePath = urlPath === "/" ? "/index.html" : urlPath;

  // Construct full file path
  const publicDir = new URL("../../dist/public", import.meta.url).pathname;
  const fullPath = publicDir + filePath;

  try {
    const file = Bun.file(fullPath);
    const exists = await file.exists();

    if (!exists) {
      // For SPA routing, serve index.html for non-existent routes
      const indexFile = Bun.file(publicDir + "/index.html");
      const indexExists = await indexFile.exists();

      if (indexExists) {
        return new Response(indexFile, {
          headers: { "Content-Type": "text/html" },
        });
      }

      return new Response("Not Found", { status: 404 });
    }

    // Determine content type based on file extension
    const ext = filePath.split(".").pop()?.toLowerCase();
    const contentTypes: Record<string, string> = {
      html: "text/html",
      css: "text/css",
      js: "application/javascript",
      json: "application/json",
      png: "image/png",
      jpg: "image/jpeg",
      jpeg: "image/jpeg",
      gif: "image/gif",
      svg: "image/svg+xml",
      ico: "image/x-icon",
      woff: "font/woff",
      woff2: "font/woff2",
      ttf: "font/ttf",
    };

    const contentType = contentTypes[ext || ""] || "application/octet-stream";

    return new Response(file, {
      headers: { "Content-Type": contentType },
    });
  } catch (error) {
    console.error("Error serving static file:", error);
    return new Response("Internal Server Error", { status: 500 });
  }
}

/**
 * Create and start the web server
 */
export function createWebServer(options: WebServerOptions) {
  const { port = 3000, host = "localhost", db, silent = false } = options;

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
      const startTime = performance.now();

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
        } else if (path === "/api/servers/test" && method === "POST") {
          response = await handleTestNewServer(req);
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
          response = await handleHealthCheck(db);
        } else if (path === "/metrics" && method === "GET") {
          response = handleMetrics();
        } else {
          // Serve static files from dist/public for non-API routes
          if (!path.startsWith("/api/") && !path.startsWith("/metrics")) {
            response = await serveStaticFile(path);
          } else {
            // 404 for unknown API routes
            response = jsonError("Not found", HttpStatus.NOT_FOUND);
          }
        }

        // Add CORS headers to response
        for (const [key, value] of headers.entries()) {
          response.headers.set(key, value);
        }

        // Track HTTP request metrics
        const endTime = performance.now();
        const durationMs = endTime - startTime;
        recordHttpRequest(method, path, response.status, durationMs);

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

        // Track HTTP request metrics for errors
        const endTime = performance.now();
        const durationMs = endTime - startTime;
        recordHttpRequest(method, path, errorResponse.status, durationMs);

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

  if (!silent) {
    console.log(`Web server listening on http://${host}:${port}`);
    console.log(`WebSocket endpoint: ws://${host}:${port}/ws/logs`);
    console.log(`API base URL: http://${host}:${port}/api`);
  }

  return server;
}

/**
 * Stop the web server
 */
export function stopWebServer(server: ReturnType<typeof Bun.serve>): void {
  server.stop();
  console.log("Web server stopped");
}
