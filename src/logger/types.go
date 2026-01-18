package logger

import (
	"context"
	"time"
)

// LogEntryType defines the type of a log entry.
type LogEntryType string

const (
	// LogTypeCycleStart indicates the start of an automation cycle.
	LogTypeCycleStart LogEntryType = "cycle_start"
	// LogTypeCycleEnd indicates the end of an automation cycle.
	LogTypeCycleEnd LogEntryType = "cycle_end"
	// LogTypeSearch indicates a search was triggered.
	LogTypeSearch LogEntryType = "search"
	// LogTypeError indicates an error occurred.
	LogTypeError LogEntryType = "error"
)

// LogEntry represents a single log entry.
type LogEntry struct {
	ID         string       `json:"id"`
	Timestamp  time.Time    `json:"timestamp"`
	Type       LogEntryType `json:"type"`
	ServerName string       `json:"serverName,omitempty"`
	ServerType string       `json:"serverType,omitempty"`
	Category   string       `json:"category,omitempty"`
	Count      int          `json:"count,omitempty"`
	Message    string       `json:"message"`
	IsManual   bool         `json:"isManual"`
}

// LogStorer defines the interface for storing and retrieving log entries.
// This interface is used by the logger service to decouple it from the concrete database implementation.
type LogStorer interface {
	AddLog(entry LogEntry) error
	GetLogs(ctx context.Context, limit, offset int, logTypeFilter, serverNameFilter *string) ([]LogEntry, error)
	ClearLogs() error
}