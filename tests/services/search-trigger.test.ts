/**
 * Tests for search trigger service
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import {
  getSearchLimits,
  setSearchLimits,
  triggerSearches,
} from "../../src/services/search-trigger";
import type { AggregatedResults } from "../../src/services/detector";

const TEST_DB_PATH = "./data/test-search-trigger.db";

let testDb: DatabaseManager;

describe("Search Trigger Service", () => {
  beforeEach(() => {
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    process.env.JANITARR_DB_PATH = TEST_DB_PATH;
    testDb = new DatabaseManager(TEST_DB_PATH);
  });

  afterEach(() => {
    testDb.close();
    closeDatabase();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    delete process.env.JANITARR_DB_PATH;
  });

  describe("Search Limits", () => {
    test("returns default limits", () => {
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(10);
      expect(limits.missingEpisodesLimit).toBe(10);
      expect(limits.cutoffMoviesLimit).toBe(5);
      expect(limits.cutoffEpisodesLimit).toBe(5);
    });

    test("updates missing movies limit", () => {
      setSearchLimits(20, undefined, undefined, undefined);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(20);
      expect(limits.missingEpisodesLimit).toBe(10); // Unchanged
      expect(limits.cutoffMoviesLimit).toBe(5); // Unchanged
      expect(limits.cutoffEpisodesLimit).toBe(5); // Unchanged
    });

    test("updates missing episodes limit", () => {
      setSearchLimits(undefined, 15, undefined, undefined);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(10); // Unchanged
      expect(limits.missingEpisodesLimit).toBe(15);
      expect(limits.cutoffMoviesLimit).toBe(5); // Unchanged
      expect(limits.cutoffEpisodesLimit).toBe(5); // Unchanged
    });

    test("updates cutoff movies limit", () => {
      setSearchLimits(undefined, undefined, 12, undefined);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(10); // Unchanged
      expect(limits.missingEpisodesLimit).toBe(10); // Unchanged
      expect(limits.cutoffMoviesLimit).toBe(12);
      expect(limits.cutoffEpisodesLimit).toBe(5); // Unchanged
    });

    test("updates cutoff episodes limit", () => {
      setSearchLimits(undefined, undefined, undefined, 8);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(10); // Unchanged
      expect(limits.missingEpisodesLimit).toBe(10); // Unchanged
      expect(limits.cutoffMoviesLimit).toBe(5); // Unchanged
      expect(limits.cutoffEpisodesLimit).toBe(8);
    });

    test("updates all limits", () => {
      setSearchLimits(25, 30, 12, 15);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(25);
      expect(limits.missingEpisodesLimit).toBe(30);
      expect(limits.cutoffMoviesLimit).toBe(12);
      expect(limits.cutoffEpisodesLimit).toBe(15);
    });

    test("allows setting limits to zero", () => {
      setSearchLimits(0, 0, 0, 0);
      const limits = getSearchLimits();
      expect(limits.missingMoviesLimit).toBe(0);
      expect(limits.missingEpisodesLimit).toBe(0);
      expect(limits.cutoffMoviesLimit).toBe(0);
      expect(limits.cutoffEpisodesLimit).toBe(0);
    });
  });

  describe("triggerSearches", () => {
    test("returns empty results when no detection results", async () => {
      const detectionResults: AggregatedResults = {
        results: [],
        totalMissing: 0,
        totalCutoff: 0,
        successCount: 0,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults);

      expect(results.results).toEqual([]);
      expect(results.missingTriggered).toBe(0);
      expect(results.cutoffTriggered).toBe(0);
      expect(results.successCount).toBe(0);
      expect(results.failureCount).toBe(0);
    });

    test("respects missing limit of zero", async () => {
      setSearchLimits(0, 0, 5, 5);

      // Add a server
      await testDb.addServer({
        id: "server-1",
        name: "Test Server",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Server",
            serverType: "radarr",
            missingCount: 5,
            cutoffCount: 3,
            missingItems: [
              { id: 1, title: "Movie 1", type: "movie" },
              { id: 2, title: "Movie 2", type: "movie" },
            ],
            cutoffItems: [{ id: 3, title: "Movie 3", type: "movie" }],
          },
        ],
        totalMissing: 5,
        totalCutoff: 3,
        successCount: 1,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults);

      // Should not trigger any missing searches
      expect(results.missingTriggered).toBe(0);
    });

    test("respects cutoff limit of zero", async () => {
      setSearchLimits(5, 5, 0, 0);

      await testDb.addServer({
        id: "server-1",
        name: "Test Server",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Server",
            serverType: "radarr",
            missingCount: 5,
            cutoffCount: 3,
            missingItems: [{ id: 1, title: "Movie 1", type: "movie" }],
            cutoffItems: [
              { id: 2, title: "Movie 2", type: "movie" },
              { id: 3, title: "Movie 3", type: "movie" },
            ],
          },
        ],
        totalMissing: 5,
        totalCutoff: 3,
        successCount: 1,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults);

      // Should not trigger any cutoff searches
      expect(results.cutoffTriggered).toBe(0);
    });

    test("skips failed detection results", async () => {
      await testDb.addServer({
        id: "server-1",
        name: "Test Server",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Server",
            serverType: "radarr",
            missingCount: 0,
            cutoffCount: 0,
            missingItems: [],
            cutoffItems: [],
            error: "Connection failed",
          },
        ],
        totalMissing: 0,
        totalCutoff: 0,
        successCount: 0,
        failureCount: 1,
      };

      const results = await triggerSearches(detectionResults);

      // Should skip the failed server
      expect(results.results).toEqual([]);
      expect(results.missingTriggered).toBe(0);
    });
  });

  describe("triggerSearches - dry run mode", () => {
    test("dry run returns expected results without triggering searches", async () => {
      setSearchLimits(2, 2, 1, 1);

      await testDb.addServer({
        id: "server-1",
        name: "Test Radarr",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      await testDb.addServer({
        id: "server-2",
        name: "Test Sonarr",
        url: "http://localhost:59998",
        apiKey: "fake-key",
        type: "sonarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Radarr",
            serverType: "radarr",
            missingCount: 5,
            cutoffCount: 3,
            missingItems: [
              { id: 1, title: "Movie 1", type: "movie" },
              { id: 2, title: "Movie 2", type: "movie" },
              { id: 3, title: "Movie 3", type: "movie" },
            ],
            cutoffItems: [
              { id: 4, title: "Movie 4", type: "movie" },
              { id: 5, title: "Movie 5", type: "movie" },
            ],
          },
          {
            serverId: "server-2",
            serverName: "Test Sonarr",
            serverType: "sonarr",
            missingCount: 8,
            cutoffCount: 4,
            missingItems: [
              { id: 10, title: "Episode 1", type: "episode" },
              { id: 11, title: "Episode 2", type: "episode" },
              { id: 12, title: "Episode 3", type: "episode" },
            ],
            cutoffItems: [
              { id: 13, title: "Episode 4", type: "episode" },
              { id: 14, title: "Episode 5", type: "episode" },
            ],
          },
        ],
        totalMissing: 13,
        totalCutoff: 7,
        successCount: 2,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults, true);

      // Should return what would be triggered (respecting limits)
      // Missing movies limit: 2, Missing episodes limit: 2
      // Total missing: 2 + 2 = 4
      expect(results.missingTriggered).toBe(4);
      // Cutoff movies limit: 1, Cutoff episodes limit: 1
      // Total cutoff: 1 + 1 = 2
      expect(results.cutoffTriggered).toBe(2);
      // All operations should succeed in dry-run
      expect(results.successCount).toBeGreaterThan(0);
      expect(results.failureCount).toBe(0);

      // Verify results contain the expected structure
      expect(results.results.length).toBeGreaterThan(0);
      // Filter to only results with items (some may be empty due to distribution)
      const nonEmptyResults = results.results.filter(r => r.itemIds.length > 0);
      expect(nonEmptyResults.length).toBeGreaterThan(0);
      for (const result of nonEmptyResults) {
        expect(result.success).toBe(true);
        expect(result.itemIds.length).toBeGreaterThan(0);
      }
    });

    test("dry run respects all four separate limits", async () => {
      setSearchLimits(1, 2, 3, 4);

      await testDb.addServer({
        id: "server-1",
        name: "Test Radarr",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      await testDb.addServer({
        id: "server-2",
        name: "Test Sonarr",
        url: "http://localhost:59998",
        apiKey: "fake-key",
        type: "sonarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Radarr",
            serverType: "radarr",
            missingCount: 10,
            cutoffCount: 10,
            missingItems: Array.from({ length: 10 }, (_, i) => ({
              id: i + 1,
              title: `Movie ${i + 1}`,
              type: "movie" as const,
            })),
            cutoffItems: Array.from({ length: 10 }, (_, i) => ({
              id: i + 100,
              title: `Movie Cutoff ${i + 1}`,
              type: "movie" as const,
            })),
          },
          {
            serverId: "server-2",
            serverName: "Test Sonarr",
            serverType: "sonarr",
            missingCount: 10,
            cutoffCount: 10,
            missingItems: Array.from({ length: 10 }, (_, i) => ({
              id: i + 200,
              title: `Episode ${i + 1}`,
              type: "episode" as const,
            })),
            cutoffItems: Array.from({ length: 10 }, (_, i) => ({
              id: i + 300,
              title: `Episode Cutoff ${i + 1}`,
              type: "episode" as const,
            })),
          },
        ],
        totalMissing: 20,
        totalCutoff: 20,
        successCount: 2,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults, true);

      // Missing movies: 1, Missing episodes: 2, Cutoff movies: 3, Cutoff episodes: 4
      // Total missing: 1 + 2 = 3
      expect(results.missingTriggered).toBe(3);
      // Total cutoff: 3 + 4 = 7
      expect(results.cutoffTriggered).toBe(7);
      // Total: 10 searches
      expect(results.missingTriggered + results.cutoffTriggered).toBe(10);
      expect(results.successCount).toBeGreaterThan(0);
      expect(results.failureCount).toBe(0);
    });

    test("dry run with zero limits returns no results", async () => {
      setSearchLimits(0, 0, 0, 0);

      await testDb.addServer({
        id: "server-1",
        name: "Test Server",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "server-1",
            serverName: "Test Server",
            serverType: "radarr",
            missingCount: 5,
            cutoffCount: 3,
            missingItems: [{ id: 1, title: "Movie 1", type: "movie" }],
            cutoffItems: [{ id: 2, title: "Movie 2", type: "movie" }],
          },
        ],
        totalMissing: 5,
        totalCutoff: 3,
        successCount: 1,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults, true);

      expect(results.missingTriggered).toBe(0);
      expect(results.cutoffTriggered).toBe(0);
      expect(results.results).toEqual([]);
    });
  });
});

// Integration tests
const RADARR_URL = process.env.RADARR_URL ?? "";
const RADARR_API_KEY = process.env.RADARR_API_KEY ?? "";
const hasRadarr = RADARR_URL && RADARR_API_KEY;

describe("Search Trigger Integration", () => {
  beforeEach(() => {
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    process.env.JANITARR_DB_PATH = TEST_DB_PATH;
    testDb = new DatabaseManager(TEST_DB_PATH);
  });

  afterEach(() => {
    testDb.close();
    closeDatabase();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    delete process.env.JANITARR_DB_PATH;
  });

  // Note: We don't actually trigger searches in integration tests to avoid
  // affecting the real media servers. We just test the plumbing works.
  test.skipIf(!hasRadarr)(
    "can configure limits and process empty detection results",
    async () => {
      await testDb.addServer({
        id: "radarr-test",
        name: "Test Radarr",
        url: RADARR_URL,
        apiKey: RADARR_API_KEY,
        type: "radarr",
      });

      setSearchLimits(5, 5, 3, 3);

      const detectionResults: AggregatedResults = {
        results: [
          {
            serverId: "radarr-test",
            serverName: "Test Radarr",
            serverType: "radarr",
            missingCount: 0,
            cutoffCount: 0,
            missingItems: [],
            cutoffItems: [],
          },
        ],
        totalMissing: 0,
        totalCutoff: 0,
        successCount: 1,
        failureCount: 0,
      };

      const results = await triggerSearches(detectionResults);

      // No items to search, so success with zero triggered
      expect(results.missingTriggered).toBe(0);
      expect(results.cutoffTriggered).toBe(0);
    }
  );
});
