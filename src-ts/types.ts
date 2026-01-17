/**
 * Core type definitions for Janitarr
 */

/** Server type discriminator */
export type ServerType = "radarr" | "sonarr";

/** Configuration for a Radarr or Sonarr server */
export interface ServerConfig {
  id: string;
  name: string;
  url: string;
  apiKey: string;
  type: ServerType;
  createdAt: Date;
  updatedAt: Date;
}

/** Media item type */
export type MediaItemType = "movie" | "episode";

/** Represents a single media item for search targeting */
export interface MediaItem {
  id: number;
  title: string;
  type: MediaItemType;
}

/** Result of content detection for a single server */
export interface DetectionResult {
  serverId: string;
  serverName: string;
  serverType: ServerType;
  missingCount: number;
  cutoffCount: number;
  missingItems: MediaItem[];
  cutoffItems: MediaItem[];
  error?: string;
}

/** Log entry types */
export type LogEntryType = "cycle_start" | "cycle_end" | "search" | "error";

/** Category of search operation */
export type SearchCategory = "missing" | "cutoff";

/** Activity log entry */
export interface LogEntry {
  id: string;
  timestamp: Date;
  type: LogEntryType;
  serverName?: string;
  serverType?: ServerType;
  category?: SearchCategory;
  count?: number;
  message: string;
  isManual?: boolean;
}

/** Schedule configuration */
export interface ScheduleConfig {
  intervalHours: number;
  enabled: boolean;
}

/** Search limit configuration */
export interface SearchLimits {
  missingMoviesLimit: number;
  missingEpisodesLimit: number;
  cutoffMoviesLimit: number;
  cutoffEpisodesLimit: number;
}

/** Application configuration */
export interface AppConfig {
  schedule: ScheduleConfig;
  searchLimits: SearchLimits;
}
