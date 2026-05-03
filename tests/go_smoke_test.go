package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SirCesarium/refinery/internal/app"
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

// TestGoInitSmoke tests that init creates correct config for Go projects.
func TestGoInitSmoke(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create a go.mod to simulate a Go project
	goMod := `module github.com/example/test-go-init

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a simple main.go
	mainGo := `package main

func main() {}
`
	if err := os.WriteFile("main.go", []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	// Create default config (simulating what init does with detection)
	cfg := config.Default("test-go-init")
	cfg.Project.Lang = "go"

	// Add artifact for the Go binary
	cfg.Artifacts["app"] = &config.ArtifactConfig{
		Type:   "bin",
		Source: ".",
		Targets: map[string]config.TargetConfig{
			"linux": {OS: "linux", Archs: []string{"amd64"}},
		},
	}

	if err := cfg.Write("refinery.toml"); err != nil {
		t.Fatalf("failed to write refinery.toml: %v", err)
	}

	// Verify the config was written correctly
	loaded, err := config.Load("refinery.toml")
	if err != nil {
		t.Fatalf("failed to load written config: %v", err)
	}

	if loaded.Project.Lang != "go" {
		t.Errorf("expected lang to be 'go', got '%s'", loaded.Project.Lang)
	}

	if len(loaded.Artifacts) == 0 {
		t.Error("expected artifacts to be defined")
	}
}

// TestGoBuildSmoke tests building a Go project.
func TestGoBuildSmoke(t *testing.T) {
	if _, err := os.Stat("/usr/local/go/bin/go"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/bin/go"); os.IsNotExist(err) {
			t.Skip("go not installed")
		}
	}

	// Get project root (where the tests/ directory is)
	origWd, _ := os.Getwd()
	// Find project root by looking for go.mod or refinery.toml
	projectRoot := origWd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		if _, err := os.Stat(filepath.Join(projectRoot, "refinery.toml")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			break
		}
		projectRoot = parent
	}

	// Use the pre-configured Go project in tests/smoke/go-project
	goProjectPath := filepath.Join(projectRoot, "tests", "smoke", "go-project")
	if _, err := os.Stat(goProjectPath); os.IsNotExist(err) {
		t.Fatalf("go-project smoke test directory not found at %s", goProjectPath)
	}

	os.Chdir(goProjectPath)
	defer os.Chdir(origWd)

	// Clean previous builds
	os.RemoveAll("dist")
	os.MkdirAll("dist", 0755)

	// Load config
	cfg, err := config.Load("refinery.toml")
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Get Go engine
	registry := app.NewDefaultEngineRegistry()
	eng := registry.Get("go")
	if eng == nil {
		t.Fatal("go engine not found in registry")
	}

	// Validate config
	if err := eng.Validate(cfg); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Build for linux/amd64
	art := cfg.Artifacts["app"]
	if art == nil {
		t.Fatal("artifact 'app' not found in config")
	}

	opts := engine.BuildOptions{
		ArtifactName: "app",
		OS:           "linux",
		Arch:         "amd64",
	}

	if err := eng.Build(cfg, art, opts); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Verify binary was created
	expectedBinary := filepath.Join(cfg.OutputDir, "app-linux-amd64")
	if _, err := os.Stat(expectedBinary); os.IsNotExist(err) {
		t.Errorf("expected binary to be created at %s", expectedBinary)
	}

	t.Log("Go build smoke test passed")
}

// TestGoFullWorkflowSmoke tests the full workflow with Go.
func TestGoFullWorkflowSmoke(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create Go project structure
	goMod := `module github.com/example/go-full-workflow

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello from Go!")
}
`
	if err := os.WriteFile("main.go", []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	// Create refinery config
	cfg := config.Default("go-full-workflow")
	cfg.Project.Lang = "go"
	cfg.OutputDir = "dist"
	cfg.Artifacts["app"] = &config.ArtifactConfig{
		Type:     "bin",
		Source:   ".",
		Packages: []string{"tar.gz"},
		Targets: map[string]config.TargetConfig{
			"linux": {OS: "linux", Archs: []string{"amd64"}},
		},
	}

	if err := cfg.Write("refinery.toml"); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Get engine and validate
	registry := app.NewDefaultEngineRegistry()
	eng := registry.Get("go")
	if eng == nil {
		t.Fatal("go engine not found")
	}

	if err := eng.Validate(cfg); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	// Build
	art := cfg.Artifacts["app"]
	opts := engine.BuildOptions{
		ArtifactName: "app",
		OS:           "linux",
		Arch:         "amd64",
	}

	if err := eng.Build(cfg, art, opts); err != nil {
		t.Fatalf("build failed: %v", err)
	}

	// Package
	if err := eng.Package(cfg, art, opts, "tar.gz"); err != nil {
		t.Fatalf("package failed: %v", err)
	}

	// Verify outputs
	expectedBinary := filepath.Join(cfg.OutputDir, "app-linux-amd64")
	if _, err := os.Stat(expectedBinary); os.IsNotExist(err) {
		t.Errorf("expected binary at %s", expectedBinary)
	}

	expectedPackage := filepath.Join(cfg.OutputDir, "app-0.0.0-linux-amd64.tar.gz")
	if _, err := os.Stat(expectedPackage); os.IsNotExist(err) {
		t.Errorf("expected package at %s", expectedPackage)
	}

	t.Log("Go full workflow smoke test passed")
}

// TestGoEngineRegistrationSmoke tests that Go engine is properly registered.
func TestGoEngineRegistrationSmoke(t *testing.T) {
	registry := app.NewDefaultEngineRegistry()

	// Check Go engine is registered
	eng := registry.Get("go")
	if eng == nil {
		t.Fatal("go engine not found in registry")
	}

	if eng.ID() != "go" {
		t.Errorf("expected engine ID to be 'go', got '%s'", eng.ID())
	}

	// Check Rust engine is still registered
	rustEng := registry.Get("rust")
	if rustEng == nil {
		t.Fatal("rust engine not found in registry")
	}
}
