/**
 * SQLite database layer using Bun's built-in SQLite support
 *
 * Handles persistent storage for server configurations, application settings,
 * and activity logs.
 */

import { Database } from "bun:sqlite";
import type {
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

  constructor(dbPath: string = process.env.JANITARR_DB_PATH ?? DEFAULT_DB_PATH) {
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

    // Set default config values if not present
    this.setConfigDefault("schedule.intervalHours", "6");
    this.setConfigDefault("schedule.enabled", "true");
    this.setConfigDefault("limits.missing", "10");
    this.setConfigDefault("limits.cutoff", "5");
  }

  /**
   * Set a config value only if not already set
   */
  private setConfigDefault(key: string, value: string): void {
    const existing = this.db.query<ConfigRow, [string]>(
      "SELECT * FROM config WHERE key = ?"
    ).get(key);

    if (!existing) {
      this.db.run("INSERT INTO config (key, value) VALUES (?, ?)", [key, value]);
    }
  }

  /**
   * Close the database connection
   */
  close(): void {
    this.db.close();
  }

  // ============== Server Operations ==============

  /**
   * Add a new server
   */
  addServer(server: Omit<ServerConfig, "createdAt" | "updatedAt">): ServerConfig {
    const now = new Date().toISOString();
    const fullServer: ServerConfig = {
      ...server,
      createdAt: new Date(now),
      updatedAt: new Date(now),
    };

    this.db.run(
      `INSERT INTO servers (id, name, url, api_key, type, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?, ?)`,
      [
        fullServer.id,
        fullServer.name,
        fullServer.url,
        fullServer.apiKey,
        fullServer.type,
        now,
        now,
      ]
    );

    return fullServer;
  }

  /**
   * Get all servers
   */
  getAllServers(): ServerConfig[] {
    const rows = this.db.query<ServerRow, []>("SELECT * FROM servers ORDER BY name").all();
    return rows.map(this.rowToServer);
  }

  /**
   * Get a server by ID
   */
  getServer(id: string): ServerConfig | null {
    const row = this.db.query<ServerRow, [string]>(
      "SELECT * FROM servers WHERE id = ?"
    ).get(id);

    return row ? this.rowToServer(row) : null;
  }

  /**
   * Get a server by name
   */
  getServerByName(name: string): ServerConfig | null {
    const row = this.db.query<ServerRow, [string]>(
      "SELECT * FROM servers WHERE name = ?"
    ).get(name);

    return row ? this.rowToServer(row) : null;
  }

  /**
   * Get servers by type
   */
  getServersByType(type: ServerType): ServerConfig[] {
    const rows = this.db.query<ServerRow, [string]>(
      "SELECT * FROM servers WHERE type = ? ORDER BY name"
    ).all(type);

    return rows.map(this.rowToServer);
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
  updateServer(
    id: string,
    updates: Partial<Pick<ServerConfig, "name" | "url" | "apiKey">>
  ): ServerConfig | null {
    const existing = this.getServer(id);
    if (!existing) return null;

    const now = new Date().toISOString();

    if (updates.name !== undefined) {
      this.db.run("UPDATE servers SET name = ?, updated_at = ? WHERE id = ?", [
        updates.name,
        now,
        id,
      ]);
    }

    if (updates.url !== undefined) {
      this.db.run("UPDATE servers SET url = ?, updated_at = ? WHERE id = ?", [
        updates.url,
        now,
        id,
      ]);
    }

    if (updates.apiKey !== undefined) {
      this.db.run("UPDATE servers SET api_key = ?, updated_at = ? WHERE id = ?", [
        updates.apiKey,
        now,
        id,
      ]);
    }

    return this.getServer(id);
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
  private rowToServer(row: ServerRow): ServerConfig {
    return {
      id: row.id,
      name: row.name,
      url: row.url,
      apiKey: row.api_key,
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
    const row = this.db.query<ConfigRow, [string]>(
      "SELECT value FROM config WHERE key = ?"
    ).get(key);

    return row?.value ?? null;
  }

  /**
   * Set a config value
   */
  setConfig(key: string, value: string): void {
    this.db.run(
      `INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`,
      [key, value]
    );
  }

  /**
   * Get full application config
   */
  getAppConfig(): AppConfig {
    return {
      schedule: {
        intervalHours: parseInt(this.getConfig("schedule.intervalHours") ?? "6", 10),
        enabled: this.getConfig("schedule.enabled") === "true",
      },
      searchLimits: {
        missingLimit: parseInt(this.getConfig("limits.missing") ?? "10", 10),
        cutoffLimit: parseInt(this.getConfig("limits.cutoff") ?? "5", 10),
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
        this.setConfig("schedule.intervalHours", config.schedule.intervalHours.toString());
      }
      if (config.schedule.enabled !== undefined) {
        this.setConfig("schedule.enabled", config.schedule.enabled.toString());
      }
    }

    if (config.searchLimits) {
      if (config.searchLimits.missingLimit !== undefined) {
        this.setConfig("limits.missing", config.searchLimits.missingLimit.toString());
      }
      if (config.searchLimits.cutoffLimit !== undefined) {
        this.setConfig("limits.cutoff", config.searchLimits.cutoffLimit.toString());
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
      ]
    );

    return { id, timestamp, ...entry };
  }

  /**
   * Get recent log entries
   */
  getLogs(limit = 100, offset = 0): LogEntry[] {
    const rows = this.db.query<LogRow, [number, number]>(
      "SELECT * FROM logs ORDER BY timestamp DESC LIMIT ? OFFSET ?"
    ).all(limit, offset);

    return rows.map(this.rowToLog);
  }

  /**
   * Get log count
   */
  getLogCount(): number {
    const row = this.db.query<{ count: number }, []>(
      "SELECT COUNT(*) as count FROM logs"
    ).get();

    return row?.count ?? 0;
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

    const result = this.db.run(
      "DELETE FROM logs WHERE timestamp < ?",
      [cutoff.toISOString()]
    );

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
