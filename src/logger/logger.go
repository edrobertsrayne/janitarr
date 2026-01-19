package logger

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid" // Import uuid
)

// Logger handles application logging to the database and broadcasting to subscribers.
type Logger struct {
	storer      LogStorer // Use the interface here
	console     *ConsoleLogger
	level       Level
	mu          sync.RWMutex
	subscribers map[chan LogEntry]bool
}

// NewLogger creates a new Logger.
func NewLogger(storer LogStorer, level Level, isDev bool) *Logger { // Accept LogStorer interface
	return &Logger{
		storer:      storer,
		console:     NewConsoleLogger(level, isDev),
		level:       level,
		subscribers: make(map[chan LogEntry]bool),
	}
}

// AddLog adds a new log entry to the storer and broadcasts it.
func (l *Logger) AddLog(entry LogEntry) *LogEntry {
	if entry.ID == "" { // Assign ID if not already set (e.g., from manual creation in tests)
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() { // Assign Timestamp if not already set
		entry.Timestamp = time.Now().UTC()
	}
	_ = l.storer.AddLog(entry) // Add error handling if necessary
	l.broadcast(&entry)
	return &entry
}

// LogCycleStart logs the start of an automation cycle.
func (l *Logger) LogCycleStart(isManual bool) *LogEntry {
	entry := LogEntry{
		Type:     LogTypeCycleStart,
		Message:  "Automation cycle started.",
		IsManual: isManual,
	}

	// Console log at info level
	l.console.Info("Automation cycle started", "manual", isManual)

	return l.AddLog(entry)
}

// LogCycleEnd logs the end of an automation cycle.
func (l *Logger) LogCycleEnd(totalSearches, failures int, isManual bool) *LogEntry {
	entry := LogEntry{
		Type:     LogTypeCycleEnd,
		Message:  "Automation cycle finished.",
		IsManual: isManual,
		Count:    totalSearches, // Store total searches in count
	}

	// Console log at info level
	l.console.Info("Automation cycle finished",
		"searches", totalSearches,
		"failures", failures,
		"manual", isManual)

	return l.AddLog(entry)
}

// LogDetectionComplete logs the completion of detection for a server.
func (l *Logger) LogDetectionComplete(serverName, serverType string, missing, cutoffUnmet int) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeDetection,
		ServerName: serverName,
		ServerType: serverType,
		Message:    "Detection complete.",
		Count:      missing + cutoffUnmet, // Store total in count for simple querying
	}

	// Console log at info level
	l.console.Info("Detection complete",
		"server", serverName,
		"missing", missing,
		"cutoff_unmet", cutoffUnmet)

	return l.AddLog(entry)
}

// LogSearches logs triggered searches.
func (l *Logger) LogSearches(serverName, serverType, category string, count int, isManual bool) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeSearch,
		ServerName: serverName,
		ServerType: serverType,
		Category:   category,
		Count:      count,
		Message:    "Triggered searches.",
		IsManual:   isManual,
	}

	// Console log at info level
	l.console.Info("Triggered searches",
		"server", serverName,
		"type", serverType,
		"category", category,
		"count", count,
		"manual", isManual)

	return l.AddLog(entry)
}

// LogMovieSearch logs a movie search with detailed metadata.
func (l *Logger) LogMovieSearch(serverName, serverType, title string, year int, qualityProfile, category string) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeSearch,
		ServerName: serverName,
		ServerType: serverType,
		Category:   category,
		Message:    "Search triggered.",
		Count:      1,
	}

	// Console log at info level with detailed metadata
	l.console.Info("Search triggered",
		"title", title,
		"year", year,
		"quality", qualityProfile,
		"server", serverName,
		"category", category)

	return l.AddLog(entry)
}

// LogEpisodeSearch logs an episode search with detailed metadata.
func (l *Logger) LogEpisodeSearch(serverName, serverType, seriesTitle, episodeTitle string, season, episode int, qualityProfile, category string) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeSearch,
		ServerName: serverName,
		ServerType: serverType,
		Category:   category,
		Message:    "Search triggered.",
		Count:      1,
	}

	// Console log at info level with detailed metadata
	episodeStr := fmt.Sprintf("S%02dE%02d", season, episode)
	l.console.Info("Search triggered",
		"series", seriesTitle,
		"episode", episodeStr,
		"title", episodeTitle,
		"quality", qualityProfile,
		"server", serverName,
		"category", category)

	return l.AddLog(entry)
}

// LogServerError logs an error related to a server.
func (l *Logger) LogServerError(serverName, serverType, reason string) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeError,
		ServerName: serverName,
		ServerType: serverType,
		Message:    reason,
	}

	// Console log at error level
	l.console.Error("Server error",
		"server", serverName,
		"type", serverType,
		"reason", reason)

	return l.AddLog(entry)
}

// LogSearchError logs an error related to a search.
func (l *Logger) LogSearchError(serverName, serverType, category, reason string) *LogEntry {
	entry := LogEntry{
		Type:       LogTypeError,
		ServerName: serverName,
		ServerType: serverType,
		Category:   category,
		Message:    reason,
	}

	// Console log at error level
	l.console.Error("Search error",
		"server", serverName,
		"type", serverType,
		"category", category,
		"reason", reason)

	return l.AddLog(entry)
}

// Subscribe returns a channel that receives log entries.
func (l *Logger) Subscribe() <-chan LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan LogEntry, 100)
	l.subscribers[ch] = true
	return ch
}

// Unsubscribe removes a channel from the subscribers.
func (l *Logger) Unsubscribe(ch chan LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.subscribers[ch]; ok {
		delete(l.subscribers, ch)
		close(ch)
	}
}

func (l *Logger) broadcast(entry *LogEntry) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for ch := range l.subscribers {
		select {
		case ch <- *entry:
		default:
			// Don't block if a subscriber is slow.
		}
	}
}

// Debug logs a debug message to console only (not stored in database).
// Used for development mode verbose logging.
func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.console.Debug(msg, keyvals...)
}

// Info logs an info message to console only (not stored in database).
func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.console.Info(msg, keyvals...)
}

// Error logs an error message to console only (not stored in database).
func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.console.Error(msg, keyvals...)
}
