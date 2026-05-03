package goengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

func TestPkgTarGz(t *testing.T) {
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

	if err := os.MkdirAll("dist", 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("dist/test-app-linux-amd64", []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		OutputDir: "dist",
		Naming: config.NamingConfig{
			Binary:  "{artifact}-{os}-{arch}{abi}",
			Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
		},
		Artifacts: map[string]*config.ArtifactConfig{
			"test-app": {
				Type: "bin",
			},
		},
	}

	if err := e.pkg(cfg, cfg.Artifacts["test-app"], "test-app", "linux", "amd64", "", "tar.gz"); err != nil {
		t.Errorf("pkg tar.gz failed: %v", err)
	}

	expectedFile := filepath.Join("dist", "test-app-0.0.0-linux-amd64.tar.gz")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected package to be created at %s", expectedFile)
	}
}

func TestPkgZip(t *testing.T) {
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

	if err := os.MkdirAll("dist", 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("dist/test-app-windows-amd64.exe", []byte("binary"), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		OutputDir: "dist",
		Naming: config.NamingConfig{
			Binary:  "{artifact}-{os}-{arch}{abi}",
			Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
		},
		Artifacts: map[string]*config.ArtifactConfig{
			"test-app": {
				Type: "bin",
			},
		},
	}

	if err := e.pkg(cfg, cfg.Artifacts["test-app"], "test-app", "windows", "amd64", "", "zip"); err != nil {
		t.Errorf("pkg zip failed: %v", err)
	}

	expectedFile := filepath.Join("dist", "test-app-0.0.0-windows-amd64.zip")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected package to be created at %s", expectedFile)
	}
}

func TestPkgUnsupportedFormat(t *testing.T) {
	e := &GoEngine{}
	cfg := &config.Config{
		OutputDir: "dist",
		Artifacts: map[string]*config.ArtifactConfig{
			"test": {Type: "bin"},
		},
	}

	if err := e.pkg(cfg, cfg.Artifacts["test"], "test", "linux", "amd64", "", "deb"); err == nil {
		t.Error("expected error for unsupported format 'deb'")
	}
}
