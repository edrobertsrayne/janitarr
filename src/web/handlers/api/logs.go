package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/user/janitarr/src/database"
	"github.com/user/janitarr/src/logger"
)

// LogHandlers provides handlers for log-related API endpoints.
type LogHandlers struct {
	DB *database.DB
}

// NewLogHandlers creates a new LogHandlers instance.
func NewLogHandlers(db *database.DB) *LogHandlers {
	return &LogHandlers{DB: db}
}

// ListLogs returns a list of log entries with pagination and filtering.
func (h *LogHandlers) ListLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	typeFilter := r.URL.Query().Get("type")
	serverFilter := r.URL.Query().Get("server")

	limit := 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var logTypeFilterPtr *string
	if typeFilter != "" {
		logTypeFilterPtr = &typeFilter
	}

	var serverFilterPtr *string
	if serverFilter != "" {
		serverFilterPtr = &serverFilter
	}

	logs, err := h.DB.GetLogs(ctx, limit, offset, logTypeFilterPtr, serverFilterPtr)
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to retrieve logs: %v", err), http.StatusInternalServerError)
		return
	}

	jsonSuccess(w, logs)
}

// ClearLogs removes all log entries from the database.
func (h *LogHandlers) ClearLogs(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.ClearLogs(); err != nil {
		jsonError(w, fmt.Sprintf("Failed to clear logs: %v", err), http.StatusInternalServerError)
		return
	}
	jsonMessage(w, "All logs cleared successfully", http.StatusOK)
}

// ExportLogs exports log entries as JSON or CSV.
func (h *LogHandlers) ExportLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // Default to JSON
	}

	logs, err := h.DB.GetLogs(ctx, 0, 0, nil, nil) // Fetch all logs for export
	if err != nil {
		jsonError(w, fmt.Sprintf("Failed to retrieve logs for export: %v", err), http.StatusInternalServerError)
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=\"janitarr_logs.json\"")
		json.NewEncoder(w).Encode(logs)
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=\"janitarr_logs.csv\"")

		// Write CSV header
		_, _ = w.Write([]byte("ID,Timestamp,Type,ServerName,ServerType,Category,Count,Message,IsManual\n"))
		for _, entry := range logs {
			// Basic CSV escaping for simplicity, may need more robust solution for complex strings
			message := strconv.Quote(entry.Message)
			fmt.Fprintf(w, "%s,%s,%s,%s,%s,%s,%d,%s,%t\n",
				entry.ID,
				entry.Timestamp.Format(time.RFC3339),
				entry.Type,
				entry.ServerName,
				entry.ServerType,
				entry.Category,
				entry.Count,
				message,
				entry.IsManual,
			)
		}
	default:
		jsonError(w, "Invalid export format. Supported: json, csv", http.StatusBadRequest)
	}
}
