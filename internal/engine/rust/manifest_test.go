package rust

import (
	"os"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	e := &RustEngine{}

	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Test valid manifest
	content := `
[package]
name = "test-pkg"
version = "1.2.3"
description = "A test"

[lib]
name = "test_lib"

[[bin]]
name = "test-bin"
`
	if err := os.WriteFile("Cargo.toml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := e.loadManifest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Package.Name != "test-pkg" || m.Package.Version != "1.2.3" {
		t.Errorf("incorrect package data: %+v", m.Package)
	}
	if m.Lib.Name != "test_lib" {
		t.Errorf("expected lib name test_lib, got %s", m.Lib.Name)
	}
	if len(m.Bin) != 1 || m.Bin[0].Name != "test-bin" {
		t.Errorf("incorrect bin data: %+v", m.Bin)
	}

	// Test invalid TOML
	if err := os.WriteFile("Cargo.toml", []byte(`invalid = [`), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = e.loadManifest()
	if err == nil {
		t.Error("expected error for invalid TOML")
	}

	// Test missing file
	os.Remove("Cargo.toml")
	_, err = e.loadManifest()
	if err == nil {
		t.Error("expected error for missing Cargo.toml")
	}
}
