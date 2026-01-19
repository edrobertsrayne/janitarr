package forms

import (
	"testing"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// TestImports verifies that charmbracelet/huh and golang.org/x/term are available
func TestImports(t *testing.T) {
	// Test huh import
	_ = huh.NewForm()

	// Test term import
	_ = term.IsTerminal(0)

	t.Log("All imports verified successfully")
}
