/**
 * Search Trigger Service
 *
 * Triggers searches for missing and cutoff content based on detection results
 * and user-configured limits.
 */

import type { ServerType, MediaItem, DetectionResult } from "../types";
import { RadarrClient, SonarrClient, createClient } from "../lib/api-client";
import { getDatabase } from "../storage/database";
import type { AggregatedResults } from "./detector";

/** Result of a single search trigger attempt */
export interface SearchTriggerResult {
  serverId: string;
  serverName: string;
  serverType: ServerType;
  category: "missing" | "cutoff";
  itemIds: number[];
  success: boolean;
  error?: string;
}

/** Aggregated results from triggering searches */
export interface TriggerResults {
  results: SearchTriggerResult[];
  missingTriggered: number;
  cutoffTriggered: number;
  successCount: number;
  failureCount: number;
}

/**
 * Distribute items fairly across servers using round-robin
 */
function distributeItems(
  detectionResults: DetectionResult[],
  category: "missing" | "cutoff",
  itemType: "movie" | "episode",
  limit: number
): Map<string, MediaItem[]> {
  const serverItems = new Map<string, MediaItem[]>();

  // Initialize map with empty arrays for each server
  for (const result of detectionResults) {
    if (!result.error) {
      serverItems.set(result.serverId, []);
    }
  }

  // Get all items with their server IDs, filtered by type
  const allItems: Array<{ serverId: string; item: MediaItem }> = [];
  for (const result of detectionResults) {
    if (result.error) continue;
    const items = category === "missing" ? result.missingItems : result.cutoffItems;
    for (const item of items) {
      // Filter by item type
      if (item.type === itemType) {
        allItems.push({ serverId: result.serverId, item });
      }
    }
  }

  // Round-robin distribution across servers up to limit
  let distributed = 0;
  let serverIds = Array.from(serverItems.keys());

  // Keep cycling through servers until we hit the limit or run out of items
  while (distributed < limit && distributed < allItems.length) {
    for (const serverId of serverIds) {
      if (distributed >= limit) break;

      // Find next available item for this server
      const itemIndex = allItems.findIndex(
        (i) => i.serverId === serverId && !serverItems.get(serverId)?.includes(i.item)
      );

      if (itemIndex !== -1) {
        const { item } = allItems[itemIndex];
        serverItems.get(serverId)!.push(item);
        allItems.splice(itemIndex, 1);
        distributed++;
      }
    }

    // Remove servers that have no more items
    serverIds = serverIds.filter((id) =>
      allItems.some((i) => i.serverId === id)
    );

    // If no servers have items left, break
    if (serverIds.length === 0) break;
  }

  return serverItems;
}

/**
 * Trigger search for items on a specific server
 */
async function triggerServerSearch(
  serverId: string,
  serverName: string,
  serverType: ServerType,
  serverUrl: string,
  apiKey: string,
  items: MediaItem[],
  category: "missing" | "cutoff",
  dryRun = false
): Promise<SearchTriggerResult> {
  if (items.length === 0) {
    return {
      serverId,
      serverName,
      serverType,
      category,
      itemIds: [],
      success: true,
    };
  }

  const itemIds = items.map((i) => i.id);

  // In dry-run mode, skip actual API calls and return what would be triggered
  if (dryRun) {
    return {
      serverId,
      serverName,
      serverType,
      category,
      itemIds,
      success: true,
    };
  }

  const client = createClient(serverUrl, apiKey, serverType);

  let result;
  if (serverType === "radarr") {
    result = await (client as RadarrClient).searchMovies(itemIds);
  } else {
    result = await (client as SonarrClient).searchEpisodes(itemIds);
  }

  if (result.success) {
    return {
      serverId,
      serverName,
      serverType,
      category,
      itemIds,
      success: true,
    };
  } else {
    return {
      serverId,
      serverName,
      serverType,
      category,
      itemIds,
      success: false,
      error: result.error,
    };
  }
}

/**
 * Trigger searches based on detection results and configured limits
 *
 * @param detectionResults - Results from detection phase
 * @param dryRun - If true, preview what would be searched without triggering actual searches
 */
export async function triggerSearches(
  detectionResults: AggregatedResults,
  dryRun = false
): Promise<TriggerResults> {
  const db = getDatabase();
  const config = db.getAppConfig();
  const servers = await db.getAllServers();

  // Create a map of server ID to server config for URL/API key lookup
  const serverMap = new Map(servers.map((s) => [s.id, s]));

  const results: SearchTriggerResult[] = [];
  let missingTriggered = 0;
  let cutoffTriggered = 0;
  let successCount = 0;
  let failureCount = 0;

  // Handle missing movies
  if (config.searchLimits.missingMoviesLimit > 0) {
    const missingMoviesDistribution = distributeItems(
      detectionResults.results,
      "missing",
      "movie",
      config.searchLimits.missingMoviesLimit
    );

    for (const [serverId, items] of missingMoviesDistribution) {
      const server = serverMap.get(serverId);
      if (!server) continue;

      const result = await triggerServerSearch(
        serverId,
        server.name,
        server.type,
        server.url,
        server.apiKey,
        items,
        "missing",
        dryRun
      );

      results.push(result);

      if (result.success) {
        successCount++;
        missingTriggered += result.itemIds.length;
      } else {
        failureCount++;
      }
    }
  }

  // Handle missing episodes
  if (config.searchLimits.missingEpisodesLimit > 0) {
    const missingEpisodesDistribution = distributeItems(
      detectionResults.results,
      "missing",
      "episode",
      config.searchLimits.missingEpisodesLimit
    );

    for (const [serverId, items] of missingEpisodesDistribution) {
      const server = serverMap.get(serverId);
      if (!server) continue;

      const result = await triggerServerSearch(
        serverId,
        server.name,
        server.type,
        server.url,
        server.apiKey,
        items,
        "missing",
        dryRun
      );

      results.push(result);

      if (result.success) {
        successCount++;
        missingTriggered += result.itemIds.length;
      } else {
        failureCount++;
      }
    }
  }

  // Handle cutoff movies
  if (config.searchLimits.cutoffMoviesLimit > 0) {
    const cutoffMoviesDistribution = distributeItems(
      detectionResults.results,
      "cutoff",
      "movie",
      config.searchLimits.cutoffMoviesLimit
    );

    for (const [serverId, items] of cutoffMoviesDistribution) {
      const server = serverMap.get(serverId);
      if (!server) continue;

      const result = await triggerServerSearch(
        serverId,
        server.name,
        server.type,
        server.url,
        server.apiKey,
        items,
        "cutoff",
        dryRun
      );

      results.push(result);

      if (result.success) {
        successCount++;
        cutoffTriggered += result.itemIds.length;
      } else {
        failureCount++;
      }
    }
  }

  // Handle cutoff episodes
  if (config.searchLimits.cutoffEpisodesLimit > 0) {
    const cutoffEpisodesDistribution = distributeItems(
      detectionResults.results,
      "cutoff",
      "episode",
      config.searchLimits.cutoffEpisodesLimit
    );

    for (const [serverId, items] of cutoffEpisodesDistribution) {
      const server = serverMap.get(serverId);
      if (!server) continue;

      const result = await triggerServerSearch(
        serverId,
        server.name,
        server.type,
        server.url,
        server.apiKey,
        items,
        "cutoff",
        dryRun
      );

      results.push(result);

      if (result.success) {
        successCount++;
        cutoffTriggered += result.itemIds.length;
      } else {
        failureCount++;
      }
    }
  }

  return {
    results,
    missingTriggered,
    cutoffTriggered,
    successCount,
    failureCount,
  };
}

/**
 * Get current search limits from configuration
 */
export function getSearchLimits(): {
  missingMoviesLimit: number;
  missingEpisodesLimit: number;
  cutoffMoviesLimit: number;
  cutoffEpisodesLimit: number;
} {
  const db = getDatabase();
  const config = db.getAppConfig();
  return config.searchLimits;
}

/**
 * Update search limits
 */
export function setSearchLimits(
  missingMoviesLimit?: number,
  missingEpisodesLimit?: number,
  cutoffMoviesLimit?: number,
  cutoffEpisodesLimit?: number
): void {
  const db = getDatabase();
  db.setAppConfig({
    searchLimits: {
      missingMoviesLimit,
      missingEpisodesLimit,
      cutoffMoviesLimit,
      cutoffEpisodesLimit,
    },
  });
}
