/**
 * Configuration API routes
 */

import type { DatabaseManager } from "../../storage/database";
import { jsonSuccess, jsonError, parseJsonBody, HttpStatus } from "../types";
import type { AppConfig } from "../../types";

/**
 * Handle GET /api/config
 */
export async function handleGetConfig(db: DatabaseManager): Promise<Response> {
  try {
    const config = db.getAppConfig();
    return jsonSuccess(config);
  } catch (error) {
    return jsonError(
      `Failed to retrieve config: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle PATCH /api/config
 */
export async function handlePatchConfig(req: Request, db: DatabaseManager): Promise<Response> {
  try {
    const body = await parseJsonBody<Partial<AppConfig>>(req);
    if (!body) {
      return jsonError("Invalid JSON body", HttpStatus.BAD_REQUEST);
    }

    // Validate schedule config if provided
    if (body.schedule) {
      if (body.schedule.intervalHours !== undefined) {
        if (body.schedule.intervalHours < 1 || body.schedule.intervalHours > 168) {
          return jsonError("Interval hours must be between 1 and 168", HttpStatus.BAD_REQUEST);
        }
      }
    }

    // Validate search limits if provided
    if (body.searchLimits) {
      const limits = body.searchLimits;
      if (limits.missingMoviesLimit !== undefined && (limits.missingMoviesLimit < 0 || limits.missingMoviesLimit > 1000)) {
        return jsonError("Missing movies limit must be between 0 and 1000", HttpStatus.BAD_REQUEST);
      }
      if (limits.missingEpisodesLimit !== undefined && (limits.missingEpisodesLimit < 0 || limits.missingEpisodesLimit > 1000)) {
        return jsonError("Missing episodes limit must be between 0 and 1000", HttpStatus.BAD_REQUEST);
      }
      if (limits.cutoffMoviesLimit !== undefined && (limits.cutoffMoviesLimit < 0 || limits.cutoffMoviesLimit > 1000)) {
        return jsonError("Cutoff movies limit must be between 0 and 1000", HttpStatus.BAD_REQUEST);
      }
      if (limits.cutoffEpisodesLimit !== undefined && (limits.cutoffEpisodesLimit < 0 || limits.cutoffEpisodesLimit > 1000)) {
        return jsonError("Cutoff episodes limit must be between 0 and 1000", HttpStatus.BAD_REQUEST);
      }
    }

    // Update config
    db.setAppConfig(body);

    // Return updated config
    const updatedConfig = db.getAppConfig();
    return jsonSuccess(updatedConfig);
  } catch (error) {
    return jsonError(
      `Failed to update config: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}

/**
 * Handle PUT /api/config/reset
 */
export async function handleResetConfig(db: DatabaseManager): Promise<Response> {
  try {
    // Reset to default values
    const defaults: AppConfig = {
      schedule: {
        intervalHours: 6,
        enabled: true,
      },
      searchLimits: {
        missingMoviesLimit: 10,
        missingEpisodesLimit: 10,
        cutoffMoviesLimit: 5,
        cutoffEpisodesLimit: 5,
      },
    };

    db.setAppConfig(defaults);

    const resetConfig = db.getAppConfig();
    return jsonSuccess(resetConfig);
  } catch (error) {
    return jsonError(
      `Failed to reset config: ${error instanceof Error ? error.message : String(error)}`,
      HttpStatus.INTERNAL_SERVER_ERROR
    );
  }
}
