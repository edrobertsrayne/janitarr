package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Metrics collects and exposes Prometheus-compatible metrics
type Metrics struct {
	mu             sync.RWMutex
	startTime      time.Time
	cyclesTotal    int64
	cyclesFailed   int64
	searchesTotal  map[string]int64 // key: "type:category"
	searchesFailed map[string]int64 // key: "type:category"
	httpRequests   map[string]int64 // key: "method:path:status"
	httpDurations  map[string][]float64
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		startTime:      time.Now(),
		searchesTotal:  make(map[string]int64),
		searchesFailed: make(map[string]int64),
		httpRequests:   make(map[string]int64),
		httpDurations:  make(map[string][]float64),
	}
}

// IncrementCycles increments the cycle counter
func (m *Metrics) IncrementCycles(failed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cyclesTotal++
	if failed {
		m.cyclesFailed++
	}
}

// IncrementSearches increments the search counter for a specific type and category
func (m *Metrics) IncrementSearches(serverType, category string, failed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", serverType, category)
	m.searchesTotal[key]++
	if failed {
		m.searchesFailed[key]++
	}
}

// RecordHTTPRequest records an HTTP request with its duration
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%d", method, path, status)
	m.httpRequests[key]++

	// Store duration in seconds
	durationSeconds := duration.Seconds()
	m.httpDurations[key] = append(m.httpDurations[key], durationSeconds)
}

// Format returns metrics in Prometheus text format
func (m *Metrics) Format() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder

	// Uptime
	uptime := time.Since(m.startTime).Seconds()
	sb.WriteString("# HELP janitarr_uptime_seconds Time since application start\n")
	sb.WriteString("# TYPE janitarr_uptime_seconds gauge\n")
	sb.WriteString(fmt.Sprintf("janitarr_uptime_seconds %.0f\n", uptime))
	sb.WriteString("\n")

	// Cycles
	sb.WriteString("# HELP janitarr_cycles_total Total number of automation cycles executed\n")
	sb.WriteString("# TYPE janitarr_cycles_total counter\n")
	sb.WriteString(fmt.Sprintf("janitarr_cycles_total %d\n", m.cyclesTotal))
	sb.WriteString("\n")

	sb.WriteString("# HELP janitarr_cycles_failed_total Total number of failed automation cycles\n")
	sb.WriteString("# TYPE janitarr_cycles_failed_total counter\n")
	sb.WriteString(fmt.Sprintf("janitarr_cycles_failed_total %d\n", m.cyclesFailed))
	sb.WriteString("\n")

	// Searches by type and category
	if len(m.searchesTotal) > 0 {
		sb.WriteString("# HELP janitarr_searches_total Total number of searches triggered\n")
		sb.WriteString("# TYPE janitarr_searches_total counter\n")

		// Sort keys for consistent output
		var keys []string
		for k := range m.searchesTotal {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				sb.WriteString(fmt.Sprintf("janitarr_searches_total{type=\"%s\",category=\"%s\"} %d\n",
					parts[0], parts[1], m.searchesTotal[key]))
			}
		}
		sb.WriteString("\n")
	}

	if len(m.searchesFailed) > 0 {
		sb.WriteString("# HELP janitarr_searches_failed_total Total number of failed searches\n")
		sb.WriteString("# TYPE janitarr_searches_failed_total counter\n")

		// Sort keys for consistent output
		var keys []string
		for k := range m.searchesFailed {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				sb.WriteString(fmt.Sprintf("janitarr_searches_failed_total{type=\"%s\",category=\"%s\"} %d\n",
					parts[0], parts[1], m.searchesFailed[key]))
			}
		}
		sb.WriteString("\n")
	}

	// HTTP requests
	if len(m.httpRequests) > 0 {
		sb.WriteString("# HELP janitarr_http_requests_total Total number of HTTP requests\n")
		sb.WriteString("# TYPE janitarr_http_requests_total counter\n")

		// Sort keys for consistent output
		var keys []string
		for k := range m.httpRequests {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) == 3 {
				sb.WriteString(fmt.Sprintf("janitarr_http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
					parts[0], parts[1], parts[2], m.httpRequests[key]))
			}
		}
		sb.WriteString("\n")
	}

	// HTTP request durations (histogram buckets)
	if len(m.httpDurations) > 0 {
		sb.WriteString("# HELP janitarr_http_request_duration_seconds HTTP request duration in seconds\n")
		sb.WriteString("# TYPE janitarr_http_request_duration_seconds histogram\n")

		// Sort keys for consistent output
		var keys []string
		for k := range m.httpDurations {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

		for _, key := range keys {
			parts := strings.Split(key, ":")
			if len(parts) != 3 {
				continue
			}

			durations := m.httpDurations[key]
			labels := fmt.Sprintf("method=\"%s\",path=\"%s\",status=\"%s\"", parts[0], parts[1], parts[2])

			// Count observations in each bucket
			for _, bucket := range buckets {
				count := 0
				for _, d := range durations {
					if d <= bucket {
						count++
					}
				}
				sb.WriteString(fmt.Sprintf("janitarr_http_request_duration_seconds_bucket{%s,le=\"%.3f\"} %d\n",
					labels, bucket, count))
			}

			// +Inf bucket (all observations)
			sb.WriteString(fmt.Sprintf("janitarr_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n",
				labels, len(durations)))

			// Sum of all observations
			sum := 0.0
			for _, d := range durations {
				sum += d
			}
			sb.WriteString(fmt.Sprintf("janitarr_http_request_duration_seconds_sum{%s} %.6f\n", labels, sum))

			// Count of observations
			sb.WriteString(fmt.Sprintf("janitarr_http_request_duration_seconds_count{%s} %d\n", labels, len(durations)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
