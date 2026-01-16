import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  getConfig,
  updateConfig,
  resetConfig,
  getServers,
  createServer,
  updateServer,
  deleteServer,
  testServer,
  getLogs,
  deleteLogs,
  getStatsSummary,
  triggerAutomation,
  getAutomationStatus,
} from './api';

// Mock global fetch
globalThis.fetch = vi.fn();

describe('API Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Configuration API', () => {
    it('getConfig - fetches configuration successfully', async () => {
      const mockConfig = {
        schedule: { enabled: true, intervalHours: 6 },
        limits: {
          missing: { movies: 10, episodes: 10 },
          cutoff: { movies: 5, episodes: 5 },
        },
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockConfig }),
      } as Response);

      const result = await getConfig();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockConfig);
      expect(fetch).toHaveBeenCalledWith('/api/config', expect.any(Object));
    });

    it('updateConfig - updates configuration successfully', async () => {
      const updates = { schedule: { enabled: false, intervalHours: 12 } };
      const mockResponse = {
        schedule: { enabled: false, intervalHours: 12 },
        limits: {
          missing: { movies: 10, episodes: 10 },
          cutoff: { movies: 5, episodes: 5 },
        },
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await updateConfig(updates);

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith(
        '/api/config',
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify(updates),
        })
      );
    });

    it('resetConfig - resets configuration successfully', async () => {
      const mockConfig = {
        schedule: { enabled: true, intervalHours: 6 },
        limits: {
          missing: { movies: 10, episodes: 10 },
          cutoff: { movies: 5, episodes: 5 },
        },
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockConfig }),
      } as Response);

      const result = await resetConfig();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockConfig);
      expect(fetch).toHaveBeenCalledWith(
        '/api/config/reset',
        expect.objectContaining({ method: 'PUT' })
      );
    });
  });

  describe('Servers API', () => {
    it('getServers - fetches servers successfully', async () => {
      const mockServers = [
        {
          id: '1',
          name: 'Radarr',
          type: 'radarr',
          url: 'http://localhost:7878',
          apiKey: 'test-key',
          enabled: true,
        },
      ];

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockServers }),
      } as Response);

      const result = await getServers();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockServers);
      expect(fetch).toHaveBeenCalledWith('/api/servers', expect.any(Object));
    });

    it('createServer - creates server successfully', async () => {
      const newServer = {
        name: 'Radarr',
        type: 'radarr' as const,
        url: 'http://localhost:7878',
        apiKey: 'test-key',
      };

      const mockResponse = {
        id: '1',
        ...newServer,
        enabled: true,
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await createServer(newServer);

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith(
        '/api/servers',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(newServer),
        })
      );
    });

    it('updateServer - updates server successfully', async () => {
      const updates = { enabled: false };
      const mockResponse = {
        id: '1',
        name: 'Radarr',
        type: 'radarr',
        url: 'http://localhost:7878',
        apiKey: 'test-key',
        enabled: false,
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await updateServer('1', updates);

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
    });

    it('deleteServer - deletes server successfully', async () => {
      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true }),
      } as Response);

      const result = await deleteServer('1');

      expect(result.success).toBe(true);
      expect(fetch).toHaveBeenCalledWith(
        '/api/servers/1',
        expect.objectContaining({ method: 'DELETE' })
      );
    });

    it('testServer - tests server connection successfully', async () => {
      const mockResponse = {
        connected: true,
        message: 'Connection successful',
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await testServer('1');

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('Logs API', () => {
    it('getLogs - fetches logs successfully', async () => {
      const mockLogs = [
        {
          id: '1',
          timestamp: '2026-01-16T10:30:00Z',
          type: 'search',
          message: 'Searched for movie',
          metadata: {},
        },
      ];

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockLogs }),
      } as Response);

      const result = await getLogs();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockLogs);
    });

    it('deleteLogs - deletes logs successfully', async () => {
      const mockResponse = { deletedCount: 42 };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await deleteLogs();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('Stats API', () => {
    it('getStatsSummary - fetches stats successfully', async () => {
      const mockStats = {
        totalMissing: 42,
        totalCutoff: 15,
        totalSearches: 128,
        lastRun: '2026-01-16T10:30:00Z',
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockStats }),
      } as Response);

      const result = await getStatsSummary();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockStats);
    });
  });

  describe('Automation API', () => {
    it('triggerAutomation - triggers automation successfully', async () => {
      const mockResponse = {
        triggered: true,
        message: 'Automation cycle started',
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockResponse }),
      } as Response);

      const result = await triggerAutomation();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse);
    });

    it('getAutomationStatus - fetches status successfully', async () => {
      const mockStatus = {
        enabled: true,
        running: false,
        nextRun: '2026-01-16T16:00:00Z',
      };

      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ success: true, data: mockStatus }),
      } as Response);

      const result = await getAutomationStatus();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockStatus);
    });
  });

  describe('Error Handling', () => {
    it('handles network errors', async () => {
      (fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

      const result = await getConfig();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Network error');
    });

    it('handles HTTP errors', async () => {
      (fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ error: 'Internal server error' }),
      } as Response);

      const result = await getConfig();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Internal server error');
    });

    it('handles non-Error exceptions', async () => {
      (fetch as ReturnType<typeof vi.fn>).mockRejectedValueOnce('String error');

      const result = await getConfig();

      expect(result.success).toBe(false);
      expect(result.error).toBe('Network error');
    });
  });
});
