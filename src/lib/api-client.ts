/**
 * API client for Radarr and Sonarr servers
 *
 * Handles HTTP communication with media servers including authentication,
 * URL normalization, timeout handling, and error mapping.
 */

import type { ServerType, MediaItem } from "../types";

/** Default request timeout in milliseconds */
const DEFAULT_TIMEOUT_MS = 15000;

/** API version prefix */
const API_PREFIX = "/api/v3";

/** Result of an API operation */
export type ApiResult<T> =
  | { success: true; data: T }
  | { success: false; error: string };

/** System status response from Radarr/Sonarr */
export interface SystemStatus {
  appName: string;
  version: string;
  instanceName?: string;
}

/** Wanted/missing response item from Radarr */
export interface RadarrWantedRecord {
  id: number;
  title: string;
  monitored: boolean;
}

/** Wanted/missing response item from Sonarr */
export interface SonarrWantedRecord {
  id: number;
  series: { title: string };
  title: string;
  monitored: boolean;
  episodeNumber: number;
  seasonNumber: number;
}

/** Paginated response wrapper */
export interface PaginatedResponse<T> {
  page: number;
  pageSize: number;
  totalRecords: number;
  records: T[];
}

/** Command response from Radarr/Sonarr */
export interface CommandResponse {
  id: number;
  name: string;
  status: string;
}

/**
 * Normalize a URL by ensuring it has a protocol and removing trailing slashes
 */
export function normalizeUrl(url: string): string {
  let normalized = url.trim();

  // Add protocol if missing
  if (!normalized.match(/^https?:\/\//i)) {
    normalized = `http://${normalized}`;
  }

  // Remove trailing slashes
  normalized = normalized.replace(/\/+$/, "");

  return normalized;
}

/**
 * Validate URL format
 */
export function validateUrl(url: string): ApiResult<string> {
  const normalized = normalizeUrl(url);

  try {
    const parsed = new URL(normalized);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return { success: false, error: "URL must use http:// or https:// protocol" };
    }
    return { success: true, data: normalized };
  } catch {
    return { success: false, error: "Invalid URL format" };
  }
}

/**
 * Base API client for Radarr/Sonarr servers
 */
export class ApiClient {
  readonly baseUrl: string;
  readonly apiKey: string;
  readonly type: ServerType;
  readonly timeoutMs: number;

  constructor(
    url: string,
    apiKey: string,
    type: ServerType,
    timeoutMs = DEFAULT_TIMEOUT_MS
  ) {
    this.baseUrl = normalizeUrl(url);
    this.apiKey = apiKey;
    this.type = type;
    this.timeoutMs = timeoutMs;
  }

  /**
   * Make an authenticated request to the API
   */
  private async request<T>(
    method: string,
    endpoint: string,
    body?: unknown
  ): Promise<ApiResult<T>> {
    const url = `${this.baseUrl}${API_PREFIX}${endpoint}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeoutMs);

    try {
      const response = await fetch(url, {
        method,
        headers: {
          "X-Api-Key": this.apiKey,
          "Content-Type": "application/json",
        },
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (response.status === 401) {
        return { success: false, error: "Invalid API key - unauthorized" };
      }

      if (response.status === 404) {
        return { success: false, error: "API endpoint not found - check server URL" };
      }

      if (!response.ok) {
        return {
          success: false,
          error: `Server returned error: ${response.status} ${response.statusText}`,
        };
      }

      const data = (await response.json()) as T;
      return { success: true, data };
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error) {
        if (error.name === "AbortError") {
          return { success: false, error: `Request timed out after ${this.timeoutMs}ms` };
        }
        if (
          error.message.includes("fetch failed") ||
          error.message.includes("ECONNREFUSED") ||
          error.message.includes("Unable to connect")
        ) {
          return { success: false, error: "Server unreachable - check URL and network" };
        }
        return { success: false, error: `Network error: ${error.message}` };
      }

      return { success: false, error: "Unknown error occurred" };
    }
  }

  /**
   * Make a GET request
   */
  protected get<T>(endpoint: string): Promise<ApiResult<T>> {
    return this.request<T>("GET", endpoint);
  }

  /**
   * Make a POST request
   */
  protected post<T>(endpoint: string, body: unknown): Promise<ApiResult<T>> {
    return this.request<T>("POST", endpoint, body);
  }

  /**
   * Test connection to the server
   */
  async testConnection(): Promise<ApiResult<SystemStatus>> {
    return this.get<SystemStatus>("/system/status");
  }

  /**
   * Get paginated wanted/missing items
   */
  async getWantedMissing(
    page = 1,
    pageSize = 50
  ): Promise<ApiResult<PaginatedResponse<RadarrWantedRecord | SonarrWantedRecord>>> {
    return this.get(`/wanted/missing?page=${page}&pageSize=${pageSize}&sortKey=id&sortDirection=ascending`);
  }

  /**
   * Get paginated cutoff unmet items
   */
  async getCutoffUnmet(
    page = 1,
    pageSize = 50
  ): Promise<ApiResult<PaginatedResponse<RadarrWantedRecord | SonarrWantedRecord>>> {
    return this.get(`/wanted/cutoff?page=${page}&pageSize=${pageSize}&sortKey=id&sortDirection=ascending`);
  }
}

/**
 * Radarr-specific API client
 */
export class RadarrClient extends ApiClient {
  constructor(url: string, apiKey: string, timeoutMs = DEFAULT_TIMEOUT_MS) {
    super(url, apiKey, "radarr", timeoutMs);
  }

  /**
   * Trigger movie search for specific movie IDs
   */
  async searchMovies(movieIds: number[]): Promise<ApiResult<CommandResponse>> {
    return this.post<CommandResponse>("/command", {
      name: "MoviesSearch",
      movieIds,
    });
  }

  /**
   * Get all missing movies with pagination
   */
  async getAllMissing(): Promise<ApiResult<MediaItem[]>> {
    const items: MediaItem[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const result = await this.getWantedMissing(page, 100);
      if (!result.success) {
        return result;
      }

      const records = result.data.records as RadarrWantedRecord[];
      for (const record of records) {
        items.push({
          id: record.id,
          title: record.title,
          type: "movie",
        });
      }

      hasMore = items.length < result.data.totalRecords;
      page++;
    }

    return { success: true, data: items };
  }

  /**
   * Get all cutoff unmet movies with pagination
   */
  async getAllCutoffUnmet(): Promise<ApiResult<MediaItem[]>> {
    const items: MediaItem[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const result = await this.getCutoffUnmet(page, 100);
      if (!result.success) {
        return result;
      }

      const records = result.data.records as RadarrWantedRecord[];
      for (const record of records) {
        items.push({
          id: record.id,
          title: record.title,
          type: "movie",
        });
      }

      hasMore = items.length < result.data.totalRecords;
      page++;
    }

    return { success: true, data: items };
  }
}

/**
 * Sonarr-specific API client
 */
export class SonarrClient extends ApiClient {
  constructor(url: string, apiKey: string, timeoutMs = DEFAULT_TIMEOUT_MS) {
    super(url, apiKey, "sonarr", timeoutMs);
  }

  /**
   * Trigger episode search for specific episode IDs
   */
  async searchEpisodes(episodeIds: number[]): Promise<ApiResult<CommandResponse>> {
    return this.post<CommandResponse>("/command", {
      name: "EpisodeSearch",
      episodeIds,
    });
  }

  /**
   * Get all missing episodes with pagination
   */
  async getAllMissing(): Promise<ApiResult<MediaItem[]>> {
    const items: MediaItem[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const result = await this.getWantedMissing(page, 100);
      if (!result.success) {
        return result;
      }

      const records = result.data.records as SonarrWantedRecord[];
      for (const record of records) {
        items.push({
          id: record.id,
          title: `${record.series.title} - S${record.seasonNumber.toString().padStart(2, "0")}E${record.episodeNumber.toString().padStart(2, "0")} - ${record.title}`,
          type: "episode",
        });
      }

      hasMore = items.length < result.data.totalRecords;
      page++;
    }

    return { success: true, data: items };
  }

  /**
   * Get all cutoff unmet episodes with pagination
   */
  async getAllCutoffUnmet(): Promise<ApiResult<MediaItem[]>> {
    const items: MediaItem[] = [];
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      const result = await this.getCutoffUnmet(page, 100);
      if (!result.success) {
        return result;
      }

      const records = result.data.records as SonarrWantedRecord[];
      for (const record of records) {
        items.push({
          id: record.id,
          title: `${record.series.title} - S${record.seasonNumber.toString().padStart(2, "0")}E${record.episodeNumber.toString().padStart(2, "0")} - ${record.title}`,
          type: "episode",
        });
      }

      hasMore = items.length < result.data.totalRecords;
      page++;
    }

    return { success: true, data: items };
  }
}

/**
 * Create the appropriate client based on server type
 */
export function createClient(
  url: string,
  apiKey: string,
  type: ServerType,
  timeoutMs = DEFAULT_TIMEOUT_MS
): RadarrClient | SonarrClient {
  return type === "radarr"
    ? new RadarrClient(url, apiKey, timeoutMs)
    : new SonarrClient(url, apiKey, timeoutMs);
}
