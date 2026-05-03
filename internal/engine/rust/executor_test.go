package rust

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

// TestGetLinkerFromConfig tests linker extraction from target config.
func TestGetLinkerFromConfig(t *testing.T) {
	e := &RustEngine{}

	// Test with linker set
	tCfg := &config.TargetConfig{
		LangOpts: map[string]any{
			"linker": "my-linker",
		},
	}
	if linker := e.getLinkerFromConfig(tCfg, "linux", "x86_64"); linker != "my-linker" {
		t.Errorf("expected 'my-linker', got '%s'", linker)
	}

	// Test without linker
	tCfg2 := &config.TargetConfig{}
	if linker := e.getLinkerFromConfig(tCfg2, "linux", "x86_64"); linker != "" {
		t.Errorf("expected empty string, got '%s'", linker)
	}
}

// TestGetDefaultLinker tests default linker selection.
func TestGetDefaultLinker(t *testing.T) {
	e := &RustEngine{}

	// Test linux aarch64
	if linker := e.getDefaultLinker("linux", "aarch64"); linker != "aarch64-linux-gnu-gcc" {
		t.Errorf("expected 'aarch64-linux-gnu-gcc', got '%s'", linker)
	}

	// Test linux i686
	if linker := e.getDefaultLinker("linux", "i686"); linker != "i686-linux-gnu-gcc" {
		t.Errorf("expected 'i686-linux-gnu-gcc', got '%s'", linker)
	}

	// Test other OS (should be empty)
	if linker := e.getDefaultLinker("darwin", "x86_64"); linker != "" {
		t.Errorf("expected empty string, got '%s'", linker)
	}
}

// TestSetupLinkerEnv tests linker environment variable setup.
func TestSetupLinkerEnv(t *testing.T) {
	e := &RustEngine{}

	// Test with empty linker (should not error)
	if err := e.setupLinkerEnv("", "x86_64", "x86_64-unknown-linux-gnu"); err != nil {
		t.Errorf("expected no error with empty linker, got: %v", err)
	}
}

// TestSetupDefaultMacOSDeployment tests macOS deployment target setup.
func TestSetupDefaultMacOSDeployment(t *testing.T) {
	e := &RustEngine{}

	// Test when env is not set
	// Note: This is hard to test without mocking os.Setenv
	// Just verify it doesn't panic
	if err := e.setupDefaultMacOSDeployment(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
