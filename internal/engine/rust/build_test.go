package rust

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

// TestGetProfile tests profile extraction from target config.
func TestGetProfile(t *testing.T) {
	e := &RustEngine{}

	// Test default profile
	tCfg := config.TargetConfig{}
	if profile := e.getProfile(tCfg); profile != "release" {
		t.Errorf("expected 'release', got '%s'", profile)
	}

	// Test custom profile
	tCfg.LangOpts = map[string]any{"profile": "debug"}
	if profile := e.getProfile(tCfg); profile != "debug" {
		t.Errorf("expected 'debug', got '%s'", profile)
	}
}

// TestGetBinaryExt tests binary extension logic.
func TestGetBinaryExt(t *testing.T) {
	e := &RustEngine{}

	// Test bin on linux
	art := &config.ArtifactConfig{Type: "bin"}
	ext := e.getBinaryExt(art, "linux", "")
	if ext != "" {
		t.Errorf("expected empty string for bin on linux, got '%s'", ext)
	}

	// Test lib on linux
	art = &config.ArtifactConfig{Type: "lib", LibraryTypes: []string{"cdylib"}}
	ext = e.getBinaryExt(art, "linux", "")
	if ext != "so" {
		t.Errorf("expected 'so' for lib on linux, got '%s'", ext)
	}
}

// TestResolveBinaryInfo tests binary name resolution.
func TestResolveBinaryInfo(t *testing.T) {
	e := &RustEngine{}

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create a manifest
	cargoContent := `
[package]
name = "test-pkg"
version = "1.0.0"
`
	if err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := e.loadManifest()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Naming: config.NamingConfig{
			Binary:  "{artifact}-{os}-{arch}{abi}",
			Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
		},
		OutputDir: "dist",
	}

	art := &config.ArtifactConfig{Type: "bin"}
	opts := engine.BuildOptions{
		ArtifactName: "test-app",
		OS:           "linux",
		Arch:         "x86_64",
	}

	binaryName, binaryPath := e.resolveBinaryInfo(cfg, art, opts, m, "x86_64-unknown-linux-gnu", "release")
	if binaryName == "" {
		t.Error("expected non-empty binary name")
	}
	if binaryPath == "" {
		t.Error("expected non-empty binary path")
	}
}
