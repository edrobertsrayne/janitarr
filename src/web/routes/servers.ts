import { jsonError, jsonSuccess, HttpStatus, CreateServerRequest, ServerTestResponse, UpdateServerRequest } from "../types";
import { ServerConfig } from "../../types";
import { DatabaseManager } from "../../storage/database";
import {
  addServer,
  updateServer,
  removeServer,
  getServerById,
  getServers,
  testServerConnection as testExistingServerConnection, // Renamed to avoid conflict
  testConnection, // For testing new/updated credentials
} from "../../services/server-manager";

// Helper to parse JSON body
async function parseJsonBody<T>(req: Request): Promise<T | undefined> {
  try {
    return (await req.json()) as T;
  } catch {
    return undefined;
  }
}

/**
 * Handle POST /api/servers/test (for new, unsaved servers)
 */
export async function handleTestNewServer(req: Request): Promise<Response> {
  try {
    const body = await parseJsonBody<CreateServerRequest>(req);
    if (!body) {
      return jsonError("Invalid JSON body", HttpStatus.BAD_REQUEST);
    }

    if (!body.name || !body.type || !body.url || !body.apiKey) {
      return jsonError("Missing required fields: name, type, url, apiKey", HttpStatus.BAD_REQUEST);
    }

    // Normalize URL
    let normalizedUrl = body.url.trim();
    if (!normalizedUrl.startsWith("http://") && !normalizedUrl.startsWith("https://")) {
      normalizedUrl = `http://${normalizedUrl}`;
    }
    normalizedUrl = normalizedUrl.replace(/\/$/, ""); // Remove trailing slash

    // Test connection using API client
    const testResult = await testConnection(normalizedUrl, body.apiKey, body.type);

    if (!testResult.success) {
      const response: ServerTestResponse = {
        success: false,
        message: testResult.error ?? "Failed to connect to server",
      };
      return jsonSuccess(response);
    }

    const response: ServerTestResponse = {
      success: true,
      message: "Connection successful",
      status: testResult.data,
    };
    return jsonSuccess(response);
  } catch (error) {
    return jsonError(
      `Failed to test server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Placeholder for GET /api/servers
 */
export async function handleGetServers(url: URL, db: DatabaseManager): Promise<Response> {
  return jsonError("Not Implemented", HttpStatus.NOT_IMPLEMENTED);
}

/**
 * Placeholder for GET /api/servers/:id
 */
export async function handleGetServer(path: string, db: DatabaseManager): Promise<Response> {
  return jsonError("Not Implemented", HttpStatus.NOT_IMPLEMENTED);
}

/**
 * Placeholder for POST /api/servers
 */
export async function handleCreateServer(req: Request, db: DatabaseManager): Promise<Response> {
  return jsonError("Not Implemented", HttpStatus.NOT_IMPLEMENTED);
}

/**
 * Placeholder for DELETE /api/servers/:id
 */
export async function handleDeleteServer(path: string, db: DatabaseManager): Promise<Response> {
  return jsonError("Not Implemented", HttpStatus.NOT_IMPLEMENTED);
}

/**
 * Placeholder for POST /api/servers/:id/test
 */
export async function handleTestServer(path: string, db: DatabaseManager): Promise<Response> {
  return jsonError("Not Implemented", HttpStatus.NOT_IMPLEMENTED);
}

/**
 * Handle PUT /api/servers/:id
 */
export async function handleUpdateServer(req: Request, path: string, db: DatabaseManager): Promise<Response> {
  try {
    const serverId = path.split('/').pop();
    if (!serverId) {
      return jsonError("Server ID not found in path", HttpStatus.BAD_REQUEST);
    }

    const body = await parseJsonBody<UpdateServerRequest>(req);
    if (!body) {
      return jsonError("Invalid JSON body", HttpStatus.BAD_REQUEST);
    }

    // Retrieve existing server to get its properties if not provided in update request
    const existingServerResult = await getServerById(db, serverId);
    if (!existingServerResult.success || !existingServerResult.data) {
      return jsonError("Server not found", HttpStatus.NOT_FOUND);
    }
    const existingServer = existingServerResult.data;

    // Use existing server's properties if not provided in update request
    const newName = body.name ?? existingServer.name;
    const newUrl = body.url ?? existingServer.url;
    const newApiKey = body.apiKey ?? existingServer.apiKey;

    // Validate if name uniqueness is maintained (if name is changed)
    if (newName !== existingServer.name) {
      const allServersResult = await getServers(db);
      if (allServersResult.success && allServersResult.data) {
        const nameExists = allServersResult.data.some(s => s.name === newName && s.id !== serverId);
        if (nameExists) {
          return jsonError("Server name already exists", HttpStatus.BAD_REQUEST);
        }
      }
    }

    // Test connection if URL or API key changed
    if (newUrl !== existingServer.url || newApiKey !== existingServer.apiKey) {
      const testResult = await testConnection(newUrl, newApiKey, existingServer.type);
      if (!testResult.success) {
        return jsonError(testResult.error || "Connection test failed with new credentials", HttpStatus.BAD_REQUEST);
      }
    }

    // Update the server in the database
    const updateResult = await updateServer(db, serverId, {
      name: newName,
      url: newUrl,
      apiKey: newApiKey,
    });

    if (!updateResult.success) {
      return jsonError(updateResult.error || "Failed to update server", HttpStatus.INTERNAL_SERVER_ERROR);
    }

    return jsonSuccess({ message: "Server updated successfully", server: updateResult.data });
  } catch (error) {
    console.error("Error updating server:", error);
    return jsonError(
      `Failed to update server: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}