package goengine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

func TestRunHookFunction(t *testing.T) {
	e := &GoEngine{}

	if err := e.runHook("echo hello from go"); err != nil {
		t.Errorf("runHook with echo failed: %v", err)
	}

	if err := e.runHook("exit 1"); err == nil {
		t.Error("expected error from failing hook")
	}
}

func TestRunHookEmpty(t *testing.T) {
	e := &GoEngine{}

	if err := e.runHook(""); err != nil {
		t.Errorf("expected no error for empty hook, got: %v", err)
	}
}

func TestAddPackageRequirements(t *testing.T) {
	e := &GoEngine{}

	reqs := []string{"go"}
	art := &config.ArtifactConfig{
		Packages: []string{"deb", "tar.gz"},
	}
	reqs = e.addPackageRequirements(reqs, art)

	found := false
	for _, r := range reqs {
		if r == "pkg:go-bin-tools" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected pkg:go-bin-tools requirement")
	}
}

func TestBuildFunction(t *testing.T) {
	if _, err := os.Stat("/usr/local/go/bin/go"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/bin/go"); os.IsNotExist(err) {
			t.Skip("go not installed")
		}
	}

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

	goModContent := `module test-go-build-extra

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

func main() {
}
`
	if err := os.WriteFile("main.go", []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("output", 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project:   config.Project{Name: "test-go-build-extra", Lang: "go"},
		OutputDir: "output",
		Naming: config.NamingConfig{
			Binary:  "{artifact}-{os}-{arch}{abi}",
			Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
		},
		Artifacts: map[string]*config.ArtifactConfig{
			"test-go-build-extra": {
				Type:   "bin",
				Source: ".",
				Targets: map[string]config.TargetConfig{
					"linux": {OS: "linux", Archs: []string{"amd64"}},
				},
			},
		},
	}

	e := &GoEngine{}
	art := cfg.Artifacts["test-go-build-extra"]
	opts := engine.BuildOptions{
		ArtifactName: "test-go-build-extra",
		OS:           "linux",
		Arch:         "amd64",
	}

	if err := e.build(cfg, art, opts); err != nil {
		t.Errorf("build failed: %v", err)
	}

	expectedBinary := filepath.Join("output", "test-go-build-extra-linux-amd64")
	if _, err := os.Stat(expectedBinary); os.IsNotExist(err) {
		t.Errorf("expected binary to be created at %s", expectedBinary)
	}
}

func TestGetCIRequirementsFunction(t *testing.T) {
	e := &GoEngine{}
	cfg := &config.Config{
		Artifacts: map[string]*config.ArtifactConfig{
			"test": {
				Type:     "bin",
				Source:   ".",
				Packages: []string{"tar.gz", "zip", "deb"},
			},
		},
	}

	reqs := e.GetCIRequirements(cfg)
	if len(reqs) == 0 {
		t.Error("expected CI requirements")
	}

	hasGo := false
	for _, r := range reqs {
		if r == "go" {
			hasGo = true
			break
		}
	}
	if !hasGo {
		t.Error("expected 'go' in requirements")
	}
}
