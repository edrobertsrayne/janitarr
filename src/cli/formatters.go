package cli

import (
	"fmt"
	"strings"
	"time" // Added for time.RFC3339 in log formatting if needed, though not for server table

	"github.com/edrobertsrayne/janitarr/src/database"
	"github.com/edrobertsrayne/janitarr/src/logger"
	"github.com/edrobertsrayne/janitarr/src/services" // Imported for ServerInfo, etc.
)

// Color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func success(msg string) string  { return colorGreen + "✓ " + msg + colorReset }
func errorMsg(msg string) string { return colorRed + "✗ " + msg + colorReset }
func warning(msg string) string  { return colorYellow + "⚠ " + msg + colorReset }
func info(msg string) string     { return colorCyan + "ℹ " + msg + colorReset }
func header(msg string) string   { return colorBold + msg + colorReset }

// keyValue formats a key-value pair for console output.
func keyValue(key, value string) string {
	return fmt.Sprintf("  %s: %s", key, value)
}

// showProgress displays a progress message on a single line.
func showProgress(msg string) {
	fmt.Printf("\r%s... \x1b[K", msg)
}

// clearLine clears the current console line.
func clearLine() {
	fmt.Print("\r\x1b[K")
}

// hideCursor hides the console cursor.
func hideCursor() {
	fmt.Print("\x1b[?25l")
}

// showCursor shows the console cursor.
func showCursor() {
	fmt.Print("\x1b[?25h")
}

// formatServerTable formats a slice of ServerInfo into a human-readable table.
func formatServerTable(servers []services.ServerInfo) string {
	if len(servers) == 0 {
		return info("No servers configured.")
	}

	var sb strings.Builder
	sb.WriteString(header("Configured Servers") + "\n")
	sb.WriteString("\n")

	// Calculate column widths
	nameWidth := 4 // "Name"
	for _, s := range servers {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}
	urlWidth := 3 // "URL"
	for _, s := range servers {
		if len(s.URL) > urlWidth {
			urlWidth = len(s.URL)
		}
	}

	// Header
	sb.WriteString(fmt.Sprintf("% -*s  %-6s  %-*s  %s\n", nameWidth, "Name", "Type", urlWidth, "URL", "Enabled"))
	sb.WriteString(fmt.Sprintf("%s  %s  %s  %s\n", strings.Repeat("-", nameWidth), strings.Repeat("-", 6), strings.Repeat("-", urlWidth), strings.Repeat("-", 7)))

	// Rows
	for _, s := range servers {
		name := s.Name
		serverType := strings.Title(s.Type) // Capitalize type for display
		url := s.URL
		enabledText := ""
		if s.Enabled {
			enabledText = success("Yes")
		} else {
			enabledText = warning("No")
		}
		sb.WriteString(fmt.Sprintf("% -*s  %-6s  %-*s  %s\n", nameWidth, name, serverType, urlWidth, url, enabledText))
	}
	return sb.String()
}

// formatLogTable formats a slice of logger.LogEntry into a human-readable table.
func formatLogTable(logs []logger.LogEntry) string {
	if len(logs) == 0 {
		return info("No log entries.")
	}

	var sb strings.Builder
	sb.WriteString(header("Activity Logs") + "\n")
	sb.WriteString("\n")

	// Calculate column widths
	// Max width for timestamp (RFC3339) is usually around 24-29 chars
	timestampWidth := 29
	typeWidth := 15  // Max length of LogEntryType enum values
	serverWidth := 6 // "Server"
	for _, l := range logs {
		if l.ServerName != "" && len(l.ServerName) > serverWidth {
			serverWidth = len(l.ServerName)
		}
	}
	messageWidth := 70 // Default, can be adjusted or dynamic

	// Header
	sb.WriteString(fmt.Sprintf("% -*s  %-*s  %-*s  %s\n", timestampWidth, "Timestamp", typeWidth, "Type", serverWidth, "Server", "Message"))
	sb.WriteString(fmt.Sprintf("%s  %s  %s  %s\n", strings.Repeat("-", timestampWidth), strings.Repeat("-", typeWidth), strings.Repeat("-", serverWidth), strings.Repeat("-", messageWidth)))

	// Rows
	for _, l := range logs {
		timestamp := l.Timestamp.Format(time.RFC3339)
		logType := string(l.Type)
		message := l.Message

		// Apply color based on log type
		switch l.Type {
		case logger.LogTypeError:
			message = errorMsg(message)
			logType = errorMsg(logType)
		case logger.LogTypeCycleStart, logger.LogTypeCycleEnd:
			logType = info(logType)
		case logger.LogTypeSearch:
			logType = success(logType)
		}

		serverName := l.ServerName
		if l.ServerName == "" {
			serverName = "N/A"
		}

		// Truncate message if too long
		if len(message) > messageWidth {
			message = message[:messageWidth-3] + "..."
		}

		sb.WriteString(fmt.Sprintf("% -*s  %-*s  %-*s  %s\n",
			timestampWidth, timestamp,
			typeWidth, logType,
			serverWidth, serverName,
			message))

		// Add additional details for search logs
		if l.Type == logger.LogTypeSearch && l.ServerName != "" {
			detailMsg := fmt.Sprintf("  └─ %s (%s) - %s: %d items",
				strings.Title(string(l.ServerType)), l.ServerName,
				l.Category, l.Count)
			sb.WriteString(info(detailMsg) + "\n")
		}
	}
	return sb.String()
}

// formatConfigTable formats an AppConfig into human-readable key-value pairs.
func formatConfigTable(config *database.AppConfig) string {
	var sb strings.Builder
	sb.WriteString(header("Configuration") + "\n")
	sb.WriteString("\n")

	sb.WriteString(colorBold + "Schedule:" + colorReset + "\n")
	enabledText := warning("No")
	if config.Schedule.Enabled {
		enabledText = success("Yes")
	}
	sb.WriteString(keyValue("Enabled", enabledText) + "\n")
	sb.WriteString(keyValue("Interval", fmt.Sprintf("%d hours", config.Schedule.IntervalHours)) + "\n")
	sb.WriteString("\n")

	sb.WriteString(colorBold + "Search Limits:" + colorReset + "\n")
	sb.WriteString(keyValue("Missing Movies", formatLimit(config.SearchLimits.MissingMoviesLimit)) + "\n")
	sb.WriteString(keyValue("Missing Episodes", formatLimit(config.SearchLimits.MissingEpisodesLimit)) + "\n")
	sb.WriteString(keyValue("Cutoff Movies", formatLimit(config.SearchLimits.CutoffMoviesLimit)) + "\n")
	sb.WriteString(keyValue("Cutoff Episodes", formatLimit(config.SearchLimits.CutoffEpisodesLimit)) + "\n")

	return sb.String()
}

func formatLimit(limit int) string {
	if limit == 0 {
		return warning("Disabled")
	}
	return fmt.Sprintf("%d items", limit)
}
