/**
 * Server management API routes
 */

import type { DatabaseManager } from "../../storage/database";
import { jsonSuccess, jsonError, parseJsonBody, extractPathParam, HttpStatus } from "../types";
import type { CreateServerRequest, UpdateServerRequest, ServerTestResponse } from "../types";
import type { ServerType } from "../../types";

/**
 * Handle GET /api/servers
 */
export async function handleGetServers(url: URL, db: DatabaseManager): Promise<Response> {
  try {
    const typeFilter = url.searchParams.get("type") as ServerType | null;

    let servers;
    if (typeFilter && (typeFilter === "radarr" || typeFilter === "sonarr")) {
      servers = await db.getServersByType(typeFilter);
    } else {
      servers = await db.getAllServers();
    }

    return jsonSuccess(servers);
  } catch (error) {
    return jsonError(
      `Failed to retrieve servers: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle GET /api/servers/:id
 */
export async function handleGetServer(path: string, db: DatabaseManager): Promise<Response> {
  try {
    const serverId = extractPathParam(path, /^\/api\/servers\/([^/]+)$/);
    if (!serverId) {
      return jsonError("Invalid server ID", HttpStatus.BAD_REQUEST);
    }

    const server = await db.getServer(serverId);
    if (!server) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }

    return jsonSuccess(server);
  } catch (error) {
    return jsonError(
      `Failed to retrieve server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle POST /api/servers
 */
export async function handleCreateServer(req: Request, db: DatabaseManager): Promise<Response> {
  try {
    const body = await parseJsonBody<CreateServerRequest>(req);
    if (!body) {
      return jsonError("Invalid JSON body", HttpStatus.BAD_REQUEST);
    }

    // Validate required fields
    if (!body.name || !body.type || !body.url || !body.apiKey) {
      return jsonError("Missing required fields: name, type, url, apiKey", HttpStatus.BAD_REQUEST);
    }

    // Validate type
    if (body.type !== "radarr" && body.type !== "sonarr") {
      return jsonError("Invalid server type. Must be 'radarr' or 'sonarr'", HttpStatus.BAD_REQUEST);
    }

    // Normalize URL
    let normalizedUrl = body.url.trim();
    if (!normalizedUrl.startsWith("http://") && !normalizedUrl.startsWith("https://")) {
      normalizedUrl = `http://${normalizedUrl}`;
    }
    normalizedUrl = normalizedUrl.replace(/\/$/, ""); // Remove trailing slash

    // Check for duplicate server
    if (db.serverExists(normalizedUrl, body.type)) {
      return jsonError("A server with this URL and type already exists", HttpStatus.BAD_REQUEST);
    }

    // Create server
    const newServer = await db.addServer({
      id: crypto.randomUUID(),
      name: body.name,
      type: body.type,
      url: normalizedUrl,
      apiKey: body.apiKey,
    });

    return jsonSuccess(newServer, HttpStatus.CREATED);
  } catch (error) {
    return jsonError(
      `Failed to create server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle PUT /api/servers/:id
 */
export async function handleUpdateServer(req: Request, path: string, db: DatabaseManager): Promise<Response> {
  try {
    const serverId = extractPathParam(path, /^\/api\/servers\/([^/]+)$/);
    if (!serverId) {
      return jsonError("Invalid server ID", HttpStatus.BAD_REQUEST);
    }

    const body = await parseJsonBody<UpdateServerRequest>(req);
    if (!body) {
      return jsonError("Invalid JSON body", HttpStatus.BAD_REQUEST);
    }

    // Check server exists
    const existing = await db.getServer(serverId);
    if (!existing) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }

    // Normalize URL if provided
    if (body.url) {
      let normalizedUrl = body.url.trim();
      if (!normalizedUrl.startsWith("http://") && !normalizedUrl.startsWith("https://")) {
        normalizedUrl = `http://${normalizedUrl}`;
      }
      normalizedUrl = normalizedUrl.replace(/\/$/, "");
      body.url = normalizedUrl;

      // Check for duplicate URL (excluding this server)
      if (db.serverExists(normalizedUrl, existing.type, serverId)) {
        return jsonError("Another server with this URL and type already exists", HttpStatus.BAD_REQUEST);
      }
    }

    // Update server
    const updated = await db.updateServer(serverId, body);
    if (!updated) {
      return jsonError("Failed to update server", HttpStatus.INTERNAL_SERVER_ERROR);
    }

    return jsonSuccess(updated);
  } catch (error) {
    return jsonError(
      `Failed to update server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle DELETE /api/servers/:id
 */
export async function handleDeleteServer(path: string, db: DatabaseManager): Promise<Response> {
  try {
    const serverId = extractPathParam(path, /^\/api\/servers\/([^/]+)$/);
    if (!serverId) {
      return jsonError("Invalid server ID", HttpStatus.BAD_REQUEST);
    }

    const deleted = db.deleteServer(serverId);
    if (!deleted) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }

    return new Response(null, { status: HttpStatus.NO_CONTENT });
  } catch (error) {
    return jsonError(
      `Failed to delete server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle POST /api/servers/:id/test
 */
export async function handleTestServer(path: string, db: DatabaseManager): Promise<Response> {
  try {
    const serverId = extractPathParam(path, /^\/api\/servers\/([^/]+)\/test$/);
    if (!serverId) {
      return jsonError("Invalid server ID", HttpStatus.BAD_REQUEST);
    }

    const server = await db.getServer(serverId);
    if (!server) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }

    // Test connection using API client
    const { createClient } = await import("../../lib/api-client");
    const client = createClient(server.url, server.apiKey, server.type);
    const status = await client.testConnection();

    if (!status.success) {
      const response: ServerTestResponse = {
        success: false,
        message: status.error ?? "Failed to connect to server",
      };
      return jsonSuccess(response);
    }

    const response: ServerTestResponse = {
      success: true,
      message: "Connection successful",
      status: status.data,
    };
    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to test server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}
