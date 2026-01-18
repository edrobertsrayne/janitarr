package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/janitarr/src/database"
)

// StatsHandlers provides handlers for statistics API endpoints.
type StatsHandlers struct {
	DB *database.DB
}

// NewStatsHandlers creates a new StatsHandlers instance.
func NewStatsHandlers(db *database.DB) *StatsHandlers {
	return &StatsHandlers{DB: db}
}

// GetSummaryStats returns a summary of system-wide statistics.
func (h *StatsHandlers) GetSummaryStats(w http.ResponseWriter, r *http.Request) {
	stats := h.DB.GetSystemStats()
	jsonSuccess(w, stats)
}

// GetServerStats returns statistics for a specific server.
func (h *StatsHandlers) GetServerStats(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "id")
	if serverID == "" {
		jsonError(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	stats := h.DB.GetServerStats(serverID)
	if stats.ServerName == "" { // Assuming ServerName is set if found
		jsonError(w, "Server not found", http.StatusNotFound)
		return
	}
	jsonSuccess(w, stats)
}
