package ui

import (
	"os"
	"testing"
)

func TestSuccess(t *testing.T) {
	Success("test %s", "message")
}

func TestInfo(t *testing.T) {
	Info("test %s", "info")
}

func TestWarn(t *testing.T) {
	Warn("test %s", "warning")
}

func TestError(t *testing.T) {
	Error(os.ErrNotExist, "help message")
	Error(nil, "help message")
}

func TestFatal(t *testing.T) {
	_ = Fatal
}

func TestSection(t *testing.T) {
	Section("Test Section")
}

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
