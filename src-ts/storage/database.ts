/**
 * SQLite database layer using Bun's built-in SQLite support
 *
 * Handles persistent storage for server configurations, application settings,
 * and activity logs.
 */

import { Database } from "bun:sqlite";
import {
  ServerConfig,
  ServerType,
  LogEntry,
  LogEntryType,
  SearchCategory,
  AppConfig,
  ScheduleConfig,
  SearchLimits,
} from "../types";

/** Default database path */
const DEFAULT_DB_PATH = "./data/janitarr.db";

/** Log retention period in days */
const LOG_RETENTION_DAYS = 30;

/** Database row types */
interface ServerRow {
  id: string;
  name: string;
  url: string;
  api_key: string;
  type: string;
  created_at: string;
  updated_at: string;
}

interface ConfigRow {
  key: string;
  value: string;
}

interface LogRow {
  id: string;
  timestamp: string;
  type: string;
  server_name: string | null;
  server_type: string | null;
  category: string | null;
  count: number | null;
  message: string;
  is_manual: number;
}

/**
 * Database manager for Janitarr
 */
export class DatabaseManager {
  private db: Database;

  constructor(
    dbPath: string = process.env.JANITARR_DB_PATH ?? DEFAULT_DB_PATH,
  ) {
    // Ensure directory exists
    const dir = dbPath.substring(0, dbPath.lastIndexOf("/"));
    if (dir) {
      Bun.spawnSync(["mkdir", "-p", dir]);
    }

    this.db = new Database(dbPath, { create: true });
    this.initialize();
  }

  /**
   * Initialize database schema
   */
  private initialize(): void {
    this.db.exec(`
      CREATE TABLE IF NOT EXISTS servers (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        url TEXT NOT NULL,
        api_key TEXT NOT NULL,
        type TEXT NOT NULL CHECK(type IN ('radarr', 'sonarr')),
        created_at TEXT NOT NULL,
        updated_at TEXT NOT NULL,
        UNIQUE(url, type)
      );

      CREATE TABLE IF NOT EXISTS config (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL
      );

      CREATE TABLE IF NOT EXISTS logs (
        id TEXT PRIMARY KEY,
        timestamp TEXT NOT NULL,
        type TEXT NOT NULL,
        server_name TEXT,
        server_type TEXT,
        category TEXT,
        count INTEGER,
        message TEXT NOT NULL,
        is_manual INTEGER DEFAULT 0
      );

      CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
    `);

    // Migrate old limit keys to new granular keys (backward compatibility)
    this.migrateLimitKeys();

    // Set default config values if not present
    this.setConfigDefault("schedule.intervalHours", "6");
    this.setConfigDefault("schedule.enabled", "true");
    this.setConfigDefault("limits.missing.movies", "10");
    this.setConfigDefault("limits.missing.episodes", "10");
    this.setConfigDefault("limits.cutoff.movies", "5");
    this.setConfigDefault("limits.cutoff.episodes", "5");
  }

  /**
   * Set a config value only if not already set
   */
  private setConfigDefault(key: string, value: string): void {
    const existing = this.db
      .query<ConfigRow, [string]>("SELECT * FROM config WHERE key = ?")
      .get(key);

    if (!existing) {
      this.db.run("INSERT INTO config (key, value) VALUES (?, ?)", [
        key,
        value,
      ]);
    }
  }

  /**
   * Migrate old limit keys to new granular limit keys
   */
  private migrateLimitKeys(): void {
    // Check if old keys exist
    const oldMissingLimit = this.db
      .query<ConfigRow, [string]>("SELECT value FROM config WHERE key = ?")
      .get("limits.missing");

    const oldCutoffLimit = this.db
      .query<ConfigRow, [string]>("SELECT value FROM config WHERE key = ?")
      .get("limits.cutoff");

    // Migrate old missing limit to both movies and episodes
    if (oldMissingLimit) {
      this.db.run("INSERT OR IGNORE INTO config (key, value) VALUES (?, ?)", [
        "limits.missing.movies",
        oldMissingLimit.value,
      ]);
      this.db.run("INSERT OR IGNORE INTO config (key, value) VALUES (?, ?)", [
        "limits.missing.episodes",
        oldMissingLimit.value,
      ]);
      // Remove old key
      this.db.run("DELETE FROM config WHERE key = ?", ["limits.missing"]);
    }

    // Migrate old cutoff limit to both movies and episodes
    if (oldCutoffLimit) {
      this.db.run("INSERT OR IGNORE INTO config (key, value) VALUES (?, ?)", [
        "limits.cutoff.movies",
        oldCutoffLimit.value,
      ]);
      this.db.run("INSERT OR IGNORE INTO config (key, value) VALUES (?, ?)", [
        "limits.cutoff.episodes",
        oldCutoffLimit.value,
      ]);
      // Remove old key
      this.db.run("DELETE FROM config WHERE key = ?", ["limits.cutoff"]);
    }
  }

  /**
   * Close the database connection
   */
  close(): void {
    this.db.close();
  }

  /**
   * Test database connection
   */
  testConnection(): boolean {
    try {
      this.db.query<{ result: number }, []>("SELECT 1 as result").get();
      return true;
    } catch {
      return false;
    }
  }

  // ============== Server Operations ==============

  /**
   * Add a new server
   */
  async addServer(
    server: Omit<ServerConfig, "createdAt" | "updatedAt">,
  ): Promise<ServerConfig> {
    const { getEncryptionKey, encryptApiKey } = await import("../lib/crypto");
    const now = new Date().toISOString();
    const fullServer: ServerConfig = {
      ...server,
      createdAt: new Date(now),
      updatedAt: new Date(now),
    };

    const encryptionKey = await getEncryptionKey();
    const encryptedApiKey = await encryptApiKey(
      fullServer.apiKey,
      encryptionKey,
    );

    this.db.run(
      `INSERT INTO servers (id, name, url, api_key, type, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?)`,
      [
        fullServer.id,
        fullServer.name,
        fullServer.url,
        encryptedApiKey,
        fullServer.type,
        now,
        now,
      ],
    );

    return fullServer;
  }

  /**
   * Get all servers
   */
  async getAllServers(): Promise<ServerConfig[]> {
    const rows = this.db
      .query<ServerRow, []>("SELECT * FROM servers ORDER BY name")
      .all();
    return Promise.all(rows.map((row) => this.rowToServer(row)));
  }

  /**
   * Get all servers without decrypting API keys (for metrics)
   */
  listServers(): Array<{
    id: string;
    name: string;
    type: ServerType;
    enabled: boolean;
  }> {
    const rows = this.db
      .query<ServerRow, []>("SELECT * FROM servers ORDER BY name")
      .all();
    return rows.map((row) => ({
      id: row.id,
      name: row.name,
      type: row.type as ServerType,
      enabled: true, // All servers in DB are enabled for now
    }));
  }

  /**
   * Get a server by ID
   */
  async getServer(id: string): Promise<ServerConfig | null> {
    const row = this.db
      .query<ServerRow, [string]>("SELECT * FROM servers WHERE id = ?")
      .get(id);

    return row ? await this.rowToServer(row) : null;
  }

  /**
   * Get a server by name
   */
  async getServerByName(name: string): Promise<ServerConfig | null> {
    const row = this.db
      .query<ServerRow, [string]>("SELECT * FROM servers WHERE name = ?")
      .get(name);

    return row ? await this.rowToServer(row) : null;
  }

  /**
   * Get servers by type
   */
  async getServersByType(type: ServerType): Promise<ServerConfig[]> {
    const rows = this.db
      .query<
        ServerRow,
        [string]
      >("SELECT * FROM servers WHERE type = ? ORDER BY name")
      .all(type);

    return Promise.all(rows.map((row) => this.rowToServer(row)));
  }

  /**
   * Check if a server with the same URL and type already exists
   */
  serverExists(url: string, type: ServerType, excludeId?: string): boolean {
    const query = excludeId
      ? "SELECT 1 FROM servers WHERE url = ? AND type = ? AND id != ?"
      : "SELECT 1 FROM servers WHERE url = ? AND type = ?";

    const params = excludeId ? [url, type, excludeId] : [url, type];
    const row = this.db.query<{ 1: number }, string[]>(query).get(...params);

    return row !== null;
  }

  /**
   * Update a server
   */
  async updateServer(
    id: string,
    updates: Partial<Pick<ServerConfig, "name" | "url" | "apiKey">>,
  ): Promise<ServerConfig | null> {
    const { getEncryptionKey, encryptApiKey } = await import("../lib/crypto");
    const existing = await this.getServer(id); // Now async
    if (!existing) return null;

    const now = new Date().toISOString();
    let changesMade = false;

    if (updates.name !== undefined) {
      this.db.run("UPDATE servers SET name = ?, updated_at = ? WHERE id = ?", [
        updates.name,
        now,
        id,
      ]);
      changesMade = true;
    }

    if (updates.url !== undefined) {
      this.db.run("UPDATE servers SET url = ?, updated_at = ? WHERE id = ?", [
        updates.url,
        now,
        id,
      ]);
      changesMade = true;
    }

    if (updates.apiKey !== undefined) {
      const encryptionKey = await getEncryptionKey();
      const encryptedApiKey = await encryptApiKey(
        updates.apiKey,
        encryptionKey,
      );
      this.db.run(
        "UPDATE servers SET api_key = ?, updated_at = ? WHERE id = ?",
        [encryptedApiKey, now, id],
      );
      changesMade = true;
    }

    if (!changesMade) return existing; // No actual updates to trigger a fetch

    return this.getServer(id); // Now async
  }

  /**
   * Delete a server
   */
  deleteServer(id: string): boolean {
    const result = this.db.run("DELETE FROM servers WHERE id = ?", [id]);
    return result.changes > 0;
  }

  /**
   * Convert database row to ServerConfig
   */
  private async rowToServer(row: ServerRow): Promise<ServerConfig> {
    const { getEncryptionKey, decryptApiKey } = await import("../lib/crypto");
    const encryptionKey = await getEncryptionKey();
    const decryptedApiKey = await decryptApiKey(row.api_key, encryptionKey);
    return {
      id: row.id,
      name: row.name,
      url: row.url,
      apiKey: decryptedApiKey,
      type: row.type as ServerType,
      createdAt: new Date(row.created_at),
      updatedAt: new Date(row.updated_at),
    };
  }

  // ============== Config Operations ==============

  /**
   * Get a config value
   */
  getConfig(key: string): string | null {
    const row = this.db
      .query<ConfigRow, [string]>("SELECT value FROM config WHERE key = ?")
      .get(key);

    return row?.value ?? null;
  }

  /**
   * Set a config value
   */
  setConfig(key: string, value: string): void {
    this.db.run(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, [
      key,
      value,
    ]);
  }

  /**
   * Get full application config
   */
  getAppConfig(): AppConfig {
    return {
      schedule: {
        intervalHours: parseInt(
          this.getConfig("schedule.intervalHours") ?? "6",
          10,
        ),
        enabled: this.getConfig("schedule.enabled") === "true",
      },
      searchLimits: {
        missingMoviesLimit: parseInt(
          this.getConfig("limits.missing.movies") ?? "10",
          10,
        ),
        missingEpisodesLimit: parseInt(
          this.getConfig("limits.missing.episodes") ?? "10",
          10,
        ),
        cutoffMoviesLimit: parseInt(
          this.getConfig("limits.cutoff.movies") ?? "5",
          10,
        ),
        cutoffEpisodesLimit: parseInt(
          this.getConfig("limits.cutoff.episodes") ?? "5",
          10,
        ),
      },
    };
  }

  /**
   * Update application config
   */
  setAppConfig(config: {
    schedule?: Partial<ScheduleConfig>;
    searchLimits?: Partial<SearchLimits>;
  }): void {
    if (config.schedule) {
      if (config.schedule.intervalHours !== undefined) {
        this.setConfig(
          "schedule.intervalHours",
          config.schedule.intervalHours.toString(),
        );
      }
      if (config.schedule.enabled !== undefined) {
        this.setConfig("schedule.enabled", config.schedule.enabled.toString());
      }
    }

    if (config.searchLimits) {
      if (config.searchLimits.missingMoviesLimit !== undefined) {
        this.setConfig(
          "limits.missing.movies",
          config.searchLimits.missingMoviesLimit.toString(),
        );
      }
      if (config.searchLimits.missingEpisodesLimit !== undefined) {
        this.setConfig(
          "limits.missing.episodes",
          config.searchLimits.missingEpisodesLimit.toString(),
        );
      }
      if (config.searchLimits.cutoffMoviesLimit !== undefined) {
        this.setConfig(
          "limits.cutoff.movies",
          config.searchLimits.cutoffMoviesLimit.toString(),
        );
      }
      if (config.searchLimits.cutoffEpisodesLimit !== undefined) {
        this.setConfig(
          "limits.cutoff.episodes",
          config.searchLimits.cutoffEpisodesLimit.toString(),
        );
      }
    }
  }

  // ============== Log Operations ==============

  /**
   * Add a log entry
   */
  addLog(entry: Omit<LogEntry, "id" | "timestamp">): LogEntry {
    const id = crypto.randomUUID();
    const timestamp = new Date();

    this.db.run(
      `INSERT INTO logs (id, timestamp, type, server_name, server_type, category, count, message, is_manual)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        id,
        timestamp.toISOString(),
        entry.type,
        entry.serverName ?? null,
        entry.serverType ?? null,
        entry.category ?? null,
        entry.count ?? null,
        entry.message,
        entry.isManual ? 1 : 0,
      ],
    );

    return { id, timestamp, ...entry };
  }

  /**
   * Get recent log entries
   */
  getLogs(limit = 100, offset = 0): LogEntry[] {
    const rows = this.db
      .query<
        LogRow,
        [number, number]
      >("SELECT * FROM logs ORDER BY timestamp DESC LIMIT ? OFFSET ?")
      .all(limit, offset);

    return rows.map(this.rowToLog);
  }

  /**
   * Get log count
   */
  getLogCount(): number {
    const row = this.db
      .query<{ count: number }, []>("SELECT COUNT(*) as count FROM logs")
      .get();

    return row?.count ?? 0;
  }

  /**
   * Get filtered logs with pagination
   */
  getLogsPaginated(
    filters: {
      type?: LogEntryType;
      server?: string;
      startDate?: string;
      endDate?: string;
      search?: string;
    },
    limit = 100,
    offset = 0,
  ): LogEntry[] {
    let query = "SELECT * FROM logs WHERE 1=1";
    const params: (string | number)[] = [];

    if (filters.type) {
      query += " AND type = ?";
      params.push(filters.type);
    }

    if (filters.server) {
      query += " AND server_name = ?";
      params.push(filters.server);
    }

    if (filters.startDate) {
      query += " AND timestamp >= ?";
      params.push(filters.startDate);
    }

    if (filters.endDate) {
      query += " AND timestamp <= ?";
      params.push(filters.endDate);
    }

    if (filters.search) {
      query += " AND message LIKE ?";
      params.push(`%${filters.search}%`);
    }

    query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?";
    params.push(limit, offset);

    const rows = this.db
      .query<LogRow, (string | number)[]>(query)
      .all(...params);
    return rows.map(this.rowToLog);
  }

  /**
   * Get statistics for a specific server
   */
  getServerStats(serverId: string): {
    totalSearches: number;
    errorCount: number;
    lastCheckTime: string | null;
  } {
    // Get server name first
    const server = this.db
      .query<ServerRow, [string]>("SELECT name FROM servers WHERE id = ?")
      .get(serverId);

    if (!server) {
      return { totalSearches: 0, errorCount: 0, lastCheckTime: null };
    }

    // Count total searches for this server
    const searchCount = this.db
      .query<
        { count: number },
        [string]
      >("SELECT COUNT(*) as count FROM logs WHERE server_name = ? AND type = 'search'")
      .get(server.name);

    // Count errors for this server
    const errorCount = this.db
      .query<
        { count: number },
        [string]
      >("SELECT COUNT(*) as count FROM logs WHERE server_name = ? AND type = 'error'")
      .get(server.name);

    // Get last check time (most recent log entry)
    const lastLog = this.db
      .query<
        { timestamp: string },
        [string]
      >("SELECT timestamp FROM logs WHERE server_name = ? ORDER BY timestamp DESC LIMIT 1")
      .get(server.name);

    return {
      totalSearches: searchCount?.count ?? 0,
      errorCount: errorCount?.count ?? 0,
      lastCheckTime: lastLog?.timestamp ?? null,
    };
  }

  /**
   * Get system-wide statistics for dashboard
   */
  getSystemStats(): {
    totalServers: number;
    lastCycleTime: string | null;
    searchesLast24h: number;
    errorsLast24h: number;
  } {
    // Count total servers
    const serverCount = this.db
      .query<{ count: number }, []>("SELECT COUNT(*) as count FROM servers")
      .get();

    // Get last cycle end time
    const lastCycle = this.db
      .query<
        { timestamp: string },
        []
      >("SELECT timestamp FROM logs WHERE type = 'cycle_end' ORDER BY timestamp DESC LIMIT 1")
      .get();

    // Count searches in last 24 hours
    const yesterday = new Date();
    yesterday.setHours(yesterday.getHours() - 24);
    const searchCount = this.db
      .query<
        { count: number },
        [string]
      >("SELECT COUNT(*) as count FROM logs WHERE type = 'search' AND timestamp >= ?")
      .get(yesterday.toISOString());

    // Count errors in last 24 hours
    const errorCount = this.db
      .query<
        { count: number },
        [string]
      >("SELECT COUNT(*) as count FROM logs WHERE type = 'error' AND timestamp >= ?")
      .get(yesterday.toISOString());

    return {
      totalServers: serverCount?.count ?? 0,
      lastCycleTime: lastCycle?.timestamp ?? null,
      searchesLast24h: searchCount?.count ?? 0,
      errorsLast24h: errorCount?.count ?? 0,
    };
  }

  /**
   * Clear all logs
   */
  clearLogs(): number {
    const result = this.db.run("DELETE FROM logs");
    return result.changes;
  }

  /**
   * Purge logs older than retention period
   */
  purgeOldLogs(): number {
    const cutoff = new Date();
    cutoff.setDate(cutoff.getDate() - LOG_RETENTION_DAYS);

    const result = this.db.run("DELETE FROM logs WHERE timestamp < ?", [
      cutoff.toISOString(),
    ]);

    return result.changes;
  }

  /**
   * Convert database row to LogEntry
   */
  private rowToLog(row: LogRow): LogEntry {
    return {
      id: row.id,
      timestamp: new Date(row.timestamp),
      type: row.type as LogEntryType,
      serverName: row.server_name ?? undefined,
      serverType: row.server_type as ServerType | undefined,
      category: row.category as SearchCategory | undefined,
      count: row.count ?? undefined,
      message: row.message,
      isManual: row.is_manual === 1,
    };
  }
}

/** Singleton database instance */
let dbInstance: DatabaseManager | null = null;

/**
 * Get or create the database instance
 */
export function getDatabase(dbPath?: string): DatabaseManager {
  if (!dbInstance) {
    dbInstance = new DatabaseManager(dbPath);
  }
  return dbInstance;
}

/**
 * Close and clear the database instance
 */
export function closeDatabase(): void {
  if (dbInstance) {
    dbInstance.close();
    dbInstance = null;
  }
}
