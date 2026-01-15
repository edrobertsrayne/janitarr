/**
 * Tests for content detection service
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import { detectAll, detectByType, detectSingleServer } from "../../src/services/detector";

const TEST_DB_PATH = "./data/test-detector.db";

// Get test database instance for setup
let testDb: DatabaseManager;

describe("Detector Service", () => {
  beforeEach(() => {
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    // Set env var before getting database
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

  describe("detectAll", () => {
    test("returns empty results when no servers configured", async () => {
      const results = await detectAll();

      expect(results.results).toEqual([]);
      expect(results.totalMissing).toBe(0);
      expect(results.totalCutoff).toBe(0);
      expect(results.successCount).toBe(0);
      expect(results.failureCount).toBe(0);
    });

    test("handles unreachable servers gracefully", async () => {
      // Add a server that won't be reachable
      await testDb.addServer({
        id: "test-1",
        name: "Unreachable",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const results = await detectAll();

      expect(results.results.length).toBe(1);
      expect(results.failureCount).toBe(1);
      expect(results.successCount).toBe(0);
      expect(results.results[0].error).toBeDefined();
    });
  });

  describe("detectByType", () => {
    test("only detects servers of specified type", async () => {
      // Add servers of both types (neither reachable)
      await testDb.addServer({
        id: "radarr-1",
        name: "Radarr",
        url: "http://localhost:59997",
        apiKey: "fake-key",
        type: "radarr",
      });
      await testDb.addServer({
        id: "sonarr-1",
        name: "Sonarr",
        url: "http://localhost:59998",
        apiKey: "fake-key",
        type: "sonarr",
      });

      const radarrResults = await detectByType("radarr");
      expect(radarrResults.results.length).toBe(1);
      expect(radarrResults.results[0].serverType).toBe("radarr");

      const sonarrResults = await detectByType("sonarr");
      expect(sonarrResults.results.length).toBe(1);
      expect(sonarrResults.results[0].serverType).toBe("sonarr");
    });
  });

  describe("detectSingleServer", () => {
    test("returns null for non-existent server", async () => {
      const result = await detectSingleServer("non-existent");
      expect(result).toBeNull();
    });

    test("finds server by ID", async () => {
      await testDb.addServer({
        id: "test-id-123",
        name: "Test Server",
        url: "http://localhost:59996",
        apiKey: "fake-key",
        type: "radarr",
      });

      const result = await detectSingleServer("test-id-123");
      expect(result).not.toBeNull();
      expect(result?.serverId).toBe("test-id-123");
    });

    test("finds server by name", async () => {
      await testDb.addServer({
        id: "some-id",
        name: "My Radarr Server",
        url: "http://localhost:59995",
        apiKey: "fake-key",
        type: "radarr",
      });

      const result = await detectSingleServer("My Radarr Server");
      expect(result).not.toBeNull();
      expect(result?.serverName).toBe("My Radarr Server");
    });
  });
});

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
