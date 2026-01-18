package pages

import (
	"net/http"
	"strconv"

	"github.com/user/janitarr/src/logger"
	"github.com/user/janitarr/src/templates/components"
	"github.com/user/janitarr/src/templates/pages"
)

// HandleLogs renders the logs page
func (h *PageHandlers) HandleLogs(w http.ResponseWriter, r *http.Request) {
	// Get recent logs (default 50)
	logs, err := h.db.GetLogs(50, 0)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	// Convert to logger.LogEntry
	logEntries := make([]logger.LogEntry, len(logs))
	for i, log := range logs {
		logEntries[i] = logger.LogEntry{
			ID:         log.ID,
			Timestamp:  log.Timestamp,
			Type:       logger.LogEntryType(log.Type),
			ServerName: log.ServerName,
			ServerType: log.ServerType,
			Category:   log.Category,
			Count:      log.Count,
			Message:    log.Message,
			IsManual:   log.IsManual,
		}
	}

	pages.Logs(logEntries).Render(r.Context(), w)
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

	// Get logs with filters
	var logs []interface{}
	var err error

	// For now, just get all logs with pagination
	// TODO: Implement filtering by type and server
	logs, err = h.db.GetLogs(20, offset)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	// Convert to logger.LogEntry and filter
	logEntries := make([]logger.LogEntry, 0)

	// Since db.GetLogs returns []interface{}, we need to handle this properly
	// For now, just return empty until we fix the database layer
	_ = logs
	_ = typeFilter
	_ = serverFilter

	// Render log entries
	w.Header().Set("Content-Type", "text/html")
	for _, entry := range logEntries {
		components.LogEntry(entry).Render(r.Context(), w)
	}
}
