package config

import (
	"os"
	"testing"
)

func TestLoadInvalidConfig(t *testing.T) {
	content := `
refinery_version = "2"
output_dir = "dist"
[project]
name = "test"
lang = "rust"
[artifacts.invalid]
type = "invalid"
source = "src/main.rs"
[artifacts.invalid.targets.linux]
archs = ["x86_64"]
`
	err := os.WriteFile("invalid.toml", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("invalid.toml")

	_, err = Load("invalid.toml")
	if err == nil {
		t.Error("expected error for invalid artifact type")
	}
}

func TestLoadValidConfig(t *testing.T) {
	content := `
refinery_version = "2"
output_dir = "dist"

[project]
name = "test-project"
lang = "rust"

[naming]
binary = "{artifact}"
package = "{artifact}.{ext}"

[artifacts.test-bin]
type = "bin"
source = "src/main.rs"
[artifacts.test-bin.targets.linux]
archs = ["x86_64"]
`
	err := os.WriteFile("valid.toml", []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("valid.toml")

	cfg, err := Load("valid.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name test-project, got %s", cfg.Project.Name)
	}
}
