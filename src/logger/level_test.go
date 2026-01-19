package logger

import (
	"testing"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{"debug", LevelDebug, "debug"},
		{"info", LevelInfo, "info"},
		{"warn", LevelWarn, "warn"},
		{"error", LevelError, "error"},
		{"unknown", Level(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Level.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Level
		expectErr bool
	}{
		{"debug lowercase", "debug", LevelDebug, false},
		{"info lowercase", "info", LevelInfo, false},
		{"warn lowercase", "warn", LevelWarn, false},
		{"warning lowercase", "warning", LevelWarn, false},
		{"error lowercase", "error", LevelError, false},
		{"debug uppercase", "DEBUG", LevelDebug, false},
		{"info mixed case", "InFo", LevelInfo, false},
		{"warn with spaces", "  warn  ", LevelWarn, false},
		{"invalid level", "invalid", LevelInfo, true},
		{"empty string", "", LevelInfo, true},
		{"trace not supported", "trace", LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLevel(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParseLevel(%q) expected error, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseLevel(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
				}
			}
		})
	}
}
