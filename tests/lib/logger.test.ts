/**
 * Tests for activity logger
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { unlinkSync, existsSync } from "fs";
import { DatabaseManager, closeDatabase } from "../../src/storage/database";
import {
  logCycleStart,
  logCycleEnd,
  logSearches,
  logServerError,
  logSearchError,
  getRecentLogs,
  getLogCount,
  clearAllLogs,
  getLastCycleSummary,
  formatLogEntry,
  getLogTypeLabel,
} from "../../src/lib/logger";

const TEST_DB_PATH = "./data/test-logger.db";

let testDb: DatabaseManager;

describe("Activity Logger", () => {
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

  describe("logCycleStart", () => {
    test("logs scheduled cycle start", () => {
      const entry = logCycleStart(false);

      expect(entry.type).toBe("cycle_start");
      expect(entry.message).toContain("Scheduled");
      expect(entry.isManual).toBe(false);
    });

    test("logs manual cycle start", () => {
      const entry = logCycleStart(true);

      expect(entry.type).toBe("cycle_start");
      expect(entry.message).toContain("Manual");
      expect(entry.isManual).toBe(true);
    });
  });

  describe("logCycleEnd", () => {
    test("logs cycle end with search count", () => {
      const entry = logCycleEnd(10, 0, false);

      expect(entry.type).toBe("cycle_end");
      expect(entry.message).toContain("10 searches");
      expect(entry.count).toBe(10);
    });

    test("logs cycle end with failures", () => {
      const entry = logCycleEnd(8, 2, false);

      expect(entry.message).toContain("8 searches");
      expect(entry.message).toContain("2 failures");
    });

    test("marks manual cycles", () => {
      const entry = logCycleEnd(5, 0, true);

      expect(entry.isManual).toBe(true);
    });
  });

  describe("logSearches", () => {
    test("logs missing searches", () => {
      const entry = logSearches("Test Radarr", "radarr", "missing", 5, false);

      expect(entry.type).toBe("search");
      expect(entry.serverName).toBe("Test Radarr");
      expect(entry.serverType).toBe("radarr");
      expect(entry.category).toBe("missing");
      expect(entry.count).toBe(5);
      expect(entry.message).toContain("5 missing searches");
    });

    test("logs cutoff searches", () => {
      const entry = logSearches("Test Sonarr", "sonarr", "cutoff", 3, false);

      expect(entry.category).toBe("cutoff");
      expect(entry.message).toContain("3 cutoff searches");
    });
  });

  describe("logServerError", () => {
    test("logs server connection errors", () => {
      const entry = logServerError("Bad Server", "radarr", "Connection refused");

      expect(entry.type).toBe("error");
      expect(entry.serverName).toBe("Bad Server");
      expect(entry.message).toContain("Connection failed");
      expect(entry.message).toContain("Connection refused");
    });
  });

  describe("logSearchError", () => {
    test("logs search trigger errors", () => {
      const entry = logSearchError("Test Server", "sonarr", "missing", "API error");

      expect(entry.type).toBe("error");
      expect(entry.category).toBe("missing");
      expect(entry.message).toContain("Failed to trigger");
      expect(entry.message).toContain("API error");
    });
  });

  describe("getRecentLogs", () => {
    test("returns logs in reverse chronological order", () => {
      logCycleStart();
      logSearches("Server", "radarr", "missing", 5);
      logCycleEnd(5, 0);

      const logs = getRecentLogs(10);

      expect(logs.length).toBe(3);
      expect(logs[0].type).toBe("cycle_end");
    });

    test("respects limit parameter", () => {
      for (let i = 0; i < 10; i++) {
        logCycleStart();
      }

      const logs = getRecentLogs(5);
      expect(logs.length).toBe(5);
    });
  });

  describe("getLogCount", () => {
    test("returns total log count", () => {
      logCycleStart();
      logCycleEnd(0, 0);
      logCycleStart();

      expect(getLogCount()).toBe(3);
    });
  });

  describe("clearAllLogs", () => {
    test("clears all logs and returns count", () => {
      logCycleStart();
      logCycleEnd(0, 0);

      const cleared = clearAllLogs();

      expect(cleared).toBe(2);
      expect(getLogCount()).toBe(0);
    });
  });

  describe("getLastCycleSummary", () => {
    test("returns null summary when no cycles", () => {
      const summary = getLastCycleSummary();

      expect(summary.lastCycleTime).toBeNull();
      expect(summary.lastCycleSearches).toBe(0);
    });

    test("returns summary from last cycle", () => {
      logCycleStart(true);
      logSearches("Server", "radarr", "missing", 5);
      logCycleEnd(5, 0, true);

      const summary = getLastCycleSummary();

      expect(summary.lastCycleTime).not.toBeNull();
      expect(summary.lastCycleSearches).toBe(5);
      expect(summary.wasManual).toBe(true);
    });

    test("counts failures in cycle", () => {
      logCycleStart();
      logServerError("Bad Server", "radarr", "Error");
      logSearchError("Bad Server", "radarr", "missing", "Error");
      logCycleEnd(3, 2);

      const summary = getLastCycleSummary();

      expect(summary.lastCycleFailures).toBe(2);
    });
  });

  describe("formatLogEntry", () => {
    test("formats cycle start", () => {
      const entry = logCycleStart();
      const formatted = formatLogEntry(entry);

      expect(formatted).toContain("Cycle started");
    });

    test("formats cycle end", () => {
      const entry = logCycleEnd(10, 0);
      const formatted = formatLogEntry(entry);

      expect(formatted).toContain("10 searches");
    });

    test("formats errors", () => {
      const entry = logServerError("Server", "radarr", "Error");
      const formatted = formatLogEntry(entry);

      expect(formatted).toContain("[ERROR]");
    });

    test("marks manual entries", () => {
      const entry = logCycleStart(true);
      const formatted = formatLogEntry(entry);

      expect(formatted).toContain("[Manual]");
    });
  });

  describe("getLogTypeLabel", () => {
    test("returns correct labels", () => {
      expect(getLogTypeLabel("cycle_start")).toBe("Cycle Start");
      expect(getLogTypeLabel("cycle_end")).toBe("Cycle End");
      expect(getLogTypeLabel("search")).toBe("Search");
      expect(getLogTypeLabel("error")).toBe("Error");
    });
  });
});
