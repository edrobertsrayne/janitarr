/**
 * Tests for automation orchestrator
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import { runAutomationCycle, formatCycleResult } from "../../src/services/automation";
import { getRecentLogs, getLastCycleSummary } from "../../src/lib/logger";

const TEST_DB_PATH = "./data/test-automation.db";

let testDb: DatabaseManager;

describe("Automation Orchestrator", () => {
  beforeEach(() => {
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    process.env.JANITARR_DB_PATH = TEST_DB_PATH;
    testDb = new DatabaseManager(TEST_DB_PATH);

    // Set reasonable search limits
    testDb.setAppConfig({
      searchLimits: {
        missingLimit: 10,
        cutoffLimit: 5,
      },
    });
  });

  afterEach(() => {
    testDb.close();
    closeDatabase();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    delete process.env.JANITARR_DB_PATH;
  });

  describe("runAutomationCycle", () => {
    test("completes successfully with no servers", async () => {
      const result = await runAutomationCycle(false);

      expect(result.success).toBe(true);
      expect(result.detectionResults.successCount).toBe(0);
      expect(result.detectionResults.failureCount).toBe(0);
      expect(result.searchResults.missingTriggered).toBe(0);
      expect(result.searchResults.cutoffTriggered).toBe(0);
      expect(result.totalSearches).toBe(0);
      expect(result.totalFailures).toBe(0);
      expect(result.errors).toEqual([]);
    });

    test("logs cycle start and end", async () => {
      await runAutomationCycle(false);

      const logs = getRecentLogs(10);

      // Should have cycle_start and cycle_end
      expect(logs.some((l) => l.type === "cycle_start")).toBe(true);
      expect(logs.some((l) => l.type === "cycle_end")).toBe(true);
    });

    test("marks manual cycles correctly", async () => {
      await runAutomationCycle(true);

      const logs = getRecentLogs(10);
      const cycleStart = logs.find((l) => l.type === "cycle_start");
      const cycleEnd = logs.find((l) => l.type === "cycle_end");

      expect(cycleStart?.isManual).toBe(true);
      expect(cycleEnd?.isManual).toBe(true);
    });

    test("marks scheduled cycles correctly", async () => {
      await runAutomationCycle(false);

      const logs = getRecentLogs(10);
      const cycleStart = logs.find((l) => l.type === "cycle_start");
      const cycleEnd = logs.find((l) => l.type === "cycle_end");

      expect(cycleStart?.isManual).toBe(false);
      expect(cycleEnd?.isManual).toBe(false);
    });

    test("handles unreachable servers gracefully", async () => {
      // Add unreachable server
      testDb.addServer({
        id: "test-1",
        name: "Unreachable",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      const result = await runAutomationCycle(false);

      expect(result.success).toBe(false);
      expect(result.detectionResults.failureCount).toBe(1);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors[0]).toContain("Detection failed for Unreachable");
    });

    test("logs server errors", async () => {
      testDb.addServer({
        id: "test-1",
        name: "Unreachable",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      await runAutomationCycle(false);

      const logs = getRecentLogs(10);
      const errorLog = logs.find((l) => l.type === "error");

      expect(errorLog).toBeDefined();
      expect(errorLog?.serverName).toBe("Unreachable");
      expect(errorLog?.serverType).toBe("radarr");
      expect(errorLog?.message).toContain("Connection failed");
    });

    test("continues with other servers if one fails", async () => {
      // Add one unreachable and one reachable (but still fake)
      testDb.addServer({
        id: "test-1",
        name: "Unreachable1",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });
      testDb.addServer({
        id: "test-2",
        name: "Unreachable2",
        url: "http://localhost:59998",
        apiKey: "fake-key",
        type: "sonarr",
      });

      const result = await runAutomationCycle(false);

      // Both should fail, but both should be attempted
      expect(result.detectionResults.failureCount).toBe(2);
      expect(result.errors.length).toBe(2);
    });

    test("updates last cycle summary", async () => {
      await runAutomationCycle(false);

      const summary = getLastCycleSummary();

      expect(summary.lastCycleTime).not.toBeNull();
      expect(summary.lastCycleSearches).toBe(0);
      expect(summary.lastCycleFailures).toBe(0);
      expect(summary.wasManual).toBe(false);
    });

    test("records failures in cycle summary", async () => {
      testDb.addServer({
        id: "test-1",
        name: "Unreachable",
        url: "http://localhost:59999",
        apiKey: "fake-key",
        type: "radarr",
      });

      await runAutomationCycle(false);

      const summary = getLastCycleSummary();

      expect(summary.lastCycleTime).not.toBeNull();
      expect(summary.lastCycleFailures).toBeGreaterThan(0);
    });
  });

  describe("formatCycleResult", () => {
    test("formats successful cycle", () => {
      const result = {
        success: true,
        detectionResults: {
          totalMissing: 5,
          totalCutoff: 3,
          successCount: 2,
          failureCount: 0,
        },
        searchResults: {
          missingTriggered: 5,
          cutoffTriggered: 3,
          successCount: 2,
          failureCount: 0,
        },
        totalSearches: 8,
        totalFailures: 0,
        errors: [],
      };

      const formatted = formatCycleResult(result);

      expect(formatted).toContain("Automation Cycle Summary");
      expect(formatted).toContain("5 missing, 3 cutoff");
      expect(formatted).toContain("5 missing, 3 cutoff");
      expect(formatted).toContain("8 searches");
      expect(formatted).toContain("✓ Success");
      expect(formatted).not.toContain("Errors:");
    });

    test("formats failed cycle with errors", () => {
      const result = {
        success: false,
        detectionResults: {
          totalMissing: 0,
          totalCutoff: 0,
          successCount: 0,
          failureCount: 2,
        },
        searchResults: {
          missingTriggered: 0,
          cutoffTriggered: 0,
          successCount: 0,
          failureCount: 0,
        },
        totalSearches: 0,
        totalFailures: 2,
        errors: [
          "Detection failed for Server1: Connection timeout",
          "Detection failed for Server2: Invalid API key",
        ],
      };

      const formatted = formatCycleResult(result);

      expect(formatted).toContain("✗ 2 failures");
      expect(formatted).toContain("Errors:");
      expect(formatted).toContain("Server1: Connection timeout");
      expect(formatted).toContain("Server2: Invalid API key");
    });

    test("formats partial success", () => {
      const result = {
        success: false,
        detectionResults: {
          totalMissing: 3,
          totalCutoff: 2,
          successCount: 1,
          failureCount: 1,
        },
        searchResults: {
          missingTriggered: 3,
          cutoffTriggered: 0,
          successCount: 1,
          failureCount: 1,
        },
        totalSearches: 3,
        totalFailures: 2,
        errors: [
          "Detection failed for BadServer: Timeout",
          "Search trigger failed for GoodServer (cutoff): Rate limited",
        ],
      };

      const formatted = formatCycleResult(result);

      expect(formatted).toContain("1 successful, 1 failed");
      expect(formatted).toContain("3 searches, 1 failures");
      expect(formatted).toContain("✗ 2 failures");
      expect(formatted).toContain("BadServer: Timeout");
      expect(formatted).toContain("GoodServer (cutoff): Rate limited");
    });

    test("formats zero counts correctly", () => {
      const result = {
        success: true,
        detectionResults: {
          totalMissing: 0,
          totalCutoff: 0,
          successCount: 1,
          failureCount: 0,
        },
        searchResults: {
          missingTriggered: 0,
          cutoffTriggered: 0,
          successCount: 0,
          failureCount: 0,
        },
        totalSearches: 0,
        totalFailures: 0,
        errors: [],
      };

      const formatted = formatCycleResult(result);

      expect(formatted).toContain("0 missing, 0 cutoff");
      expect(formatted).toContain("0 searches");
      expect(formatted).toContain("✓ Success");
    });
  });

  describe("integration with search limits", () => {
    test("respects disabled search limits", async () => {
      testDb.setAppConfig({
        searchLimits: {
          missingLimit: 0,
          cutoffLimit: 0,
        },
      });

      const result = await runAutomationCycle(false);

      expect(result.searchResults.missingTriggered).toBe(0);
      expect(result.searchResults.cutoffTriggered).toBe(0);
    });

    test("uses configured search limits", async () => {
      testDb.setAppConfig({
        searchLimits: {
          missingLimit: 5,
          cutoffLimit: 3,
        },
      });

      const result = await runAutomationCycle(false);

      // With no servers, should trigger 0 searches
      expect(result.searchResults.missingTriggered).toBe(0);
      expect(result.searchResults.cutoffTriggered).toBe(0);
    });
  });

  describe("logging integration", () => {
    test("creates structured log entries", async () => {
      await runAutomationCycle(false);

      const logs = getRecentLogs(10);

      // Verify log structure
      for (const log of logs) {
        expect(log.id).toBeDefined();
        expect(log.timestamp).toBeInstanceOf(Date);
        expect(log.type).toBeDefined();
        expect(log.message).toBeDefined();
      }
    });

    test("maintains chronological order", async () => {
      await runAutomationCycle(false);

      const logs = getRecentLogs(10);

      // Logs should be in reverse chronological order (newest first)
      const cycleEndIndex = logs.findIndex((l) => l.type === "cycle_end");
      const cycleStartIndex = logs.findIndex((l) => l.type === "cycle_start");

      expect(cycleEndIndex).toBeLessThan(cycleStartIndex);
    });
  });
});
