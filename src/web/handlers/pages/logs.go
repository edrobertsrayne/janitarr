package pages

import (
	"net/http"
	"strconv"

	"github.com/edrobertsrayne/janitarr/src/logger"
	"github.com/edrobertsrayne/janitarr/src/templates/components"
	"github.com/edrobertsrayne/janitarr/src/templates/pages"
)

// HandleLogs renders the logs page
func (h *PageHandlers) HandleLogs(w http.ResponseWriter, r *http.Request) {
	// Get recent logs (default 50)
	logs, err := h.db.GetLogs(r.Context(), 50, 0, logger.LogFilters{})
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	pages.Logs(logs).Render(r.Context(), w)
}

// HandleLogEntriesPartial handles the htmx log entries refresh
func (h *PageHandlers) HandleLogEntriesPartial(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	typeFilter := r.URL.Query().Get("type")
	serverFilter := r.URL.Query().Get("server")
	operationFilter := r.URL.Query().Get("operation")
	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")

	// Prepare filters
	filters := logger.LogFilters{}
	if typeFilter != "" {
		filters.Type = &typeFilter
	}
	if serverFilter != "" {
		filters.Server = &serverFilter
	}
	if operationFilter != "" {
		filters.Operation = &operationFilter
	}
	if fromDate != "" {
		filters.FromDate = &fromDate
	}
	if toDate != "" {
		filters.ToDate = &toDate
	}

	// Get logs with filters
	logs, err := h.db.GetLogs(r.Context(), 20, offset, filters)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	logEntries := logs

	// Render log entries
	w.Header().Set("Content-Type", "text/html")
	for _, entry := range logEntries {
		components.LogEntry(entry).Render(r.Context(), w)
	}
}
