/**
 * Integration tests for API client against real Radarr/Sonarr servers
 *
 * These tests require the test servers from .env to be accessible.
 * Skip these tests in CI or when servers are unavailable.
 */

import { describe, test, expect } from "bun:test";
import { RadarrClient, SonarrClient } from "../../src/lib/api-client";

// Load test credentials from environment
const RADARR_URL = process.env.RADARR_URL ?? "";
const RADARR_API_KEY = process.env.RADARR_API_KEY ?? "";
const SONARR_URL = process.env.SONARR_URL ?? "";
const SONARR_API_KEY = process.env.SONARR_API_KEY ?? "";

const hasRadarr = RADARR_URL && RADARR_API_KEY;
const hasSonarr = SONARR_URL && SONARR_API_KEY;

describe("Radarr Integration", () => {
  test.skipIf(!hasRadarr)("connects to Radarr server", async () => {
    const client = new RadarrClient(RADARR_URL, RADARR_API_KEY);
    const result = await client.testConnection();

    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.appName).toBe("Radarr");
      expect(result.data.version).toBeDefined();
    }
  });

  test.skipIf(!hasRadarr)("fetches missing movies", async () => {
    const client = new RadarrClient(RADARR_URL, RADARR_API_KEY);
    const result = await client.getWantedMissing(1, 10);

    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.page).toBe(1);
      expect(result.data.pageSize).toBe(10);
      expect(typeof result.data.totalRecords).toBe("number");
      expect(Array.isArray(result.data.records)).toBe(true);
    }
  });

  test.skipIf(!hasRadarr)("fetches cutoff unmet movies", async () => {
    const client = new RadarrClient(RADARR_URL, RADARR_API_KEY);
    const result = await client.getCutoffUnmet(1, 10);

    expect(result.success).toBe(true);
    if (result.success) {
      expect(typeof result.data.totalRecords).toBe("number");
      expect(Array.isArray(result.data.records)).toBe(true);
    }
  });

  test.skipIf(!hasRadarr)("handles invalid API key", async () => {
    const client = new RadarrClient(RADARR_URL, "invalid-key");
    const result = await client.testConnection();

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error).toContain("unauthorized");
    }
  });
});

describe("Sonarr Integration", () => {
  test.skipIf(!hasSonarr)("connects to Sonarr server", async () => {
    const client = new SonarrClient(SONARR_URL, SONARR_API_KEY);
    const result = await client.testConnection();

    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.appName).toBe("Sonarr");
      expect(result.data.version).toBeDefined();
    }
  });

  test.skipIf(!hasSonarr)("fetches missing episodes", async () => {
    const client = new SonarrClient(SONARR_URL, SONARR_API_KEY);
    const result = await client.getWantedMissing(1, 10);

    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.page).toBe(1);
      expect(result.data.pageSize).toBe(10);
      expect(typeof result.data.totalRecords).toBe("number");
      expect(Array.isArray(result.data.records)).toBe(true);
    }
  });

  test.skipIf(!hasSonarr)("fetches cutoff unmet episodes", async () => {
    const client = new SonarrClient(SONARR_URL, SONARR_API_KEY);
    const result = await client.getCutoffUnmet(1, 10);

    expect(result.success).toBe(true);
    if (result.success) {
      expect(typeof result.data.totalRecords).toBe("number");
      expect(Array.isArray(result.data.records)).toBe(true);
    }
  });

  test.skipIf(!hasSonarr)("handles invalid API key", async () => {
    const client = new SonarrClient(SONARR_URL, "invalid-key");
    const result = await client.testConnection();

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error).toContain("unauthorized");
    }
  });
});

describe("Error Handling", () => {
  test("handles unreachable server", async () => {
    const client = new RadarrClient("http://localhost:59999", "any-key", 2000);
    const result = await client.testConnection();

    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error).toContain("unreachable");
    }
  });

  test("handles timeout", async () => {
    // Use a very short timeout to force a timeout error
    const client = new RadarrClient("http://10.255.255.1:7878", "any-key", 100);
    const result = await client.testConnection();

    expect(result.success).toBe(false);
    if (!result.success) {
      // Either timeout or unreachable is acceptable
      expect(
        result.error.includes("timed out") || result.error.includes("unreachable")
      ).toBe(true);
    }
  });
});
