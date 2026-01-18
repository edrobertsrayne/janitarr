package middleware

import (
	"net/http"
	"time"
	"sync"
	"fmt"

	"github.com/go-chi/chi/v5/middleware"
)

// Metrics records basic HTTP request metrics.
type Metrics struct {
	mu           sync.RWMutex
	requestsTotal map[string]int64 // key: method:path:status
	durationSum   map[string]time.Duration // key: method:path
	durationCount map[string]int64 // key: method:path
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		requestsTotal: make(map[string]int64),
		durationSum: make(map[string]time.Duration),
		durationCount: make(map[string]int64),
	}
}

// MetricsMiddleware records HTTP request metrics.
func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		end := time.Since(start)
		routeKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
		statusKey := fmt.Sprintf("%s:%s:%d", r.Method, r.URL.Path, ww.Status())

		m.mu.Lock()
		defer m.mu.Unlock()

		m.requestsTotal[statusKey]++
		m.durationSum[routeKey] += end
		m.durationCount[routeKey]++
	})
}

// GetMetrics returns the collected metrics.
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := make(map[string]interface{})
	data["requests_total"] = m.requestsTotal
	
	avgDurations := make(map[string]time.Duration)
	for k, v := range m.durationSum {
		if m.durationCount[k] > 0 {
			avgDurations[k] = v / time.Duration(m.durationCount[k])
		} else {
			avgDurations[k] = 0
		}
	}
	data["request_duration_avg"] = avgDurations

	return data
}
