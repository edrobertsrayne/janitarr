/**
 * Metrics Endpoint Handler
 *
 * Exposes Prometheus-compatible metrics for monitoring application health,
 * performance, and behavior.
 */

import { collectMetrics } from "../../lib/metrics";

/**
 * Handle GET /metrics request
 * Returns Prometheus text format metrics
 */
export function handleMetrics(): Response {
  const metrics = collectMetrics();

  return new Response(metrics, {
    status: 200,
    headers: {
      "Content-Type": "text/plain; version=0.0.4; charset=utf-8",
    },
  });
}
