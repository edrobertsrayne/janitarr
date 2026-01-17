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
import { getDatabase, DatabaseManager } from "../storage/database"; // Import DatabaseManager

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
 * Get all configured servers (ServerConfig objects)
 */
export async function getServers(db: DatabaseManager): Promise<ServerResult<ServerConfig[]>> {
  const allServers = await db.getAllServers();
  return { success: true, data: allServers };
}

/**
 * Get all configured servers for display (ServerInfo objects)
 */
export async function listServers(): Promise<ServerInfo[]> {
  const db = getDatabase();
  const allServers = await db.getAllServers();
  return allServers.map(toServerInfo);
}

/**
 * Get a server by ID (ServerConfig object)
 */
export async function getServerById(db: DatabaseManager, id: string): Promise<ServerResult<ServerConfig>> {
  const server = await db.getServer(id);
  if (!server) {
    return { success: false, error: `Server with ID "${id}" not found` };
  }
  return { success: true, data: server };
}

/**
 * Get a server by ID or name (ServerConfig object)
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
  const server = await db.addServer({
    id: crypto.randomUUID(),
    name,
    url: normalizedUrl,
    apiKey,
    type,
  });

  return { success: true, data: toServerInfo(server) };
}

/**
 * Update a server with connection re-validation
 */
export async function updateServer(
  db: DatabaseManager, // Pass db explicitly for web routes
  id: string,
  updates: { name?: string; url?: string; apiKey?: string; enabled?: boolean; type?: ServerType }
): Promise<ServerResult<ServerInfo>> {
  const serverResult = await getServerById(db, id);
  if (!serverResult.success) {
    return serverResult;
  }

  const server = serverResult.data;

  // Determine final values for validation
  let newUrl = server.url;
  const newApiKey = updates.apiKey ?? server.apiKey;
  const newName = updates.name ?? server.name;
  const newType = updates.type ?? server.type;
  const newEnabled = updates.enabled !== undefined ? updates.enabled : server.enabled;


  // Validate and normalize URL if changed
  if (updates.url) {
    const urlResult = validateUrl(updates.url);
    if (!urlResult.success) {
      return urlResult;
    }
    newUrl = urlResult.data;

    // Check for duplicates with new URL
    if (db.serverExists(newUrl, newType, id)) { // Use newType
      return {
        success: false,
        error: `A ${newType} server with this URL already exists`,
      };
    }
  }

  // Check for duplicate name
  if (newName !== server.name) {
    const existing = await db.getServerByName(newName);
    if (existing && existing.id !== id) {
      return {
        success: false,
        error: `A server named "${newName}" already exists`,
      };
    }
  }

  // Test connection if URL, API key or Type changed
  if (newUrl !== server.url || newApiKey !== server.apiKey || newType !== server.type) {
    const connectionResult = await testConnection(newUrl, newApiKey, newType); // Use the global testConnection
    if (!connectionResult.success) {
      return {
        success: false,
        error: `Connection failed with new settings: ${connectionResult.error}`,
      };
    }
  }

  // Apply updates
  const updated = await db.updateServer(id, {
    name: newName,
    url: newUrl,
    apiKey: newApiKey,
    type: newType,
    enabled: newEnabled,
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
 * Test connection to a specific server (by ID or Name)
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
 * Test connection to a given server configuration (not necessarily saved in DB)
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

/**
 * Validate a server configuration against rules (e.g. name uniqueness, URL format).
 * Does not test connection.
 */
export async function validateServerConfig(
  config: Partial<ServerConfig>,
  excludeId?: string
): Promise<ServerResult<true>> {
  const db = getDatabase();

  if (config.url) {
    const urlResult = validateUrl(config.url);
    if (!urlResult.success) {
      return urlResult;
    }
  }

  if (config.name) {
    const existing = await db.getServerByName(config.name);
    if (existing && existing.id !== excludeId) {
      return {
        success: false,
        error: `A server named "${config.name}" already exists`,
      };
    }
  }
  return { success: true, data: true };
}
