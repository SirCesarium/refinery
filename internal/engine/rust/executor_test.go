package rust

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

func TestGetLinkerFromConfig(t *testing.T) {
	e := &RustEngine{}

	tCfg := &config.TargetConfig{
		LangOpts: map[string]any{
			"linker": "my-linker",
		},
	}
	if linker := e.getLinkerFromConfig(tCfg, "linux", "x86_64"); linker != "my-linker" {
		t.Errorf("expected 'my-linker', got '%s'", linker)
	}

	tCfg2 := &config.TargetConfig{}
	if linker := e.getLinkerFromConfig(tCfg2, "linux", "x86_64"); linker != "" {
		t.Errorf("expected empty string, got '%s'", linker)
	}

	if linker := e.getLinkerFromConfig(nil, "linux", "x86_64"); linker != "" {
		t.Errorf("expected empty string for nil config, got '%s'", linker)
	}
}

func TestGetDefaultLinker(t *testing.T) {
	e := &RustEngine{}

	if linker := e.getDefaultLinker("linux", "aarch64"); linker != "aarch64-linux-gnu-gcc" {
		t.Errorf("expected 'aarch64-linux-gnu-gcc', got '%s'", linker)
	}

	if linker := e.getDefaultLinker("linux", "i686"); linker != "i686-linux-gnu-gcc" {
		t.Errorf("expected 'i686-linux-gnu-gcc', got '%s'", linker)
	}

	if linker := e.getDefaultLinker("darwin", "x86_64"); linker != "" {
		t.Errorf("expected empty string, got '%s'", linker)
	}

	if linker := e.getDefaultLinker("linux", "x86_64"); linker != "" {
		t.Errorf("expected empty string for x86_64, got '%s'", linker)
	}
}

func TestSetupLinkerEnv(t *testing.T) {
	e := &RustEngine{}

	if err := e.setupLinkerEnv("", "x86_64", "x86_64-unknown-linux-gnu"); err != nil {
		t.Errorf("expected no error with empty linker, got: %v", err)
	}

	if err := e.setupLinkerEnv("aarch64-linux-gnu-gcc", "aarch64", "aarch64-unknown-linux-gnu"); err != nil {
		t.Errorf("expected no error with valid linker, got: %v", err)
	}

	if err := e.setupLinkerEnv("x86_64-linux-gnu-gcc", "aarch64", "aarch64-unknown-linux-gnu"); err != nil {
		t.Errorf("expected no error even with mismatched linker, got: %v", err)
	}
}

func TestSetupMacOSVars(t *testing.T) {
	e := &RustEngine{}

	if envVars := e.setupMacOSVars(nil); envVars != nil {
		t.Error("expected nil for nil target config")
	}

	tCfg := &config.TargetConfig{
		OS:    "linux",
		Archs: []string{"x86_64"},
	}
	if envVars := e.setupMacOSVars(tCfg); envVars != nil {
		t.Errorf("expected nil for non-darwin target, got %v", envVars)
	}

	darwinCfg := &config.TargetConfig{
		OS:    "darwin",
		Archs: []string{"x86_64"},
		LangOpts: map[string]any{
			"deployment_target": "12.0",
			"sdk_root":          "/path/to/sdk",
		},
	}
	if err := e.setupMacOSVars(darwinCfg); err != nil {
		t.Errorf("expected no error for darwin config, got: %v", err)
	}
	if os.Getenv("MACOSX_DEPLOYMENT_TARGET") != "12.0" {
		t.Error("expected MACOSX_DEPLOYMENT_TARGET to be set")
	}
	os.Unsetenv("MACOSX_DEPLOYMENT_TARGET")
	os.Unsetenv("SDKROOT")
}

func TestSetupDefaultMacOSDeployment(t *testing.T) {
	e := &RustEngine{}

	os.Unsetenv("MACOSX_DEPLOYMENT_TARGET")
	if err := e.setupDefaultMacOSDeployment(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if os.Getenv("MACOSX_DEPLOYMENT_TARGET") != "11.0" {
		t.Error("expected MACOSX_DEPLOYMENT_TARGET to be set to 11.0")
	}

	os.Setenv("MACOSX_DEPLOYMENT_TARGET", "12.0")
	if err := e.setupDefaultMacOSDeployment(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if os.Getenv("MACOSX_DEPLOYMENT_TARGET") != "12.0" {
		t.Error("expected MACOSX_DEPLOYMENT_TARGET to remain 12.0")
	}
	os.Unsetenv("MACOSX_DEPLOYMENT_TARGET")
}

func TestRunHook(t *testing.T) {
	e := &RustEngine{}

	if err := e.runHook("echo hello"); err != nil {
		t.Errorf("runHook with echo failed: %v", err)
	}

	if err := e.runHook("exit 1"); err == nil {
		t.Error("expected error from failing hook")
	}

	if err := e.runHook(""); err != nil {
		t.Errorf("expected no error for empty hook, got: %v", err)
	}
}
