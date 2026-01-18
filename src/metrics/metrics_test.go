package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	if m.cyclesTotal != 0 {
		t.Errorf("expected cyclesTotal to be 0, got %d", m.cyclesTotal)
	}

	if m.cyclesFailed != 0 {
		t.Errorf("expected cyclesFailed to be 0, got %d", m.cyclesFailed)
	}

	if m.searchesTotal == nil {
		t.Error("searchesTotal map is nil")
	}

	if m.searchesFailed == nil {
		t.Error("searchesFailed map is nil")
	}

	if m.httpRequests == nil {
		t.Error("httpRequests map is nil")
	}

	if m.httpDurations == nil {
		t.Error("httpDurations map is nil")
	}
}

func TestIncrementCycles(t *testing.T) {
	m := NewMetrics()

	// Increment successful cycle
	m.IncrementCycles(false)
	if m.cyclesTotal != 1 {
		t.Errorf("expected cyclesTotal to be 1, got %d", m.cyclesTotal)
	}
	if m.cyclesFailed != 0 {
		t.Errorf("expected cyclesFailed to be 0, got %d", m.cyclesFailed)
	}

	// Increment failed cycle
	m.IncrementCycles(true)
	if m.cyclesTotal != 2 {
		t.Errorf("expected cyclesTotal to be 2, got %d", m.cyclesTotal)
	}
	if m.cyclesFailed != 1 {
		t.Errorf("expected cyclesFailed to be 1, got %d", m.cyclesFailed)
	}

	// Increment more successful cycles
	m.IncrementCycles(false)
	m.IncrementCycles(false)
	if m.cyclesTotal != 4 {
		t.Errorf("expected cyclesTotal to be 4, got %d", m.cyclesTotal)
	}
	if m.cyclesFailed != 1 {
		t.Errorf("expected cyclesFailed to be 1, got %d", m.cyclesFailed)
	}
}

func TestIncrementSearches(t *testing.T) {
	m := NewMetrics()

	// Increment successful searches
	m.IncrementSearches("radarr", "missing", false)
	key := "radarr:missing"
	if m.searchesTotal[key] != 1 {
		t.Errorf("expected searchesTotal[%s] to be 1, got %d", key, m.searchesTotal[key])
	}
	if m.searchesFailed[key] != 0 {
		t.Errorf("expected searchesFailed[%s] to be 0, got %d", key, m.searchesFailed[key])
	}

	// Increment failed search
	m.IncrementSearches("radarr", "missing", true)
	if m.searchesTotal[key] != 2 {
		t.Errorf("expected searchesTotal[%s] to be 2, got %d", key, m.searchesTotal[key])
	}
	if m.searchesFailed[key] != 1 {
		t.Errorf("expected searchesFailed[%s] to be 1, got %d", key, m.searchesFailed[key])
	}

	// Different type and category
	m.IncrementSearches("sonarr", "cutoff", false)
	key2 := "sonarr:cutoff"
	if m.searchesTotal[key2] != 1 {
		t.Errorf("expected searchesTotal[%s] to be 1, got %d", key2, m.searchesTotal[key2])
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	m := NewMetrics()

	// Record a request
	m.RecordHTTPRequest("GET", "/api/health", 200, 50*time.Millisecond)

	key := "GET:/api/health:200"
	if m.httpRequests[key] != 1 {
		t.Errorf("expected httpRequests[%s] to be 1, got %d", key, m.httpRequests[key])
	}

	if len(m.httpDurations[key]) != 1 {
		t.Errorf("expected httpDurations[%s] to have 1 entry, got %d", key, len(m.httpDurations[key]))
	}

	expectedDuration := 0.05 // 50ms in seconds
	if m.httpDurations[key][0] != expectedDuration {
		t.Errorf("expected duration to be %f, got %f", expectedDuration, m.httpDurations[key][0])
	}

	// Record another request to the same endpoint
	m.RecordHTTPRequest("GET", "/api/health", 200, 100*time.Millisecond)
	if m.httpRequests[key] != 2 {
		t.Errorf("expected httpRequests[%s] to be 2, got %d", key, m.httpRequests[key])
	}

	if len(m.httpDurations[key]) != 2 {
		t.Errorf("expected httpDurations[%s] to have 2 entries, got %d", key, len(m.httpDurations[key]))
	}

	// Record request to different endpoint
	m.RecordHTTPRequest("POST", "/api/servers", 201, 200*time.Millisecond)
	key2 := "POST:/api/servers:201"
	if m.httpRequests[key2] != 1 {
		t.Errorf("expected httpRequests[%s] to be 1, got %d", key2, m.httpRequests[key2])
	}
}

func TestFormat_PrometheusFormat(t *testing.T) {
	m := NewMetrics()

	// Add some data
	m.IncrementCycles(false)
	m.IncrementCycles(true)
	m.IncrementSearches("radarr", "missing", false)
	m.IncrementSearches("sonarr", "cutoff", true)
	m.RecordHTTPRequest("GET", "/api/health", 200, 50*time.Millisecond)
	m.RecordHTTPRequest("GET", "/api/health", 200, 100*time.Millisecond)

	output := m.Format()

	// Check for required Prometheus format elements
	requiredStrings := []string{
		"# HELP janitarr_uptime_seconds",
		"# TYPE janitarr_uptime_seconds gauge",
		"janitarr_uptime_seconds",
		"# HELP janitarr_cycles_total",
		"# TYPE janitarr_cycles_total counter",
		"janitarr_cycles_total 2",
		"# HELP janitarr_cycles_failed_total",
		"# TYPE janitarr_cycles_failed_total counter",
		"janitarr_cycles_failed_total 1",
		"# HELP janitarr_searches_total",
		"# TYPE janitarr_searches_total counter",
		"janitarr_searches_total{type=\"radarr\",category=\"missing\"} 1",
		"janitarr_searches_total{type=\"sonarr\",category=\"cutoff\"} 1",
		"# HELP janitarr_searches_failed_total",
		"# TYPE janitarr_searches_failed_total counter",
		"janitarr_searches_failed_total{type=\"sonarr\",category=\"cutoff\"} 1",
		"# HELP janitarr_http_requests_total",
		"# TYPE janitarr_http_requests_total counter",
		"janitarr_http_requests_total{method=\"GET\",path=\"/api/health\",status=\"200\"} 2",
		"# HELP janitarr_http_request_duration_seconds",
		"# TYPE janitarr_http_request_duration_seconds histogram",
	}

	for _, s := range requiredStrings {
		if !strings.Contains(output, s) {
			t.Errorf("output missing expected string: %q", s)
		}
	}

	// Check histogram buckets
	histogramBuckets := []string{
		"janitarr_http_request_duration_seconds_bucket{method=\"GET\",path=\"/api/health\",status=\"200\",le=\"0.005\"}",
		"janitarr_http_request_duration_seconds_bucket{method=\"GET\",path=\"/api/health\",status=\"200\",le=\"+Inf\"} 2",
		"janitarr_http_request_duration_seconds_sum{method=\"GET\",path=\"/api/health\",status=\"200\"}",
		"janitarr_http_request_duration_seconds_count{method=\"GET\",path=\"/api/health\",status=\"200\"} 2",
	}

	for _, s := range histogramBuckets {
		if !strings.Contains(output, s) {
			t.Errorf("output missing expected histogram bucket: %q", s)
		}
	}
}

func TestFormat_EmptyMetrics(t *testing.T) {
	m := NewMetrics()

	output := m.Format()

	// Should still have basic metrics
	if !strings.Contains(output, "janitarr_uptime_seconds") {
		t.Error("output missing uptime metric")
	}

	if !strings.Contains(output, "janitarr_cycles_total 0") {
		t.Error("output missing cycles total")
	}

	// Should not have search or HTTP metrics
	if strings.Contains(output, "janitarr_searches_total") {
		t.Error("output should not contain searches when none recorded")
	}

	if strings.Contains(output, "janitarr_http_requests_total") {
		t.Error("output should not contain HTTP requests when none recorded")
	}
}

func TestIncrementCycles_Monotonic(t *testing.T) {
	m := NewMetrics()

	// Test that counters are monotonically increasing
	for i := 0; i < 100; i++ {
		m.IncrementCycles(i%3 == 0) // Fail every 3rd cycle
	}

	if m.cyclesTotal != 100 {
		t.Errorf("expected cyclesTotal to be 100, got %d", m.cyclesTotal)
	}

	expectedFailed := int64(34) // 0, 3, 6, 9, ..., 99 = 34 cycles
	if m.cyclesFailed != expectedFailed {
		t.Errorf("expected cyclesFailed to be %d, got %d", expectedFailed, m.cyclesFailed)
	}
}

func TestRecordHTTPRequest_Labels(t *testing.T) {
	m := NewMetrics()

	// Test different combinations of labels
	tests := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/health", 200},
		{"GET", "/api/health", 500},
		{"POST", "/api/servers", 201},
		{"PUT", "/api/servers/123", 200},
		{"DELETE", "/api/servers/123", 204},
	}

	for _, tt := range tests {
		m.RecordHTTPRequest(tt.method, tt.path, tt.status, 100*time.Millisecond)
	}

	output := m.Format()

	// Verify each label combination appears in output
	for _, tt := range tests {
		expected := "method=\"" + tt.method + "\",path=\"" + tt.path
		if !strings.Contains(output, expected) {
			t.Errorf("output missing expected labels: %q", expected)
		}
	}
}
