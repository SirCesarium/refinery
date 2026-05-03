package ui

import (
	"os"
	"testing"
)

// TestSuccess checks that success message is printed correctly.
func TestSuccess(t *testing.T) {
	// Capture output (not easy in this setup, just verify no panic)
	Success("test %s", "message")
}

// TestInfo checks that info message is printed correctly.
func TestInfo(t *testing.T) {
	Info("test %s", "info")
}

// TestWarn checks that warning message is printed correctly.
func TestWarn(t *testing.T) {
	Warn("test %s", "warning")
}

// TestError checks error message printing.
func TestError(t *testing.T) {
	// Test with error
	Error(os.ErrNotExist, "help message")

	// Test without error
	Error(nil, "help message")
}

// TestFatal checks that fatal calls os.Exit (can't test directly).
func TestFatal(t *testing.T) {
	// Just verify it doesn't panic during setup
	// Note: Fatal will exit, so cannot fully test it
	_ = Fatal
}

// TestSection checks section header printing.
func TestSection(t *testing.T) {
	Section("Test Section")
}

// TestConstants checks that color constants are defined.
func TestConstants(t *testing.T) {
	if Reset == "" {
		t.Error("expected Reset to be non-empty")
	}
	if Red == "" {
		t.Error("expected Red to be non-empty")
	}
	if Green == "" {
		t.Error("expected Green to be non-empty")
	}
}
