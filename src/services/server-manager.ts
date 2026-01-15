/**
 * Server Manager Service
 *
 * Handles CRUD operations for Radarr/Sonarr server configurations with
 * connection validation and duplicate prevention.
 */

import type { ServerConfig, ServerType } from "../types";
import {
  type ApiResult,
  type SystemStatus,
  createClient,
  validateUrl,
} from "../lib/api-client";
import { getDatabase } from "../storage/database";

/** Result of a server operation */
export type ServerResult<T> =
  | { success: true; data: T }
  | { success: false; error: string };

/** Server info for display (with masked API key) */
export interface ServerInfo {
  id: string;
  name: string;
  url: string;
  maskedApiKey: string;
  type: ServerType;
  createdAt: Date;
  updatedAt: Date;
}

/**
 * Mask an API key for display (show first 4 and last 4 characters)
 */
export function maskApiKey(apiKey: string): string {
  if (apiKey.length <= 8) {
    return "*".repeat(apiKey.length);
  }
  const first = apiKey.slice(0, 4);
  const last = apiKey.slice(-4);
  const middle = "*".repeat(Math.min(apiKey.length - 8, 20));
  return `${first}${middle}${last}`;
}

/**
 * Convert ServerConfig to ServerInfo with masked API key
 */
function toServerInfo(server: ServerConfig): ServerInfo {
  return {
    id: server.id,
    name: server.name,
    url: server.url,
    maskedApiKey: maskApiKey(server.apiKey),
    type: server.type,
    createdAt: server.createdAt,
    updatedAt: server.updatedAt,
  };
}

/**
 * Add a new server with connection validation
 */
export async function addServer(
  name: string,
  url: string,
  apiKey: string,
  type: ServerType
): Promise<ServerResult<ServerInfo>> {
  // Validate URL format
  const urlResult = validateUrl(url);
  if (!urlResult.success) {
    return urlResult;
  }
  const normalizedUrl = urlResult.data;

  // Check for duplicates
  const db = getDatabase();
  if (db.serverExists(normalizedUrl, type)) {
    return {
      success: false,
      error: `A ${type} server with this URL already exists`,
    };
  }

  // Check for duplicate name
  if (await db.getServerByName(name)) {
    return {
      success: false,
      error: `A server named "${name}" already exists`,
    };
  }

  // Test connection
  const client = createClient(normalizedUrl, apiKey, type);
  const connectionResult = await client.testConnection();

  if (!connectionResult.success) {
    return {
      success: false,
      error: `Connection failed: ${connectionResult.error}`,
    };
  }

  // Validate server type matches
  const status = connectionResult.data;
  const expectedApp = type === "radarr" ? "Radarr" : "Sonarr";
  if (status.appName !== expectedApp) {
    return {
      success: false,
      error: `Server is ${status.appName}, but ${expectedApp} was specified`,
    };
  }

  // Save to database
  const server = await db.addServer({ // Await addServer
    id: crypto.randomUUID(),
    name,
    url: normalizedUrl,
    apiKey,
    type,
  });

  return { success: true, data: toServerInfo(server) };
}

/**
 * Get all configured servers
 */
export async function listServers(): Promise<ServerInfo[]> {
  const db = getDatabase();
  const allServers = await db.getAllServers();
  return allServers.map(toServerInfo);
}

/**
 * Get a server by ID or name
 */
export async function getServer(idOrName: string): Promise<ServerResult<ServerConfig>> {
  const db = getDatabase();

  // Try by ID first
  let server = await db.getServer(idOrName);

  // Then try by name
  if (!server) {
    server = await db.getServerByName(idOrName);
  }

  if (!server) {
    return { success: false, error: `Server "${idOrName}" not found` };
  }

  return { success: true, data: server };
}

/**
 * Get servers by type
 */
export async function getServersByType(type: ServerType): Promise<ServerConfig[]> {
  const db = getDatabase();
  return await db.getServersByType(type);
}

/**
 * Edit a server with connection re-validation
 */
export async function editServer(
  idOrName: string,
  updates: { name?: string; url?: string; apiKey?: string }
): Promise<ServerResult<ServerInfo>> {
  const serverResult = await getServer(idOrName); // Await getServer
  if (!serverResult.success) {
    return serverResult;
  }

  const server = serverResult.data;
  const db = getDatabase();

  // Determine final values for validation
  let newUrl = server.url;
  const newApiKey = updates.apiKey ?? server.apiKey;

  // Validate and normalize URL if changed
  if (updates.url) {
    const urlResult = validateUrl(updates.url);
    if (!urlResult.success) {
      return urlResult;
    }
    newUrl = urlResult.data;

    // Check for duplicates with new URL
    if (db.serverExists(newUrl, server.type, server.id)) {
      return {
        success: false,
        error: `A ${server.type} server with this URL already exists`,
      };
    }
  }

  // Check for duplicate name
  if (updates.name && updates.name !== server.name) {
    const existing = await db.getServerByName(updates.name); // Await db.getServerByName
    if (existing && existing.id !== server.id) {
      return {
        success: false,
        error: `A server named "${updates.name}" already exists`,
      };
    }
  }

  // Test connection if URL or API key changed
  if (updates.url || updates.apiKey) {
    const client = createClient(newUrl, newApiKey, server.type);
    const connectionResult = await client.testConnection();

    if (!connectionResult.success) {
      return {
        success: false,
        error: `Connection failed with new settings: ${connectionResult.error}`,
      };
    }
  }

  // Apply updates
  const updated = await db.updateServer(server.id, { // Await db.updateServer
    name: updates.name,
    url: updates.url ? newUrl : undefined,
    apiKey: updates.apiKey,
  });

  if (!updated) {
    return { success: false, error: "Failed to update server" };
  }

  return { success: true, data: toServerInfo(updated) };
}

/**
 * Remove a server
 */
export async function removeServer(idOrName: string): Promise<ServerResult<void>> {
  const serverResult = await getServer(idOrName);
  if (!serverResult.success) {
    return serverResult;
  }

  const db = getDatabase();
  const deleted = db.deleteServer(serverResult.data.id);

  if (!deleted) {
    return { success: false, error: "Failed to delete server" };
  }

  return { success: true, data: undefined };
}

/**
 * Test connection to a specific server
 */
export async function testServerConnection(
  idOrName: string
): Promise<ServerResult<SystemStatus>> {
  const serverResult = await getServer(idOrName);
  if (!serverResult.success) {
    return serverResult;
  }

  const server = serverResult.data;
  const client = createClient(server.url, server.apiKey, server.type);

  const result = await client.testConnection();
  if (!result.success) {
    return { success: false, error: result.error };
  }

  return { success: true, data: result.data };
}

/**
 * Test connection to any URL/API key combination (for validation before add/edit)
 */
export async function testConnection(
  url: string,
  apiKey: string,
  type: ServerType
): Promise<ApiResult<SystemStatus>> {
  const urlResult = validateUrl(url);
  if (!urlResult.success) {
    return urlResult;
  }

  const client = createClient(urlResult.data, apiKey, type);
  return client.testConnection();
}
