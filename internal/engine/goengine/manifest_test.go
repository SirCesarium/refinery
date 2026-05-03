package goengine

import (
	"os"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	goModContent := `module github.com/test/my-go-app

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	manifest, err := e.loadManifest()
	if err != nil {
		t.Fatalf("loadManifest failed: %v", err)
	}

	if manifest.Module.Path != "github.com/test/my-go-app" {
		t.Errorf("expected module path 'github.com/test/my-go-app', got '%s'", manifest.Module.Path)
	}
	if manifest.Go != "1.21" {
		t.Errorf("expected go version '1.21', got '%s'", manifest.Go)
	}
}

func TestLoadManifestMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	e := &GoEngine{}
	if _, err := e.loadManifest(); err == nil {
		t.Error("expected error for missing go.mod")
	}
}

func TestLoadManifestInvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	if err := os.WriteFile("go.mod", []byte("invalid content"), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	manifest, err := e.loadManifest()
	if err != nil {
		t.Fatalf("loadManifest failed: %v", err)
	}

	if manifest.Module.Path != "" {
		t.Errorf("expected empty module path for invalid content, got '%s'", manifest.Module.Path)
	}
}

func TestLoadManifestWithToolchain(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	goModContent := `module test-app

go 1.21

toolchain go1.21.0
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	manifest, err := e.loadManifest()
	if err != nil {
		t.Fatalf("loadManifest failed: %v", err)
	}

	if manifest.Toolchain != "go1.21.0" {
		t.Errorf("expected toolchain 'go1.21.0', got '%s'", manifest.Toolchain)
	}
}
