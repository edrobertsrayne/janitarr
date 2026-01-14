/**
 * Tests for SQLite database layer
 */

import { describe, test, expect, beforeEach, afterEach } from "bun:test";
import { DatabaseManager } from "../../src/storage/database";
import { unlinkSync, existsSync } from "fs";

const TEST_DB_PATH = "./data/test-janitarr.db";

describe("DatabaseManager", () => {
  let db: DatabaseManager;

  beforeEach(() => {
    // Clean up any existing test database
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
    db = new DatabaseManager(TEST_DB_PATH);
  });

  afterEach(() => {
    db.close();
    if (existsSync(TEST_DB_PATH)) {
      unlinkSync(TEST_DB_PATH);
    }
  });

  describe("Server Operations", () => {
    test("adds a new server", () => {
      const server = db.addServer({
        id: "test-id-1",
        name: "Test Radarr",
        url: "http://localhost:7878",
        apiKey: "test-api-key",
        type: "radarr",
      });

      expect(server.id).toBe("test-id-1");
      expect(server.name).toBe("Test Radarr");
      expect(server.url).toBe("http://localhost:7878");
      expect(server.type).toBe("radarr");
      expect(server.createdAt).toBeInstanceOf(Date);
      expect(server.updatedAt).toBeInstanceOf(Date);
    });

    test("retrieves all servers", () => {
      db.addServer({
        id: "test-id-1",
        name: "Radarr",
        url: "http://localhost:7878",
        apiKey: "key1",
        type: "radarr",
      });
      db.addServer({
        id: "test-id-2",
        name: "Sonarr",
        url: "http://localhost:8989",
        apiKey: "key2",
        type: "sonarr",
      });

      const servers = db.getAllServers();
      expect(servers.length).toBe(2);
    });

    test("retrieves server by ID", () => {
      db.addServer({
        id: "test-id-1",
        name: "Test Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const server = db.getServer("test-id-1");
      expect(server).not.toBeNull();
      expect(server?.name).toBe("Test Server");
    });

    test("retrieves server by name", () => {
      db.addServer({
        id: "test-id-1",
        name: "My Radarr",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const server = db.getServerByName("My Radarr");
      expect(server).not.toBeNull();
      expect(server?.id).toBe("test-id-1");
    });

    test("retrieves servers by type", () => {
      db.addServer({
        id: "id1",
        name: "Radarr 1",
        url: "http://host1:7878",
        apiKey: "key1",
        type: "radarr",
      });
      db.addServer({
        id: "id2",
        name: "Radarr 2",
        url: "http://host2:7878",
        apiKey: "key2",
        type: "radarr",
      });
      db.addServer({
        id: "id3",
        name: "Sonarr",
        url: "http://host1:8989",
        apiKey: "key3",
        type: "sonarr",
      });

      const radarrs = db.getServersByType("radarr");
      expect(radarrs.length).toBe(2);

      const sonarrs = db.getServersByType("sonarr");
      expect(sonarrs.length).toBe(1);
    });

    test("checks for duplicate servers", () => {
      db.addServer({
        id: "id1",
        name: "Radarr",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      expect(db.serverExists("http://localhost:7878", "radarr")).toBe(true);
      expect(db.serverExists("http://localhost:7878", "sonarr")).toBe(false);
      expect(db.serverExists("http://other:7878", "radarr")).toBe(false);
    });

    test("excludes ID when checking duplicates", () => {
      db.addServer({
        id: "id1",
        name: "Radarr",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      expect(db.serverExists("http://localhost:7878", "radarr", "id1")).toBe(false);
      expect(db.serverExists("http://localhost:7878", "radarr", "other-id")).toBe(true);
    });

    test("updates server name", () => {
      db.addServer({
        id: "id1",
        name: "Old Name",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const updated = db.updateServer("id1", { name: "New Name" });
      expect(updated?.name).toBe("New Name");
      expect(updated?.url).toBe("http://localhost:7878");
    });

    test("updates server URL", () => {
      db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const updated = db.updateServer("id1", { url: "http://newhost:7878" });
      expect(updated?.url).toBe("http://newhost:7878");
    });

    test("updates server API key", () => {
      db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "old-key",
        type: "radarr",
      });

      const updated = db.updateServer("id1", { apiKey: "new-key" });
      expect(updated?.apiKey).toBe("new-key");
    });

    test("deletes a server", () => {
      db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      expect(db.deleteServer("id1")).toBe(true);
      expect(db.getServer("id1")).toBeNull();
    });

    test("returns false when deleting non-existent server", () => {
      expect(db.deleteServer("non-existent")).toBe(false);
    });
  });

  describe("Config Operations", () => {
    test("gets default config values", () => {
      const config = db.getAppConfig();

      expect(config.schedule.intervalHours).toBe(6);
      expect(config.schedule.enabled).toBe(true);
      expect(config.searchLimits.missingLimit).toBe(10);
      expect(config.searchLimits.cutoffLimit).toBe(5);
    });

    test("sets and retrieves config values", () => {
      db.setConfig("schedule.intervalHours", "12");
      expect(db.getConfig("schedule.intervalHours")).toBe("12");
    });

    test("updates app config partially", () => {
      db.setAppConfig({
        schedule: { intervalHours: 24 },
      });

      const config = db.getAppConfig();
      expect(config.schedule.intervalHours).toBe(24);
      expect(config.schedule.enabled).toBe(true); // Unchanged
    });

    test("updates search limits", () => {
      db.setAppConfig({
        searchLimits: { missingLimit: 20, cutoffLimit: 10 },
      });

      const config = db.getAppConfig();
      expect(config.searchLimits.missingLimit).toBe(20);
      expect(config.searchLimits.cutoffLimit).toBe(10);
    });
  });

  describe("Log Operations", () => {
    test("adds log entries", () => {
      const entry = db.addLog({
        type: "cycle_start",
        message: "Starting automation cycle",
        isManual: false,
      });

      expect(entry.id).toBeDefined();
      expect(entry.timestamp).toBeInstanceOf(Date);
      expect(entry.message).toBe("Starting automation cycle");
    });

    test("retrieves logs in reverse chronological order", () => {
      db.addLog({ type: "cycle_start", message: "First" });
      db.addLog({ type: "cycle_end", message: "Second" });
      db.addLog({ type: "search", message: "Third" });

      const logs = db.getLogs(10);
      expect(logs.length).toBe(3);
      expect(logs[0].message).toBe("Third");
      expect(logs[2].message).toBe("First");
    });

    test("respects limit parameter", () => {
      for (let i = 0; i < 10; i++) {
        db.addLog({ type: "search", message: `Log ${i}` });
      }

      const logs = db.getLogs(5);
      expect(logs.length).toBe(5);
    });

    test("respects offset parameter", () => {
      for (let i = 0; i < 10; i++) {
        db.addLog({ type: "search", message: `Log ${i}` });
      }

      const logs = db.getLogs(5, 5);
      expect(logs.length).toBe(5);
      expect(logs[0].message).toBe("Log 4");
    });

    test("counts total logs", () => {
      for (let i = 0; i < 5; i++) {
        db.addLog({ type: "search", message: `Log ${i}` });
      }

      expect(db.getLogCount()).toBe(5);
    });

    test("clears all logs", () => {
      for (let i = 0; i < 5; i++) {
        db.addLog({ type: "search", message: `Log ${i}` });
      }

      const cleared = db.clearLogs();
      expect(cleared).toBe(5);
      expect(db.getLogCount()).toBe(0);
    });

    test("stores log entry with all fields", () => {
      db.addLog({
        type: "search",
        serverName: "My Radarr",
        serverType: "radarr",
        category: "missing",
        count: 5,
        message: "Triggered 5 searches",
        isManual: true,
      });

      const logs = db.getLogs(1);
      expect(logs[0].serverName).toBe("My Radarr");
      expect(logs[0].serverType).toBe("radarr");
      expect(logs[0].category).toBe("missing");
      expect(logs[0].count).toBe(5);
      expect(logs[0].isManual).toBe(true);
    });
  });
});
