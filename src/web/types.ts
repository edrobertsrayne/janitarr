/**
 * Type definitions for Web API requests and responses
 */

import type {
  ServerType,
  LogEntry,
  LogEntryType,
} from "../types";

/** HTTP methods supported by the API */
export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

/** Generic API response wrapper */
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
}

/** API request context */
export interface RequestContext {
  method: HttpMethod;
  path: string;
  query: Record<string, string>;
  body?: unknown;
}

// ========== Server Endpoints ==========

/** Request body for creating a server */
export interface CreateServerRequest {
  name: string;
  type: ServerType;
  url: string;
  apiKey: string;
  enabled?: boolean;
}

/** Request body for updating a server */
export interface UpdateServerRequest {
  name?: string;
  url?: string;
  apiKey?: string;
  enabled?: boolean;
}

/** Response from server test endpoint */
export interface ServerTestResponse {
  success: boolean;
  message: string;
  status?: {
    version: string;
    appName: string;
  };
}

/** Server statistics response */
export interface ServerStatsResponse {
  totalSearches: number;
  successRate: number;
  lastCheckTime: string | null;
  errorCount: number;
}

// ========== Log Endpoints ==========

/** Query parameters for log listing */
export interface LogQueryParams {
  limit?: number;
  offset?: number;
  type?: LogEntryType;
  server?: string;
  startDate?: string;
  endDate?: string;
  search?: string;
}

/** Response from log deletion */
export interface LogDeleteResponse {
  deletedCount: number;
}

// ========== Automation Endpoints ==========

/** Request body for triggering automation */
export interface TriggerAutomationRequest {
  type?: "full" | "missing" | "cutoff";
}

/** Response from automation trigger */
export interface TriggerAutomationResponse {
  jobId: string;
  message: string;
}

/** Automation status response */
export interface AutomationStatusResponse {
  running: boolean;
  nextScheduledRun: string | null;
  lastRunTime: string | null;
  lastRunResults?: {
    searchesTriggered: number;
    failedServers: number;
  };
}

// ========== Statistics Endpoints ==========

/** Summary statistics for dashboard */
export interface StatsSummaryResponse {
  totalServers: number;
  activeServers: number;
  lastCycleTime: string | null;
  nextScheduledTime: string | null;
  searchesLast24h: number;
  errorsLast24h: number;
}

// ========== WebSocket Messages ==========

/** WebSocket subscription filters */
export interface WebSocketFilters {
  types?: LogEntryType[];
  servers?: string[];
}

/** Client-to-server WebSocket message types */
export type WSClientMessage =
  | { type: "subscribe"; filters?: WebSocketFilters }
  | { type: "unsubscribe" }
  | { type: "ping" };

/** Server-to-client WebSocket message types */
export type WSServerMessage =
  | { type: "log"; data: LogEntry }
  | { type: "connected"; message: string }
  | { type: "pong" };

// ========== Utility Types ==========

/** HTTP status codes */
export const HttpStatus = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500,
} as const;

/** Helper to create successful JSON response */
export function jsonSuccess<T>(data: T, status: number = HttpStatus.OK): Response {
  return new Response(JSON.stringify({ success: true, data }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

/** Helper to create error JSON response */
export function jsonError(error: string, status: number = HttpStatus.BAD_REQUEST): Response {
  return new Response(JSON.stringify({ success: false, error }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

/** Helper to parse JSON body safely */
export async function parseJsonBody<T>(req: Request): Promise<T | null> {
  try {
    return await req.json();
  } catch {
    return null;
  }
}

/** Helper to parse query parameters from URL */
export function parseQueryParams(url: URL): Record<string, string> {
  const params: Record<string, string> = {};
  url.searchParams.forEach((value, key) => {
    params[key] = value;
  });
  return params;
}

/** Helper to extract path parameters */
export function extractPathParam(path: string, pattern: RegExp, groupIndex: number = 1): string | null {
  const match = path.match(pattern);
  return match?.[groupIndex] ?? null;
}
