/**
 * REST API client for Janitarr backend
 */

import type {
  ApiResponse,
  ServerConfig,
  CreateServerRequest,
  UpdateServerRequest,
  ServerTestResponse,
  ServerStatsResponse,
  LogEntry,
  LogQueryParams,
  LogDeleteResponse,
  AppConfig,
  TriggerAutomationRequest,
  TriggerAutomationResponse,
  AutomationStatusResponse,
  StatsSummaryResponse,
} from '../types';

const API_BASE = '/api';

/**
 * Generic fetch wrapper with error handling
 */
async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<ApiResponse<T>> {
  try {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    const data = await response.json();

    if (!response.ok) {
      return {
        success: false,
        error: data.error || `Request failed with status ${response.status}`,
      };
    }

    return data as ApiResponse<T>;
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Network error',
    };
  }
}

// ========== Configuration API ==========

export async function getConfig(): Promise<ApiResponse<AppConfig>> {
  return apiFetch<AppConfig>('/config');
}

export async function updateConfig(
  config: Partial<AppConfig>
): Promise<ApiResponse<AppConfig>> {
  return apiFetch<AppConfig>('/config', {
    method: 'PATCH',
    body: JSON.stringify(config),
  });
}

export async function resetConfig(): Promise<ApiResponse<AppConfig>> {
  return apiFetch<AppConfig>('/config/reset', {
    method: 'PUT',
  });
}

// ========== Servers API ==========

export async function getServers(): Promise<ApiResponse<ServerConfig[]>> {
  return apiFetch<ServerConfig[]>('/servers');
}

export async function getServer(id: string): Promise<ApiResponse<ServerConfig>> {
  return apiFetch<ServerConfig>(`/servers/${id}`);
}

export async function createServer(
  server: CreateServerRequest
): Promise<ApiResponse<ServerConfig>> {
  return apiFetch<ServerConfig>('/servers', {
    method: 'POST',
    body: JSON.stringify(server),
  });
}

export async function updateServer(
  id: string,
  updates: UpdateServerRequest
): Promise<ApiResponse<ServerConfig>> {
  return apiFetch<ServerConfig>(`/servers/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  });
}

export async function deleteServer(id: string): Promise<ApiResponse<void>> {
  return apiFetch<void>(`/servers/${id}`, {
    method: 'DELETE',
  });
}

export async function testServer(
  server: CreateServerRequest
): Promise<ApiResponse<ServerTestResponse>> {
  return apiFetch<ServerTestResponse>(`/servers/test`, {
    method: 'POST',
    body: JSON.stringify(server),
  });
}

export async function testServerConnectionById(
  id: string
): Promise<ApiResponse<ServerTestResponse>> {
  return apiFetch<ServerTestResponse>(`/servers/${id}/test`, {
    method: 'POST',
  });
}

export async function getServerStats(
  id: string
): Promise<ApiResponse<ServerStatsResponse>> {
  return apiFetch<ServerStatsResponse>(`/stats/servers/${id}`);
}

// ========== Logs API ==========

export async function getLogs(
  params?: LogQueryParams
): Promise<ApiResponse<LogEntry[]>> {
  const queryString = params
    ? '?' + new URLSearchParams(params as Record<string, string>).toString()
    : '';
  return apiFetch<LogEntry[]>(`/logs${queryString}`);
}

export async function deleteLogs(): Promise<ApiResponse<LogDeleteResponse>> {
  return apiFetch<LogDeleteResponse>('/logs', {
    method: 'DELETE',
  });
}

export async function exportLogs(
  params?: LogQueryParams,
  format: 'json' | 'csv' = 'json'
): Promise<Blob | null> {
  try {
    const queryString = params
      ? '?' +
        new URLSearchParams({
          ...(params as Record<string, string>),
          format,
        }).toString()
      : `?format=${format}`;

    const response = await fetch(`${API_BASE}/logs/export${queryString}`);

    if (!response.ok) {
      return null;
    }

    return await response.blob();
  } catch {
    return null;
  }
}

// ========== Automation API ==========

export async function triggerAutomation(
  request?: TriggerAutomationRequest
): Promise<ApiResponse<TriggerAutomationResponse>> {
  return apiFetch<TriggerAutomationResponse>('/automation/trigger', {
    method: 'POST',
    body: request ? JSON.stringify(request) : undefined,
  });
}

export async function getAutomationStatus(): Promise<
  ApiResponse<AutomationStatusResponse>
> {
  return apiFetch<AutomationStatusResponse>('/automation/status');
}

// ========== Statistics API ==========

export async function getStatsSummary(): Promise<
  ApiResponse<StatsSummaryResponse>
> {
  return apiFetch<StatsSummaryResponse>('/stats/summary');
}
