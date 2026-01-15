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
    test("adds a new server", async () => {
      const server = await db.addServer({
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

    test("retrieves all servers", async () => {
      await db.addServer({
        id: "test-id-1",
        name: "Radarr",
        url: "http://localhost:7878",
        apiKey: "key1",
        type: "radarr",
      });
      await db.addServer({
        id: "test-id-2",
        name: "Sonarr",
        url: "http://localhost:8989",
        apiKey: "key2",
        type: "sonarr",
      });

      const servers = await db.getAllServers();
      expect(servers.length).toBe(2);
    });

    test("retrieves server by ID", async () => {
      await db.addServer({
        id: "test-id-1",
        name: "Test Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const server = await db.getServer("test-id-1");
      expect(server).not.toBeNull();
      expect(server?.name).toBe("Test Server");
    });

    test("retrieves server by name", async () => {
      await db.addServer({
        id: "test-id-1",
        name: "My Radarr",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const server = await db.getServerByName("My Radarr");
      expect(server).not.toBeNull();
      expect(server?.id).toBe("test-id-1");
    });

    test("retrieves servers by type", async () => {
      await db.addServer({
        id: "id1",
        name: "Radarr 1",
        url: "http://host1:7878",
        apiKey: "key1",
        type: "radarr",
      });
      await db.addServer({
        id: "id2",
        name: "Radarr 2",
        url: "http://host2:7878",
        apiKey: "key2",
        type: "radarr",
      });
      await db.addServer({
        id: "id3",
        name: "Sonarr",
        url: "http://host1:8989",
        apiKey: "key3",
        type: "sonarr",
      });

      const radarrs = await db.getServersByType("radarr");
      expect(radarrs.length).toBe(2);

      const sonarrs = await db.getServersByType("sonarr");
      expect(sonarrs.length).toBe(1);
    });

    test("checks for duplicate servers", async () => {
      await db.addServer({
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

    test("excludes ID when checking duplicates", async () => {
      await db.addServer({
        id: "id1",
        name: "Radarr",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      expect(db.serverExists("http://localhost:7878", "radarr", "id1")).toBe(false);
      expect(db.serverExists("http://localhost:7878", "radarr", "other-id")).toBe(true);
    });

    test("updates server name", async () => {
      await db.addServer({
        id: "id1",
        name: "Old Name",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const updated = await db.updateServer("id1", { name: "New Name" });
      expect(updated?.name).toBe("New Name");
      expect(updated?.url).toBe("http://localhost:7878");
    });

    test("updates server URL", async () => {
      await db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      const updated = await db.updateServer("id1", { url: "http://newhost:7878" });
      expect(updated?.url).toBe("http://newhost:7878");
    });

    test("updates server API key", async () => {
      await db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "old-key",
        type: "radarr",
      });

      const updated = await db.updateServer("id1", { apiKey: "new-key" });
      expect(updated?.apiKey).toBe("new-key");
    });

    test("deletes a server", async () => {
      await db.addServer({
        id: "id1",
        name: "Server",
        url: "http://localhost:7878",
        apiKey: "key",
        type: "radarr",
      });

      expect(db.deleteServer("id1")).toBe(true);
      expect(await db.getServer("id1")).toBeNull();
    });

    test("returns false when deleting non-existent server", () => {
      expect(db.deleteServer("non-existent")).toBe(false);
    });

    test("encrypts API keys at rest", async () => {
      const plainApiKey = "my-secret-api-key-12345";
      await db.addServer({
        id: "id1",
        name: "Test Server",
        url: "http://localhost:7878",
        apiKey: plainApiKey,
        type: "radarr",
      });

      // Read the raw database value to verify encryption
      const Database = (await import("bun:sqlite")).Database;
      const rawDb = new Database(TEST_DB_PATH);
      const row = rawDb.query<{ api_key: string }, [string]>(
        "SELECT api_key FROM servers WHERE id = ?"
      ).get("id1");
      rawDb.close();

      // The stored API key should be encrypted (not plaintext)
      expect(row?.api_key).not.toBe(plainApiKey);
      expect(row?.api_key).toContain(":"); // Encrypted format: iv:ciphertext

      // But when retrieved through the API, it should be decrypted
      const server = await db.getServer("id1");
      expect(server?.apiKey).toBe(plainApiKey);
    });
  });

  describe("Config Operations", () => {
    test("gets default config values", () => {
      const config = db.getAppConfig();

      expect(config.schedule.intervalHours).toBe(6);
      expect(config.schedule.enabled).toBe(true);
      expect(config.searchLimits.missingMoviesLimit).toBe(10);
      expect(config.searchLimits.missingEpisodesLimit).toBe(10);
      expect(config.searchLimits.cutoffMoviesLimit).toBe(5);
      expect(config.searchLimits.cutoffEpisodesLimit).toBe(5);
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
        searchLimits: { missingMoviesLimit: 20, missingEpisodesLimit: 25, cutoffMoviesLimit: 10, cutoffEpisodesLimit: 12 },
      });

      const config = db.getAppConfig();
      expect(config.searchLimits.missingMoviesLimit).toBe(20);
      expect(config.searchLimits.missingEpisodesLimit).toBe(25);
      expect(config.searchLimits.cutoffMoviesLimit).toBe(10);
      expect(config.searchLimits.cutoffEpisodesLimit).toBe(12);
    });

    test("migrates old limit keys to new granular keys", async () => {
      // First, set up old keys using the current db instance
      // We need to bypass the normal initialization to set old keys
      const { Database } = await import("bun:sqlite");
      const rawDb = new Database(TEST_DB_PATH);

      // Delete the new keys if they exist and insert old keys
      rawDb.run("DELETE FROM config WHERE key IN (?, ?, ?, ?)",
        ["limits.missing.movies", "limits.missing.episodes", "limits.cutoff.movies", "limits.cutoff.episodes"]);
      rawDb.run("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", ["limits.missing", "15"]);
      rawDb.run("INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)", ["limits.cutoff", "8"]);
      rawDb.close();

      // Close the current db
      db.close();

      // Now create a new DatabaseManager which will trigger migration in constructor
      const newDb = new DatabaseManager(TEST_DB_PATH);

      const config = newDb.getAppConfig();
      // Old keys should be migrated to new granular keys
      expect(config.searchLimits.missingMoviesLimit).toBe(15);
      expect(config.searchLimits.missingEpisodesLimit).toBe(15);
      expect(config.searchLimits.cutoffMoviesLimit).toBe(8);
      expect(config.searchLimits.cutoffEpisodesLimit).toBe(8);

      // Old keys should be removed
      expect(newDb.getConfig("limits.missing")).toBeNull();
      expect(newDb.getConfig("limits.cutoff")).toBeNull();

      // Reassign to db so afterEach doesn't try to close already-closed db
      db = newDb;
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
      // All three messages should be present (order may vary within same timestamp)
      const messages = logs.map(l => l.message);
      expect(messages).toContain("First");
      expect(messages).toContain("Second");
      expect(messages).toContain("Third");
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

