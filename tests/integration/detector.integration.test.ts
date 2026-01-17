/**
 * Integration tests for content detection service
 * These tests require real Radarr/Sonarr servers to be running
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import { detectAll } from "../../src/services/detector";

const TEST_DB_PATH = "./data/test-detector-integration.db";

// Get test database instance for setup
let testDb: DatabaseManager;

// Integration tests against real servers
const RADARR_URL = process.env.RADARR_URL ?? "";
const RADARR_API_KEY = process.env.RADARR_API_KEY ?? "";
const SONARR_URL = process.env.SONARR_URL ?? "";
const SONARR_API_KEY = process.env.SONARR_API_KEY ?? "";

const hasRadarr = RADARR_URL && RADARR_API_KEY;
const hasSonarr = SONARR_URL && SONARR_API_KEY;

describe("Detector Integration", () => {
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

  test.skipIf(!hasRadarr)("detects missing movies from Radarr", async () => {
    await testDb.addServer({
      id: "radarr-test",
      name: "Test Radarr",
      url: RADARR_URL,
      apiKey: RADARR_API_KEY,
      type: "radarr",
    });

    const results = await detectAll();

    expect(results.successCount).toBe(1);
    expect(results.failureCount).toBe(0);
    expect(results.results[0].error).toBeUndefined();
    expect(results.results[0].missingCount).toBeGreaterThanOrEqual(0);
    expect(results.results[0].cutoffCount).toBeGreaterThanOrEqual(0);
  });

  test.skipIf(!hasSonarr)("detects missing episodes from Sonarr", async () => {
    await testDb.addServer({
      id: "sonarr-test",
      name: "Test Sonarr",
      url: SONARR_URL,
      apiKey: SONARR_API_KEY,
      type: "sonarr",
    });

    const results = await detectAll();

    expect(results.successCount).toBe(1);
    expect(results.failureCount).toBe(0);
    expect(results.results[0].error).toBeUndefined();
    expect(results.results[0].missingCount).toBeGreaterThanOrEqual(0);
    expect(results.results[0].cutoffCount).toBeGreaterThanOrEqual(0);
  });

  test.skipIf(!hasRadarr || !hasSonarr)(
    "detects from multiple servers",
    async () => {
      await testDb.addServer({
        id: "radarr-multi",
        name: "Multi Radarr",
        url: RADARR_URL,
        apiKey: RADARR_API_KEY,
        type: "radarr",
      });
      await testDb.addServer({
        id: "sonarr-multi",
        name: "Multi Sonarr",
        url: SONARR_URL,
        apiKey: SONARR_API_KEY,
        type: "sonarr",
      });

      const results = await detectAll();

      expect(results.results.length).toBe(2);
      expect(results.successCount).toBe(2);
      expect(results.failureCount).toBe(0);
      expect(results.totalMissing).toBeGreaterThanOrEqual(0);
      expect(results.totalCutoff).toBeGreaterThanOrEqual(0);
    }
  );
});
