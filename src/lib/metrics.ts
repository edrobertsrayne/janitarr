/**
 * Prometheus Metrics Collection
 *
 * Provides utilities for collecting and exposing Prometheus-compatible metrics
 * for monitoring application health, performance, and behavior.
 */

import { getDatabase } from "../storage/database";
import { getStatus } from "./scheduler";

/** Process start time for uptime calculation */
const startTime = Date.now();

/** Counter storage for monotonic metrics */
interface Counters {
  schedulerCyclesTotal: number;
  schedulerCyclesFailed: number;
  searchesTriggered: Map<string, number>; // key: "serverType:category"
  searchesFailed: Map<string, number>; // key: "serverType:category"
  httpRequests: Map<string, number>; // key: "method:path:status"
}

/** Histogram buckets for HTTP request duration */
interface HistogramBucket {
  le: number; // less than or equal to
  count: number;
}

/** Histogram storage for HTTP request duration */
interface DurationHistogram {
  buckets: HistogramBucket[];
  sum: number;
  count: number;
}

/** Map of HTTP request durations by method:path */
const durationHistograms = new Map<string, DurationHistogram>();

/** Standard buckets for HTTP request duration (in seconds) */
const DURATION_BUCKETS = [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10];

/** Singleton counter storage */
const counters: Counters = {
  schedulerCyclesTotal: 0,
  schedulerCyclesFailed: 0,
  searchesTriggered: new Map(),
  searchesFailed: new Map(),
  httpRequests: new Map(),
};

/**
 * Increment scheduler cycle counter
 */
export function incrementCycleCounter(failed = false): void {
  counters.schedulerCyclesTotal++;
  if (failed) {
    counters.schedulerCyclesFailed++;
  }
}

/**
 * Increment search trigger counter
 */
export function incrementSearchCounter(
  serverType: "radarr" | "sonarr",
  category: "missing" | "cutoff",
  count: number,
  failed = false
): void {
  const key = `${serverType}:${category}`;
  const map = failed ? counters.searchesFailed : counters.searchesTriggered;
  const current = map.get(key) || 0;
  map.set(key, current + count);
}

/**
 * Record HTTP request
 */
export function recordHttpRequest(
  method: string,
  path: string,
  status: number,
  durationMs: number
): void {
  // Record request counter
  const counterKey = `${method}:${path}:${status}`;
  const current = counters.httpRequests.get(counterKey) || 0;
  counters.httpRequests.set(counterKey, current + 1);

  // Record duration histogram
  const histogramKey = `${method}:${path}`;
  let histogram = durationHistograms.get(histogramKey);

  if (!histogram) {
    histogram = {
      buckets: DURATION_BUCKETS.map((le) => ({ le, count: 0 })),
      sum: 0,
      count: 0,
    };
    durationHistograms.set(histogramKey, histogram);
  }

  // Update histogram
  const durationSec = durationMs / 1000;
  histogram.sum += durationSec;
  histogram.count++;

  // Increment buckets
  for (const bucket of histogram.buckets) {
    if (durationSec <= bucket.le) {
      bucket.count++;
    }
  }
}

/**
 * Get application version from package.json
 */
function getAppVersion(): string {
  try {
    // In production, this should be set via environment variable or build process
    return process.env.npm_package_version || "0.1.0";
  } catch {
    return "0.1.0";
  }
}

/**
 * Get uptime in seconds
 */
function getUptimeSeconds(): number {
  return Math.floor((Date.now() - startTime) / 1000);
}

/**
 * Format a metric with HELP and TYPE annotations
 */
function formatMetric(
  name: string,
  type: "counter" | "gauge" | "histogram",
  help: string,
  values: Array<{ labels?: Record<string, string>; value: number | string }>
): string {
  const lines: string[] = [];

  lines.push(`# HELP ${name} ${help}`);
  lines.push(`# TYPE ${name} ${type}`);

  for (const { labels, value } of values) {
    if (labels && Object.keys(labels).length > 0) {
      const labelStr = Object.entries(labels)
        .map(([k, v]) => `${k}="${v}"`)
        .join(",");
      lines.push(`${name}{${labelStr}} ${value}`);
    } else {
      lines.push(`${name} ${value}`);
    }
  }

  return lines.join("\n");
}

/**
 * Format histogram metric
 */
function formatHistogram(
  name: string,
  help: string,
  histograms: Map<string, DurationHistogram>,
  labelKeys: string[]
): string {
  const lines: string[] = [];

  lines.push(`# HELP ${name} ${help}`);
  lines.push(`# TYPE ${name} histogram`);

  for (const [key, histogram] of histograms) {
    const labelValues = key.split(":");
    const labels: Record<string, string> = {};

    // Build labels from key
    labelKeys.forEach((labelKey, index) => {
      labels[labelKey] = labelValues[index] || "";
    });

    // Output buckets
    for (const bucket of histogram.buckets) {
      const bucketLabels = { ...labels, le: bucket.le.toString() };
      const labelStr = Object.entries(bucketLabels)
        .map(([k, v]) => `${k}="${v}"`)
        .join(",");
      lines.push(`${name}_bucket{${labelStr}} ${bucket.count}`);
    }

    // Output +Inf bucket
    const infLabels = { ...labels, le: "+Inf" };
    const infLabelStr = Object.entries(infLabels)
      .map(([k, v]) => `${k}="${v}"`)
      .join(",");
    lines.push(`${name}_bucket{${infLabelStr}} ${histogram.count}`);

    // Output sum and count
    const labelStr = Object.entries(labels)
      .map(([k, v]) => `${k}="${v}"`)
      .join(",");
    lines.push(`${name}_sum{${labelStr}} ${histogram.sum}`);
    lines.push(`${name}_count{${labelStr}} ${histogram.count}`);
  }

  return lines.join("\n");
}

/**
 * Collect and format all metrics in Prometheus text format
 */
export function collectMetrics(): string {
  const metrics: string[] = [];

  // Application info
  metrics.push(
    formatMetric("janitarr_info", "gauge", "Application version information", [
      { labels: { version: getAppVersion() }, value: 1 },
    ])
  );

  // Uptime
  metrics.push(
    formatMetric("janitarr_uptime_seconds", "counter", "Time since process start", [
      { value: getUptimeSeconds() },
    ])
  );

  // Scheduler metrics
  const schedulerStatus = getStatus();

  metrics.push(
    formatMetric("janitarr_scheduler_enabled", "gauge", "Whether scheduler is enabled", [
      { value: schedulerStatus.config.enabled ? 1 : 0 },
    ])
  );

  metrics.push(
    formatMetric("janitarr_scheduler_running", "gauge", "Whether scheduler is running", [
      { value: schedulerStatus.isRunning ? 1 : 0 },
    ])
  );

  metrics.push(
    formatMetric(
      "janitarr_scheduler_cycle_active",
      "gauge",
      "Whether automation cycle is active",
      [{ value: schedulerStatus.isCycleActive ? 1 : 0 }]
    )
  );

  metrics.push(
    formatMetric("janitarr_scheduler_cycles_total", "counter", "Total automation cycles executed", [
      { value: counters.schedulerCyclesTotal },
    ])
  );

  metrics.push(
    formatMetric(
      "janitarr_scheduler_cycles_failed_total",
      "counter",
      "Total failed automation cycles",
      [{ value: counters.schedulerCyclesFailed }]
    )
  );

  // Next run timestamp
  if (schedulerStatus.nextRunTime) {
    const timestamp = Math.floor(schedulerStatus.nextRunTime.getTime() / 1000);
    metrics.push(
      formatMetric(
        "janitarr_scheduler_next_run_timestamp",
        "gauge",
        "Unix timestamp of next scheduled run",
        [{ value: timestamp }]
      )
    );
  }

  // Search metrics - triggered
  const triggeredValues: Array<{ labels: Record<string, string>; value: number }> = [];
  for (const [key, count] of counters.searchesTriggered) {
    const [serverType, category] = key.split(":");
    triggeredValues.push({
      labels: { server_type: serverType, category },
      value: count,
    });
  }
  if (triggeredValues.length > 0) {
    metrics.push(
      formatMetric(
        "janitarr_searches_triggered_total",
        "counter",
        "Total searches triggered by type",
        triggeredValues
      )
    );
  }

  // Search metrics - failed
  const failedValues: Array<{ labels: Record<string, string>; value: number }> = [];
  for (const [key, count] of counters.searchesFailed) {
    const [serverType, category] = key.split(":");
    failedValues.push({
      labels: { server_type: serverType, category },
      value: count,
    });
  }
  if (failedValues.length > 0) {
    metrics.push(
      formatMetric(
        "janitarr_searches_failed_total",
        "counter",
        "Total failed searches by type",
        failedValues
      )
    );
  }

  // Server metrics
  try {
    const db = getDatabase();
    const servers = db.listServers();

    // Count servers by type
    const radarrCount = servers.filter((s) => s.type === "radarr").length;
    const sonarrCount = servers.filter((s) => s.type === "sonarr").length;
    const radarrEnabled = servers.filter((s) => s.type === "radarr" && s.enabled).length;
    const sonarrEnabled = servers.filter((s) => s.type === "sonarr" && s.enabled).length;

    metrics.push(
      formatMetric(
        "janitarr_servers_configured",
        "gauge",
        "Number of configured servers by type",
        [
          { labels: { type: "radarr" }, value: radarrCount },
          { labels: { type: "sonarr" }, value: sonarrCount },
        ]
      )
    );

    metrics.push(
      formatMetric("janitarr_servers_enabled", "gauge", "Number of enabled servers by type", [
        { labels: { type: "radarr" }, value: radarrEnabled },
        { labels: { type: "sonarr" }, value: sonarrEnabled },
      ])
    );
  } catch (error) {
    // Database not available, skip server metrics
    console.error("Failed to collect server metrics:", error);
  }

  // Database metrics
  try {
    const db = getDatabase();

    // Test database connection
    const dbConnected = db.testConnection();
    metrics.push(
      formatMetric("janitarr_database_connected", "gauge", "Database connection status", [
        { value: dbConnected ? 1 : 0 },
      ])
    );

    // Get log count
    const logCount = db.getLogCount();
    metrics.push(
      formatMetric("janitarr_logs_total", "gauge", "Total log entries in database", [
        { value: logCount },
      ])
    );
  } catch (error) {
    // Database not available
    console.error("Failed to collect database metrics:", error);
    metrics.push(
      formatMetric("janitarr_database_connected", "gauge", "Database connection status", [
        { value: 0 },
      ])
    );
  }

  // HTTP metrics - request counter
  const httpValues: Array<{ labels: Record<string, string>; value: number }> = [];
  for (const [key, count] of counters.httpRequests) {
    const [method, path, status] = key.split(":");
    httpValues.push({
      labels: { method, path, status },
      value: count,
    });
  }
  if (httpValues.length > 0) {
    metrics.push(
      formatMetric("janitarr_http_requests_total", "counter", "Total HTTP requests", httpValues)
    );
  }

  // HTTP metrics - duration histogram
  if (durationHistograms.size > 0) {
    metrics.push(
      formatHistogram(
        "janitarr_http_request_duration_seconds",
        "HTTP request duration in seconds",
        durationHistograms,
        ["method", "path"]
      )
    );
  }

  return metrics.join("\n\n") + "\n";
}
