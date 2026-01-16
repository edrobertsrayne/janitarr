/**
 * Shared type definitions for the frontend application
 */

// ========== Core Types (from backend) ==========

export type ServerType = "radarr" | "sonarr";

export interface ServerConfig {
  id: string;
  name: string;
  url: string;
  apiKey: string;
  type: ServerType;
  enabled?: boolean;
  createdAt: string;
  updatedAt: string;
}

export type MediaItemType = "movie" | "episode";

export interface MediaItem {
  id: number;
  title: string;
  type: MediaItemType;
}

export type LogEntryType = "cycle_start" | "cycle_end" | "search" | "error";
export type SearchCategory = "missing" | "cutoff";

export interface LogEntry {
  id: string;
  timestamp: string;
  type: LogEntryType;
  serverName?: string;
  serverType?: ServerType;
  category?: SearchCategory;
  count?: number;
  message: string;
  isManual?: boolean;
}

export interface ScheduleConfig {
  intervalHours: number;
  enabled: boolean;
}

export interface SearchLimits {
  missingMoviesLimit: number;
  missingEpisodesLimit: number;
  cutoffMoviesLimit: number;
  cutoffEpisodesLimit: number;
}

export interface AppConfig {
  schedule: ScheduleConfig;
  searchLimits: SearchLimits;
}

// ========== API Request/Response Types ==========

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface CreateServerRequest {
  name: string;
  type: ServerType;
  url: string;
  apiKey: string;
  enabled?: boolean;
}

export interface UpdateServerRequest {
  name?: string;
  url?: string;
  apiKey?: string;
  enabled?: boolean;
}

export interface ServerTestResponse {
  success: boolean;
  message: string;
  status?: {
    version: string;
    appName: string;
  };
}

export interface ServerStatsResponse {
  totalSearches: number;
  successRate: number;
  lastCheckTime: string | null;
  errorCount: number;
}

export interface LogQueryParams {
  limit?: number;
  offset?: number;
  type?: LogEntryType;
  server?: string;
  startDate?: string;
  endDate?: string;
  search?: string;
}

export interface LogDeleteResponse {
  deletedCount: number;
}

export interface TriggerAutomationRequest {
  type?: "full" | "missing" | "cutoff";
}

export interface TriggerAutomationResponse {
  jobId: string;
  message: string;
}

export interface AutomationStatusResponse {
  running: boolean;
  nextScheduledRun: string | null;
  lastRunTime: string | null;
  lastRunResults?: {
    searchesTriggered: number;
    failedServers: number;
  };
}

export interface StatsSummaryResponse {
  totalServers: number;
  activeServers: number;
  lastCycleTime: string | null;
  nextScheduledTime: string | null;
  searchesLast24h: number;
  errorsLast24h: number;
}

// ========== WebSocket Types ==========

export interface WebSocketFilters {
  types?: LogEntryType[];
  servers?: string[];
}

export type WSClientMessage =
  | { type: "subscribe"; filters?: WebSocketFilters }
  | { type: "unsubscribe" }
  | { type: "ping" };

export type WSServerMessage =
  | { type: "log"; data: LogEntry }
  | { type: "connected"; message: string }
  | { type: "pong" };

// ========== Frontend-specific Types ==========

export type ThemeMode = "light" | "dark" | "system";

export interface ViewOptions {
  serversView: "list" | "card";
}
