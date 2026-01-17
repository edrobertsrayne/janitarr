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
