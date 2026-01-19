package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestNewConsoleLogger(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		isDev bool
	}{
		{"debug dev mode", LevelDebug, true},
		{"info production mode", LevelInfo, false},
		{"warn production mode", LevelWarn, false},
		{"error dev mode", LevelError, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewConsoleLogger(tt.level, tt.isDev)
			if logger == nil {
				t.Fatal("NewConsoleLogger returned nil")
			}
			if logger.level != tt.level {
				t.Errorf("level = %v, want %v", logger.level, tt.level)
			}
		})
	}
}

func TestConsoleLogger_LevelFiltering(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := &ConsoleLogger{
		logger: log.NewWithOptions(&buf, log.Options{
			ReportTimestamp: false, // Disable timestamp for easier testing
		}),
		level: LevelInfo,
	}

	// Debug should be filtered out
	logger.Debug("debug message", "key", "value")
	if buf.Len() > 0 {
		t.Errorf("Debug message should be filtered at Info level, got: %s", buf.String())
	}

	// Info should pass through
	buf.Reset()
	logger.Info("info message", "key", "value")
	if buf.Len() == 0 {
		t.Error("Info message should not be filtered at Info level")
	}
	if !strings.Contains(buf.String(), "info message") {
		t.Errorf("expected 'info message' in output, got: %s", buf.String())
	}

	// Warn should pass through
	buf.Reset()
	logger.Warn("warn message", "key", "value")
	if buf.Len() == 0 {
		t.Error("Warn message should not be filtered at Info level")
	}

	// Error should pass through
	buf.Reset()
	logger.Error("error message", "key", "value")
	if buf.Len() == 0 {
		t.Error("Error message should not be filtered at Info level")
	}
}

func TestConsoleLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := &ConsoleLogger{
		logger: log.NewWithOptions(&buf, log.Options{
			ReportTimestamp: false,
		}),
		level: LevelError,
	}

	// Info should be filtered at Error level
	logger.Info("info message")
	if buf.Len() > 0 {
		t.Error("Info message should be filtered at Error level")
	}

	// Change to Debug level
	logger.SetLevel(LevelDebug)

	// Now info should pass through
	buf.Reset()
	logger.Info("info message after level change")
	if buf.Len() == 0 {
		t.Error("Info message should not be filtered after changing to Debug level")
	}
}

func TestToCharmLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected log.Level
	}{
		{"debug", LevelDebug, log.DebugLevel},
		{"info", LevelInfo, log.InfoLevel},
		{"warn", LevelWarn, log.WarnLevel},
		{"error", LevelError, log.ErrorLevel},
		{"unknown defaults to info", Level(999), log.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCharmLevel(tt.level)
			if result != tt.expected {
				t.Errorf("toCharmLevel(%v) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}
