/**
 * Unit tests for metrics endpoint
 */

import { describe, test, expect } from "bun:test";
import { handleMetrics } from "../../../src/web/routes/metrics";
import { incrementCycleCounter, recordHttpRequest } from "../../../src/lib/metrics";

describe("Metrics Endpoint", () => {
  test("returns HTTP 200 status", () => {
    const response = handleMetrics();

    expect(response.status).toBe(200);
  });

  test("returns correct Content-Type header", () => {
    const response = handleMetrics();

    const contentType = response.headers.get("Content-Type");
    expect(contentType).toBe("text/plain; version=0.0.4; charset=utf-8");
  });

  test("returns Prometheus text format", async () => {
    const response = handleMetrics();
    const text = await response.text();

    // Should have HELP and TYPE annotations
    expect(text).toContain("# HELP");
    expect(text).toContain("# TYPE");

    // Should have metric data
    expect(text).toMatch(/^janitarr_/m);
  });

  test("includes all required metrics", async () => {
    const response = handleMetrics();
    const text = await response.text();

    // Application info metrics
    expect(text).toContain("janitarr_info");
    expect(text).toContain("janitarr_uptime_seconds");

    // Scheduler metrics
    expect(text).toContain("janitarr_scheduler_enabled");
    expect(text).toContain("janitarr_scheduler_running");
    expect(text).toContain("janitarr_scheduler_cycle_active");
    expect(text).toContain("janitarr_scheduler_cycles_total");
    expect(text).toContain("janitarr_scheduler_cycles_failed_total");

    // Database metrics
    expect(text).toContain("janitarr_database_connected");
    expect(text).toContain("janitarr_logs_total");
  });

  test("reflects runtime data correctly", async () => {
    // Record some data
    incrementCycleCounter(false);
    incrementCycleCounter(true);
    recordHttpRequest("GET", "/metrics", 200, 5);

    const response = handleMetrics();
    const text = await response.text();

    // Should show the cycle counts
    expect(text).toContain("janitarr_scheduler_cycles_total");
    expect(text).toContain("janitarr_scheduler_cycles_failed_total");

    // Should show HTTP requests
    expect(text).toContain("janitarr_http_requests_total");
  });

  test("response is valid Prometheus format", async () => {
    const response = handleMetrics();
    const text = await response.text();

    const lines = text.split("\n");

    // Check that HELP lines come before TYPE lines
    let lastHelp = -1;
    let lastType = -1;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      if (line.startsWith("# HELP")) {
        lastHelp = i;
      } else if (line.startsWith("# TYPE")) {
        lastType = i;

        // TYPE should come immediately after HELP for the same metric
        expect(lastType).toBe(lastHelp + 1);
      }
    }
  });

  test("all metric names follow naming convention", async () => {
    const response = handleMetrics();
    const text = await response.text();

    const lines = text.split("\n");
    const metricLines = lines.filter(
      (line) => line.length > 0 && !line.startsWith("#")
    );

    for (const line of metricLines) {
      // Should start with janitarr_
      expect(line).toMatch(/^janitarr_[a-z_]+/);

      // Metric name should be snake_case (lowercase with underscores)
      const metricName = line.split(/[\s{]/)[0];
      expect(metricName).toMatch(/^[a-z_]+$/);
    }
  });

  test("response ends with newline", async () => {
    const response = handleMetrics();
    const text = await response.text();

    expect(text).toMatch(/\n$/);
  });

  test("response is lightweight and fast", async () => {
    const start = Date.now();
    const response = handleMetrics();
    await response.text();
    const duration = Date.now() - start;

    // Should be very fast (< 200ms as per spec)
    expect(duration).toBeLessThan(200);
  });

  test("handles multiple calls correctly", async () => {
    // First call
    const response1 = handleMetrics();
    const text1 = await response1.text();

    // Second call
    const response2 = handleMetrics();
    const text2 = await response2.text();

    // Both should be valid
    expect(text1).toContain("janitarr_info");
    expect(text2).toContain("janitarr_info");

    // Uptime should increase (or stay same if too fast)
    const uptime1 = parseInt(text1.match(/janitarr_uptime_seconds (\d+)/)?.[1] || "0");
    const uptime2 = parseInt(text2.match(/janitarr_uptime_seconds (\d+)/)?.[1] || "0");

    expect(uptime2).toBeGreaterThanOrEqual(uptime1);
  });
});
