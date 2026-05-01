package engine

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

func TestRustEngineValidate(t *testing.T) {
	cargoContent := `
[package]
name = "my-rust-app"
version = "0.1.0"

[[bin]]
name = "my-bin"
path = "src/main.rs"
`
	err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("Cargo.toml")

	e := &RustEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-rust-app", Lang: "rust"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-bin": {Type: "bin"},
		},
	}

	if err := e.Validate(cfg); err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}

	cfg.Artifacts["non-existent"] = &config.ArtifactConfig{Type: "bin"}
	if err := e.Validate(cfg); err == nil {
		t.Error("expected validation to fail for non-existent artifact")
	}
}
