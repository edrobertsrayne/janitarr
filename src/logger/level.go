package logger

import (
	"fmt"
	"strings"
)

// Level represents the severity level of a log entry.
type Level int

const (
	// LevelDebug is the debug log level.
	LevelDebug Level = iota
	// LevelInfo is the info log level.
	LevelInfo
	// LevelWarn is the warning log level.
	LevelWarn
	// LevelError is the error log level.
	LevelError
)

// String returns the string representation of the Level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseLevel parses a string into a Level.
// Valid values are: "debug", "info", "warn", "error" (case-insensitive).
// Returns an error if the level is invalid.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, fmt.Errorf("invalid log level: %q (valid values: debug, info, warn, error)", s)
	}
}
