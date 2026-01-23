package api

import (
	"net/http"

	"github.com/edrobertsrayne/janitarr/src/metrics"
)

// MetricsHandlers handles Prometheus metrics endpoint
type MetricsHandlers struct {
	metrics *metrics.Metrics
}

// NewMetricsHandlers creates a new MetricsHandlers instance
func NewMetricsHandlers(m *metrics.Metrics) *MetricsHandlers {
	return &MetricsHandlers{
		metrics: m,
	}
}

// GetMetrics returns metrics in Prometheus text format
// GET /metrics
func (h *MetricsHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(h.metrics.Format()))
}
