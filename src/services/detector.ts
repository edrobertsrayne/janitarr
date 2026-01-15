/**
 * Content Detection Service
 *
 * Detects missing content and content below quality cutoff across all
 * configured Radarr and Sonarr servers.
 */

import type { ServerConfig, DetectionResult } from "../types";
import { RadarrClient, SonarrClient, createClient } from "../lib/api-client";
import { getDatabase } from "../storage/database";

/** Aggregated detection results across all servers */
export interface AggregatedResults {
  results: DetectionResult[];
  totalMissing: number;
  totalCutoff: number;
  successCount: number;
  failureCount: number;
}

/**
 * Detect missing and cutoff content on a single server (internal)
 */
async function detectOnServer(server: ServerConfig): Promise<DetectionResult> {
  const client = createClient(server.url, server.apiKey, server.type);

  const result: DetectionResult = {
    serverId: server.id,
    serverName: server.name,
    serverType: server.type,
    missingCount: 0,
    cutoffCount: 0,
    missingItems: [],
    cutoffItems: [],
  };

  // Get missing items
  const missingResult =
    server.type === "radarr"
      ? await (client as RadarrClient).getAllMissing()
      : await (client as SonarrClient).getAllMissing();

  if (missingResult.success) {
    result.missingItems = missingResult.data;
    result.missingCount = missingResult.data.length;
  } else {
    result.error = `Missing detection failed: ${missingResult.error}`;
    return result;
  }

  // Get cutoff unmet items
  const cutoffResult =
    server.type === "radarr"
      ? await (client as RadarrClient).getAllCutoffUnmet()
      : await (client as SonarrClient).getAllCutoffUnmet();

  if (cutoffResult.success) {
    result.cutoffItems = cutoffResult.data;
    result.cutoffCount = cutoffResult.data.length;
  } else {
    result.error = `Cutoff detection failed: ${cutoffResult.error}`;
    return result;
  }

  return result;
}

/**
 * Run detection on all configured servers
 */
export async function detectAll(): Promise<AggregatedResults> {
  const db = getDatabase();
  const servers = await db.getAllServers();

  const results: DetectionResult[] = [];
  let totalMissing = 0;
  let totalCutoff = 0;
  let successCount = 0;
  let failureCount = 0;

  // Run detection on all servers concurrently
  const detectionPromises = servers.map((server) => detectOnServer(server));
  const detectionResults = await Promise.all(detectionPromises);

  for (const result of detectionResults) {
    results.push(result);

    if (result.error) {
      failureCount++;
    } else {
      successCount++;
      totalMissing += result.missingCount;
      totalCutoff += result.cutoffCount;
    }
  }

  return {
    results,
    totalMissing,
    totalCutoff,
    successCount,
    failureCount,
  };
}

/**
 * Run detection on servers of a specific type
 */
export async function detectByType(
  type: "radarr" | "sonarr"
): Promise<AggregatedResults> {
  const db = getDatabase();
  const servers = await db.getServersByType(type);

  const results: DetectionResult[] = [];
  let totalMissing = 0;
  let totalCutoff = 0;
  let successCount = 0;
  let failureCount = 0;

  const detectionPromises = servers.map((server) => detectOnServer(server));
  const detectionResults = await Promise.all(detectionPromises);

  for (const result of detectionResults) {
    results.push(result);

    if (result.error) {
      failureCount++;
    } else {
      successCount++;
      totalMissing += result.missingCount;
      totalCutoff += result.cutoffCount;
    }
  }

  return {
    results,
    totalMissing,
    totalCutoff,
    successCount,
    failureCount,
  };
}

/**
 * Run detection on a single server by ID or name
 */
export async function detectSingleServer(
  idOrName: string
): Promise<DetectionResult | null> {
  const db = getDatabase();

  // Try by ID first
  let server = await db.getServer(idOrName);

  // Then try by name
  if (!server) {
    server = await db.getServerByName(idOrName);
  }

  if (!server) {
    return null;
  }

  return detectOnServer(server);
}
