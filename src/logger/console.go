package logger

import (
	"io"
	"os"

	"github.com/charmbracelet/log"
)

// ConsoleLogger wraps charmbracelet/log for colorized console output.
type ConsoleLogger struct {
	logger *log.Logger
	level  Level
}

// NewConsoleLogger creates a new ConsoleLogger.
// In development mode (isDev=true), logs go to stdout.
// In production mode (isDev=false), logs go to stderr.
func NewConsoleLogger(level Level, isDev bool) *ConsoleLogger {
	var output io.Writer = os.Stderr
	if isDev {
		output = os.Stdout
	}

	logger := log.NewWithOptions(output, log.Options{
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Level:           toCharmLevel(level),
	})

	return &ConsoleLogger{
		logger: logger,
		level:  level,
	}
}

// toCharmLevel converts our Level to charmbracelet/log.Level.
func toCharmLevel(level Level) log.Level {
	switch level {
	case LevelDebug:
		return log.DebugLevel
	case LevelInfo:
		return log.InfoLevel
	case LevelWarn:
		return log.WarnLevel
	case LevelError:
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

// Debug logs a debug message with structured key-value pairs.
func (c *ConsoleLogger) Debug(msg string, keyvals ...interface{}) {
	if c.level <= LevelDebug {
		c.logger.Debug(msg, keyvals...)
	}
}

// Info logs an info message with structured key-value pairs.
func (c *ConsoleLogger) Info(msg string, keyvals ...interface{}) {
	if c.level <= LevelInfo {
		c.logger.Info(msg, keyvals...)
	}
}

// Warn logs a warning message with structured key-value pairs.
func (c *ConsoleLogger) Warn(msg string, keyvals ...interface{}) {
	if c.level <= LevelWarn {
		c.logger.Warn(msg, keyvals...)
	}
}

// Error logs an error message with structured key-value pairs.
func (c *ConsoleLogger) Error(msg string, keyvals ...interface{}) {
	if c.level <= LevelError {
		c.logger.Error(msg, keyvals...)
	}
}

// SetLevel changes the log level.
func (c *ConsoleLogger) SetLevel(level Level) {
	c.level = level
	c.logger.SetLevel(toCharmLevel(level))
}
