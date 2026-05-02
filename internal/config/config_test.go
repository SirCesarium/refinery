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

func TestDefaultConfig(t *testing.T) {
	cfg := Default("my-project")
	if cfg.Project.Name != "my-project" {
		t.Errorf("expected my-project, got %s", cfg.Project.Name)
	}
	if cfg.OutputDir != "dist" {
		t.Errorf("expected dist, got %s", cfg.OutputDir)
	}
}

func TestConfigWrite(t *testing.T) {
	cfg := Default("write-test")
	err := cfg.Write("test.toml")
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	defer os.Remove("test.toml")

	if _, err := os.Stat("test.toml"); os.IsNotExist(err) {
		t.Error("test.toml was not created")
	}
}

func TestNamingResolution(t *testing.T) {
	n := NamingConfig{
		Binary:  "{artifact}-{os}-{arch}-{version}",
		Package: "{artifact}-{version}.{ext}",
	}
	res := n.Resolve(n.Binary, "myart", "linux", "x64", "1.0.0", "gnu", "so")
	expected := "myart-linux-x64-1.0.0.so"
	if res != expected {
		t.Errorf("binary resolve failed: got %s, want %s", res, expected)
	}

	res = n.Resolve(n.Binary, "myart", "linux", "x64", "1.0.0", "gnu", "")
	expected = "myart-linux-x64-1.0.0"
	if res != expected {
		t.Errorf("binary resolve without ext failed: got %s, want %s", res, expected)
	}

	res = n.Resolve(n.Package, "myart", "linux", "x64", "1.0.0", "gnu", "tar.gz")
	expected = "myart-1.0.0.tar.gz"
	if res != expected {
		t.Errorf("package resolve failed: got %s, want %s", res, expected)
	}
}
