package tests

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/app"
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
)

func TestFullWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "refinery-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	cfg := config.Default("integration-test")
	cfg.Project.Lang = "rust"
	cfg.Artifacts["test-bin"] = &config.ArtifactConfig{
		Type: "bin",
		Targets: map[string]config.TargetConfig{
			"linux-x64": {OS: "linux", Archs: []string{"x86_64"}, ABIs: []string{"gnu"}},
		},
	}
	
	err = cfg.Write("refinery.toml")
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cargoContent := `
[package]
name = "integration-test"
version = "0.1.0"

[[bin]]
name = "test-bin"
path = "src/main.rs"
`
	err = os.WriteFile("Cargo.toml", []byte(cargoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	registry := app.NewDefaultEngineRegistry()
	eng := registry.Get("rust")
	if eng == nil {
		t.Fatal("rust engine not found in registry")
	}

	if err := eng.Validate(cfg); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	pRegistry, err := app.NewDefaultProviderRegistry()
	if err != nil {
		t.Fatal(err)
	}
	provider := pRegistry.Get("github")
	if provider == nil {
		t.Fatal("github provider not found")
	}

	gen := pipeline.NewGenerator(provider, eng)
	data, err := gen.Generate(cfg)
	if err != nil {
		t.Fatalf("CI generation failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("generated CI data is empty")
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	t.Log("Integration test completed successfully (pre-execution steps)")
}

func TestEdgeCases(t *testing.T) {
	_, err := config.Load("non-existent.toml")
	if err == nil {
		t.Error("expected error for missing config")
	}

	registry := app.NewDefaultEngineRegistry()
	eng := registry.Get("rust")
	cfg := config.Default("missing-cargo")
	err = eng.Validate(cfg)
	if err == nil {
		t.Error("expected error for missing Cargo.toml in validation")
	}

	tmpDir, _ := os.MkdirTemp("", "mismatch-*")
	defer os.RemoveAll(tmpDir)
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	os.WriteFile("Cargo.toml", []byte("[package]\nname=\"real-name\""), 0644)
	cfg.Artifacts["fake-name"] = &config.ArtifactConfig{Type: "bin"}
	err = eng.Validate(cfg)
	if err == nil {
		t.Error("expected error for artifact name mismatch")
	}
}
