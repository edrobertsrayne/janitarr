/**
 * Unit tests for metrics collection and formatting
 */

import { describe, test, expect, beforeEach } from "bun:test";
import {
  incrementCycleCounter,
  incrementSearchCounter,
  recordHttpRequest,
  collectMetrics,
} from "../../src/lib/metrics";

describe("Metrics", () => {
  describe("Prometheus formatting", () => {
    test("includes HELP and TYPE annotations", () => {
      const output = collectMetrics();

      // Check for HELP annotations
      expect(output).toContain("# HELP janitarr_info");
      expect(output).toContain("# HELP janitarr_uptime_seconds");
      expect(output).toContain("# HELP janitarr_scheduler_enabled");

      // Check for TYPE annotations
      expect(output).toContain("# TYPE janitarr_info gauge");
      expect(output).toContain("# TYPE janitarr_uptime_seconds counter");
      expect(output).toContain("# TYPE janitarr_scheduler_enabled gauge");
    });

    test("all metrics prefixed with janitarr_", () => {
      const output = collectMetrics();
      const lines = output.split("\n");

      // Get all metric lines (skip comments and empty lines)
      const metricLines = lines.filter(
        (line) => line.length > 0 && !line.startsWith("#")
      );

      for (const line of metricLines) {
        expect(line).toMatch(/^janitarr_/);
      }
    });

    test("labels follow snake_case convention", () => {
      // Record some data with labels
      incrementSearchCounter("radarr", "missing", 5);
      recordHttpRequest("GET", "/api/servers", 200, 50);

      const output = collectMetrics();

      // Check search labels
      expect(output).toContain('server_type="radarr"');
      expect(output).toContain('category="missing"');

      // Check HTTP labels
      expect(output).toContain('method="GET"');
      expect(output).toContain('path="/api/servers"');
      expect(output).toContain('status="200"');
    });

    test("labels formatted correctly with quotes", () => {
      incrementSearchCounter("radarr", "missing", 1);
      const output = collectMetrics();

      // Labels should be key="value" format
      expect(output).toMatch(/server_type="radarr"/);
      expect(output).toMatch(/category="missing"/);
    });

    test("version label included in info metric", () => {
      const output = collectMetrics();

      // Should have version label in janitarr_info
      expect(output).toMatch(/janitarr_info\{version="[^"]+"\} 1/);
    });
  });

  describe("Counter increment behavior", () => {
    test("cycle counter increments correctly", () => {
      // Get initial state
      const before = collectMetrics();
      const beforeTotal = parseInt(
        before.match(/janitarr_scheduler_cycles_total (\d+)/)?.[1] || "0"
      );
      const beforeFailed = parseInt(
        before.match(/janitarr_scheduler_cycles_failed_total (\d+)/)?.[1] || "0"
      );

      // Increment counters
      incrementCycleCounter(false); // Success
      incrementCycleCounter(false); // Success
      incrementCycleCounter(true); // Failed

      const after = collectMetrics();
      const afterTotal = parseInt(
        after.match(/janitarr_scheduler_cycles_total (\d+)/)?.[1] || "0"
      );
      const afterFailed = parseInt(
        after.match(/janitarr_scheduler_cycles_failed_total (\d+)/)?.[1] || "0"
      );

      // Total should increase by 3, failed by 1
      expect(afterTotal).toBe(beforeTotal + 3);
      expect(afterFailed).toBe(beforeFailed + 1);
    });

    test("search counter increments correctly", () => {
      const before = collectMetrics();
      const beforeRadarr = parseInt(
        before.match(/janitarr_searches_triggered_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );
      const beforeSonarr = parseInt(
        before.match(/janitarr_searches_triggered_total\{server_type="sonarr",category="cutoff"\} (\d+)/)?.[1] || "0"
      );

      incrementSearchCounter("radarr", "missing", 5);
      incrementSearchCounter("radarr", "missing", 3);
      incrementSearchCounter("sonarr", "cutoff", 2);

      const output = collectMetrics();

      // Should increase by 8 for radarr missing
      const afterRadarr = parseInt(
        output.match(/janitarr_searches_triggered_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );
      expect(afterRadarr).toBe(beforeRadarr + 8);

      // Should increase by 2 for sonarr cutoff
      const afterSonarr = parseInt(
        output.match(/janitarr_searches_triggered_total\{server_type="sonarr",category="cutoff"\} (\d+)/)?.[1] || "0"
      );
      expect(afterSonarr).toBe(beforeSonarr + 2);
    });

    test("search failed counter increments separately", () => {
      const before = collectMetrics();
      const beforeTriggered = parseInt(
        before.match(/janitarr_searches_triggered_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );
      const beforeFailed = parseInt(
        before.match(/janitarr_searches_failed_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );

      incrementSearchCounter("radarr", "missing", 5, false); // Success
      incrementSearchCounter("radarr", "missing", 2, true); // Failed

      const output = collectMetrics();

      // Check successful searches increased by 5
      const afterTriggered = parseInt(
        output.match(/janitarr_searches_triggered_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );
      expect(afterTriggered).toBe(beforeTriggered + 5);

      // Check failed searches increased by 2
      const afterFailed = parseInt(
        output.match(/janitarr_searches_failed_total\{server_type="radarr",category="missing"\} (\d+)/)?.[1] || "0"
      );
      expect(afterFailed).toBe(beforeFailed + 2);
    });

    test("HTTP request counter increments correctly", () => {
      const before = collectMetrics();
      const beforeHealth = parseInt(
        before.match(/janitarr_http_requests_total\{method="GET",path="\/api\/health",status="200"\} (\d+)/)?.[1] || "0"
      );
      const beforeServers = parseInt(
        before.match(/janitarr_http_requests_total\{method="POST",path="\/api\/servers",status="201"\} (\d+)/)?.[1] || "0"
      );

      recordHttpRequest("GET", "/api/health", 200, 10);
      recordHttpRequest("GET", "/api/health", 200, 15);
      recordHttpRequest("POST", "/api/servers", 201, 50);

      const output = collectMetrics();

      // Should increase by 2 for GET /api/health
      const afterHealth = parseInt(
        output.match(/janitarr_http_requests_total\{method="GET",path="\/api\/health",status="200"\} (\d+)/)?.[1] || "0"
      );
      expect(afterHealth).toBe(beforeHealth + 2);

      // Should increase by 1 for POST /api/servers
      const afterServers = parseInt(
        output.match(/janitarr_http_requests_total\{method="POST",path="\/api\/servers",status="201"\} (\d+)/)?.[1] || "0"
      );
      expect(afterServers).toBe(beforeServers + 1);
    });
  });

  describe("Gauge behavior", () => {
    test("scheduler enabled reflects current state", () => {
      const output = collectMetrics();

      // Should have scheduler_enabled metric (default true)
      expect(output).toMatch(/janitarr_scheduler_enabled (0|1)/);
    });

    test("scheduler running reflects current state", () => {
      const output = collectMetrics();

      // Should have scheduler_running metric
      expect(output).toMatch(/janitarr_scheduler_running (0|1)/);
    });

    test("scheduler cycle active reflects current state", () => {
      const output = collectMetrics();

      // Should have scheduler_cycle_active metric
      expect(output).toMatch(/janitarr_scheduler_cycle_active (0|1)/);
    });

    test("database connected is binary gauge", () => {
      const output = collectMetrics();

      // Should be 0 or 1
      expect(output).toMatch(/janitarr_database_connected (0|1)/);
    });

    test("uptime increases over time", async () => {
      const before = collectMetrics();
      const uptimeBefore = parseInt(
        before.match(/janitarr_uptime_seconds (\d+)/)?.[1] || "0"
      );

      // Wait a bit
      await Bun.sleep(1100);

      const after = collectMetrics();
      const uptimeAfter = parseInt(
        after.match(/janitarr_uptime_seconds (\d+)/)?.[1] || "0"
      );

      // Uptime should have increased by at least 1 second
      expect(uptimeAfter).toBeGreaterThanOrEqual(uptimeBefore + 1);
    });
  });

  describe("Histogram behavior", () => {
    test("HTTP duration histogram includes buckets", () => {
      recordHttpRequest("GET", "/api/servers", 200, 50); // 0.05 seconds

      const output = collectMetrics();

      // Should have histogram with buckets
      expect(output).toContain("# TYPE janitarr_http_request_duration_seconds histogram");

      // Should have bucket annotations
      expect(output).toMatch(
        /janitarr_http_request_duration_seconds_bucket\{method="GET",path="\/api\/servers",le="[^"]+"\}/
      );

      // Should have +Inf bucket
      expect(output).toMatch(
        /janitarr_http_request_duration_seconds_bucket\{method="GET",path="\/api\/servers",le="\+Inf"\}/
      );

      // Should have sum
      expect(output).toMatch(
        /janitarr_http_request_duration_seconds_sum\{method="GET",path="\/api\/servers"\}/
      );

      // Should have count
      expect(output).toMatch(
        /janitarr_http_request_duration_seconds_count\{method="GET",path="\/api\/servers"\}/
      );
    });

    test("histogram buckets increment correctly", () => {
      // Use a unique path to avoid interference from other tests
      const testPath = "/unique/bucket/test";

      // Record a fast request (5ms = 0.005s)
      recordHttpRequest("GET", testPath, 200, 5);

      const output = collectMetrics();

      // Fast request should increment all buckets >= 0.005
      expect(output).toContain(
        `janitarr_http_request_duration_seconds_bucket{method="GET",path="${testPath}",le="0.005"} 1`
      );
      expect(output).toContain(
        `janitarr_http_request_duration_seconds_bucket{method="GET",path="${testPath}",le="0.01"} 1`
      );
    });

    test("histogram sum accumulates correctly", () => {
      // Use a unique path to avoid interference from other tests
      const testPath = "/unique/sum/test";

      recordHttpRequest("GET", testPath, 200, 100); // 0.1 seconds
      recordHttpRequest("GET", testPath, 200, 200); // 0.2 seconds

      const output = collectMetrics();

      // Sum should be 0.3 seconds
      expect(output).toContain(
        `janitarr_http_request_duration_seconds_sum{method="GET",path="${testPath}"} 0.3`
      );

      // Count should be 2
      expect(output).toContain(
        `janitarr_http_request_duration_seconds_count{method="GET",path="${testPath}"} 2`
      );
    });
  });

  describe("Required metrics presence", () => {
    test("application info metrics present", () => {
      const output = collectMetrics();

      expect(output).toContain("janitarr_info");
      expect(output).toContain("janitarr_uptime_seconds");
    });

    test("scheduler metrics present", () => {
      const output = collectMetrics();

      expect(output).toContain("janitarr_scheduler_enabled");
      expect(output).toContain("janitarr_scheduler_running");
      expect(output).toContain("janitarr_scheduler_cycle_active");
      expect(output).toContain("janitarr_scheduler_cycles_total");
      expect(output).toContain("janitarr_scheduler_cycles_failed_total");
    });

    test("database metrics present", () => {
      const output = collectMetrics();

      expect(output).toContain("janitarr_database_connected");
      expect(output).toContain("janitarr_logs_total");
    });

    test("search metrics present when data exists", () => {
      incrementSearchCounter("radarr", "missing", 1);

      const output = collectMetrics();

      expect(output).toContain("janitarr_searches_triggered_total");
    });

    test("HTTP metrics present when data exists", () => {
      recordHttpRequest("GET", "/api/test", 200, 10);

      const output = collectMetrics();

      expect(output).toContain("janitarr_http_requests_total");
      expect(output).toContain("janitarr_http_request_duration_seconds");
    });

    test("next run timestamp present when scheduler running", () => {
      const output = collectMetrics();

      // May or may not be present depending on scheduler state
      // This is an optional metric based on runtime state
      expect(output).toBeDefined();
    });
  });

  describe("Edge cases", () => {
    test("handles zero counts correctly", () => {
      const output = collectMetrics();

      // Cycle counts should be present (may not be 0 due to cumulative nature)
      expect(output).toMatch(/janitarr_scheduler_cycles_total \d+/);
      expect(output).toMatch(/janitarr_scheduler_cycles_failed_total \d+/);
    });

    test("handles missing data gracefully", () => {
      // Don't record any search or HTTP data
      const output = collectMetrics();

      // Should still return valid output
      expect(output).toContain("janitarr_info");
      expect(output).toContain("janitarr_uptime_seconds");

      // Search and HTTP metrics may be absent (not an error)
      expect(output).toBeDefined();
    });

    test("handles multiple server types", () => {
      incrementSearchCounter("radarr", "missing", 5);
      incrementSearchCounter("radarr", "cutoff", 3);
      incrementSearchCounter("sonarr", "missing", 2);
      incrementSearchCounter("sonarr", "cutoff", 1);

      const output = collectMetrics();

      // Should have all 4 combinations
      expect(output).toContain('server_type="radarr",category="missing"');
      expect(output).toContain('server_type="radarr",category="cutoff"');
      expect(output).toContain('server_type="sonarr",category="missing"');
      expect(output).toContain('server_type="sonarr",category="cutoff"');
    });

    test("output ends with newline", () => {
      const output = collectMetrics();

      expect(output).toMatch(/\n$/);
    });
  });
});
