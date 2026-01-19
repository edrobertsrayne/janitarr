package logger

import (
	"testing"

	"github.com/charmbracelet/log"
)

// TestCharmLogImport verifies that charmbracelet/log can be imported.
func TestCharmLogImport(t *testing.T) {
	// Create a basic logger to verify the import works
	logger := log.New(nil)
	if logger == nil {
		t.Fatal("failed to create charmbracelet/log logger")
	}
}
