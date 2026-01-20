package metrics

import (
	"context"
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
	version        string
	cyclesTotal    int64
	cyclesFailed   int64
	searchesTotal  map[string]int64 // key: "type:category"
	searchesFailed map[string]int64 // key: "type:category"
	httpRequests   map[string]int64 // key: "method:path:status"
	httpDurations  map[string][]float64
	scheduler      SchedulerStatusProvider
	database       DatabaseProvider
	cacheExpiry    time.Time
	cachedLogCount int
	cachedDbStatus int // 1 for connected, 0 for disconnected
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

// SetScheduler sets the scheduler provider for metrics
func (m *Metrics) SetScheduler(scheduler SchedulerStatusProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scheduler = scheduler
}

// SetDatabase sets the database provider for metrics
func (m *Metrics) SetDatabase(database DatabaseProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.database = database
}

// SetVersion sets the application version
func (m *Metrics) SetVersion(version string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version = version
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

	// Capture references to providers under lock
	scheduler := m.scheduler
	database := m.database
	version := m.version

	m.mu.RUnlock()

	var sb strings.Builder

	// Application info
	if version != "" {
		sb.WriteString("# HELP janitarr_info Application version information\n")
		sb.WriteString("# TYPE janitarr_info gauge\n")
		sb.WriteString(fmt.Sprintf("janitarr_info{version=\"%s\"} 1\n", version))
		sb.WriteString("\n")
	}

	// Uptime
	m.mu.RLock()
	uptime := time.Since(m.startTime).Seconds()
	m.mu.RUnlock()

	sb.WriteString("# HELP janitarr_uptime_seconds Time since application start\n")
	sb.WriteString("# TYPE janitarr_uptime_seconds gauge\n")
	sb.WriteString(fmt.Sprintf("janitarr_uptime_seconds %.0f\n", uptime))
	sb.WriteString("\n")

	// Scheduler metrics
	if scheduler != nil {
		status := scheduler.GetStatus()

		sb.WriteString("# HELP janitarr_scheduler_enabled Whether scheduler is enabled\n")
		sb.WriteString("# TYPE janitarr_scheduler_enabled gauge\n")
		// Scheduler enabled is inferred from intervalHours > 0
		if status.IntervalHours > 0 {
			sb.WriteString("janitarr_scheduler_enabled 1\n")
		} else {
			sb.WriteString("janitarr_scheduler_enabled 0\n")
		}
		sb.WriteString("\n")

		sb.WriteString("# HELP janitarr_scheduler_running Whether scheduler is running\n")
		sb.WriteString("# TYPE janitarr_scheduler_running gauge\n")
		if status.IsRunning {
			sb.WriteString("janitarr_scheduler_running 1\n")
		} else {
			sb.WriteString("janitarr_scheduler_running 0\n")
		}
		sb.WriteString("\n")

		sb.WriteString("# HELP janitarr_scheduler_cycle_active Whether automation cycle is active\n")
		sb.WriteString("# TYPE janitarr_scheduler_cycle_active gauge\n")
		if status.IsCycleActive {
			sb.WriteString("janitarr_scheduler_cycle_active 1\n")
		} else {
			sb.WriteString("janitarr_scheduler_cycle_active 0\n")
		}
		sb.WriteString("\n")

		if status.NextRun != nil {
			sb.WriteString("# HELP janitarr_scheduler_next_run_timestamp Unix timestamp of next scheduled run\n")
			sb.WriteString("# TYPE janitarr_scheduler_next_run_timestamp gauge\n")
			sb.WriteString(fmt.Sprintf("janitarr_scheduler_next_run_timestamp %d\n", status.NextRun.Unix()))
			sb.WriteString("\n")
		}
	}

	m.mu.RLock()
	cyclesTotal := m.cyclesTotal
	cyclesFailed := m.cyclesFailed
	m.mu.RUnlock()

	// Cycles
	sb.WriteString("# HELP janitarr_cycles_total Total number of automation cycles executed\n")
	sb.WriteString("# TYPE janitarr_cycles_total counter\n")
	sb.WriteString(fmt.Sprintf("janitarr_cycles_total %d\n", cyclesTotal))
	sb.WriteString("\n")

	sb.WriteString("# HELP janitarr_cycles_failed_total Total number of failed automation cycles\n")
	sb.WriteString("# TYPE janitarr_cycles_failed_total counter\n")
	sb.WriteString(fmt.Sprintf("janitarr_cycles_failed_total %d\n", cyclesFailed))
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

	// Server metrics
	if database != nil {
		serverCounts, err := database.GetServerCounts()
		if err == nil && len(serverCounts) > 0 {
			sb.WriteString("# HELP janitarr_servers_configured Number of configured servers by type\n")
			sb.WriteString("# TYPE janitarr_servers_configured gauge\n")

			// Sort server types for consistent output
			var types []string
			for t := range serverCounts {
				types = append(types, t)
			}
			sort.Strings(types)

			for _, serverType := range types {
				counts := serverCounts[serverType]
				sb.WriteString(fmt.Sprintf("janitarr_servers_configured{type=\"%s\"} %d\n", serverType, counts.Configured))
			}
			sb.WriteString("\n")

			sb.WriteString("# HELP janitarr_servers_enabled Number of enabled servers by type\n")
			sb.WriteString("# TYPE janitarr_servers_enabled gauge\n")
			for _, serverType := range types {
				counts := serverCounts[serverType]
				sb.WriteString(fmt.Sprintf("janitarr_servers_enabled{type=\"%s\"} %d\n", serverType, counts.Enabled))
			}
			sb.WriteString("\n")
		}
	}

	// Database metrics
	if database != nil {
		// Database connection status
		sb.WriteString("# HELP janitarr_database_connected Database connection status\n")
		sb.WriteString("# TYPE janitarr_database_connected gauge\n")

		m.mu.Lock()
		// Check cache expiry (15 second TTL)
		now := time.Now()
		if now.After(m.cacheExpiry) {
			// Cache expired, refresh
			if err := database.Ping(); err == nil {
				m.cachedDbStatus = 1
			} else {
				m.cachedDbStatus = 0
			}

			// Get log count
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if count, err := database.GetLogCount(ctx); err == nil {
				m.cachedLogCount = count
			}

			m.cacheExpiry = now.Add(15 * time.Second)
		}
		dbStatus := m.cachedDbStatus
		logCount := m.cachedLogCount
		m.mu.Unlock()

		sb.WriteString(fmt.Sprintf("janitarr_database_connected %d\n", dbStatus))
		sb.WriteString("\n")

		// Log count
		sb.WriteString("# HELP janitarr_logs_total Total log entries in database\n")
		sb.WriteString("# TYPE janitarr_logs_total gauge\n")
		sb.WriteString(fmt.Sprintf("janitarr_logs_total %d\n", logCount))
		sb.WriteString("\n")
	}

	return sb.String()
}
