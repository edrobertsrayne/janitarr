package pages

import (
	"fmt"
	"net/http"
	"time"

	"github.com/user/janitarr/src/templates/pages"
)

// HandleDashboard renders the dashboard page
func (h *PageHandlers) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get scheduler status
	schedulerStatus := h.scheduler.GetStatus()

	// Get server count
	servers, err := h.db.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	// Convert servers to display format
	serverDisplays := make([]pages.ServerDisplay, len(servers))
	for i, srv := range servers {
		serverDisplays[i] = pages.ServerDisplay{
			Name:    srv.Name,
			Type:    srv.Type,
			URL:     srv.URL,
			Enabled: srv.Enabled,
		}
	}

	// Get recent logs (last 10)
	logs, err := h.db.GetLogs(10, 0)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	// Convert logs to display format
	logDisplays := make([]pages.LogDisplay, len(logs))
	for i, log := range logs {
		timestamp, _ := time.Parse(time.RFC3339, log.Timestamp)
		logDisplays[i] = pages.LogDisplay{
			Timestamp: formatRelativeTime(timestamp),
			Type:      log.Type,
			Message:   log.Message,
			IsError:   log.Type == "error",
		}
	}

	// Calculate total searches and failures from logs
	totalSearches, totalFailures := calculateStats(logs)

	data := pages.DashboardData{
		ServerCount:     len(servers),
		TotalSearches:   totalSearches,
		TotalFailures:   totalFailures,
		SchedulerStatus: &schedulerStatus,
		RecentLogs:      logDisplays,
		Servers:         serverDisplays,
	}

	// Render the dashboard
	pages.Dashboard(data).Render(r.Context(), w)
}

// HandleStatsPartial handles the htmx stats refresh
func (h *PageHandlers) HandleStatsPartial(w http.ResponseWriter, r *http.Request) {
	// Get scheduler status
	schedulerStatus := h.scheduler.GetStatus()

	// Get server count
	servers, err := h.db.ListServers()
	if err != nil {
		http.Error(w, "Failed to load servers", http.StatusInternalServerError)
		return
	}

	// Get recent logs to calculate stats
	logs, err := h.db.GetLogs(100, 0)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	totalSearches, totalFailures := calculateStats(logs)

	// Return just the stats cards HTML (would need to extract this into a separate templ component)
	// For now, render as JSON or simple HTML
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8" hx-get="/partials/stats" hx-trigger="every 30s" hx-swap="outerHTML">
			<!-- Stats cards would go here -->
		</div>
	`)
}

// HandleRecentActivityPartial handles the htmx recent activity refresh
func (h *PageHandlers) HandleRecentActivityPartial(w http.ResponseWriter, r *http.Request) {
	// Get recent logs (last 10)
	logs, err := h.db.GetLogs(10, 0)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)
		return
	}

	// Convert logs to display format
	logDisplays := make([]pages.LogDisplay, len(logs))
	for i, log := range logs {
		timestamp, _ := time.Parse(time.RFC3339, log.Timestamp)
		logDisplays[i] = pages.LogDisplay{
			Timestamp: formatRelativeTime(timestamp),
			Type:      log.Type,
			Message:   log.Message,
			IsError:   log.Type == "error",
		}
	}

	// Return HTML for recent activity
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="space-y-4">`)
	for _, log := range logDisplays {
		isErrorClass := ""
		textClass := "text-gray-800 dark:text-gray-200"
		if log.IsError {
			isErrorClass = "bg-red-50 dark:bg-red-900/20"
			textClass = "text-red-800 dark:text-red-200"
		} else {
			isErrorClass = "bg-gray-50 dark:bg-gray-900/20"
		}
		fmt.Fprintf(w, `
			<div class="p-4 rounded-lg %s">
				<p class="text-sm %s">%s</p>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">%s</p>
			</div>
		`, isErrorClass, textClass, log.Message, log.Timestamp)
	}
	fmt.Fprintf(w, `</div>`)
}

// Helper functions

func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, pluralize(minutes))
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, pluralize(hours))
	} else {
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d day%s ago", days, pluralize(days))
	}
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

type logWithType interface {
	GetType() string
}

func calculateStats(logs []interface{}) (totalSearches, totalFailures int) {
	// This is a simplified calculation - in reality, you'd want to:
	// 1. Query specific log types (search, error)
	// 2. Sum counts from log entries
	// 3. Filter by time range (e.g., last cycle)

	// For now, return placeholder values
	totalSearches = 0
	totalFailures = 0

	// Count based on log types
	for _, log := range logs {
		if logEntry, ok := log.(logWithType); ok {
			if logEntry.GetType() == "search" {
				totalSearches++
			} else if logEntry.GetType() == "error" {
				totalFailures++
			}
		}
	}

	return
}
