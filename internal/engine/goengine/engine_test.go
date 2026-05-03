package goengine

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

func TestGoEngineID(t *testing.T) {
	e := &GoEngine{}
	if e.ID() != "go" {
		t.Errorf("expected ID to be 'go', got %s", e.ID())
	}
}

func TestPrepare(t *testing.T) {
	e := &GoEngine{}
	if err := e.Prepare(nil); err != nil {
		t.Errorf("expected no error from Prepare, got: %v", err)
	}
}

func TestValidate(t *testing.T) {
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

	goModContent := `module my-go-app

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-go-app", Lang: "go"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-app": {
				Type:   "bin",
				Source: ".",
			},
		},
	}

	if err := e.Validate(cfg); err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}
}

func TestValidateInvalidTarget(t *testing.T) {
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

	goModContent := `module my-go-app

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-go-app", Lang: "go"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-app": {
				Type:   "bin",
				Source: ".",
				Targets: map[string]config.TargetConfig{
					"invalid": {OS: "invalid_os", Archs: []string{"amd64"}},
				},
			},
		},
	}

	if err := e.Validate(cfg); err == nil {
		t.Error("expected validation to fail for invalid OS")
	}
}

func TestValidateMissingSource(t *testing.T) {
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

	goModContent := `module my-go-app

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-go-app", Lang: "go"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-app": {
				Type: "bin",
			},
		},
	}

	if err := e.Validate(cfg); err == nil {
		t.Error("expected validation to fail for missing source")
	}
}

func TestValidateWithABI(t *testing.T) {
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

	goModContent := `module my-go-app

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-go-app", Lang: "go"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-app": {
				Type:   "bin",
				Source: ".",
				Targets: map[string]config.TargetConfig{
					"linux": {OS: "linux", Archs: []string{"amd64"}, ABIs: []string{"gnu"}},
				},
			},
		},
	}

	if err := e.Validate(cfg); err == nil {
		t.Error("expected validation to fail for Go with ABI specified")
	}
}

func TestGetCIRequirements(t *testing.T) {
	e := &GoEngine{}
	cfg := &config.Config{
		Artifacts: map[string]*config.ArtifactConfig{
			"test": {
				Type:     "bin",
				Source:   ".",
				Packages: []string{"tar.gz", "zip"},
			},
		},
	}

	reqs := e.GetCIRequirements(cfg)
	if len(reqs) == 0 {
		t.Error("expected CI requirements")
	}

	found := false
	for _, r := range reqs {
		if r == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'go' in CI requirements")
	}
}

func TestBuild(t *testing.T) {
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

	goModContent := `module test-go-build

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	if err := os.WriteFile("main.go", []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("dist", 0755); err != nil {
		t.Fatal(err)
	}

	e := &GoEngine{}
	cfg := &config.Config{
		Project:   config.Project{Name: "test-go-build", Lang: "go"},
		OutputDir: "dist",
		Artifacts: map[string]*config.ArtifactConfig{
			"test-go-build": {
				Type:   "bin",
				Source: ".",
				Targets: map[string]config.TargetConfig{
					"linux-x64": {OS: "linux", Archs: []string{"amd64"}},
				},
			},
		},
	}

	if err := e.Build(cfg, cfg.Artifacts["test-go-build"], engine.BuildOptions{
		ArtifactName: "test-go-build",
		OS:           "linux",
		Arch:         "amd64",
	}); err != nil {
		t.Errorf("Build returned error: %v", err)
	}
}

func TestSliceContains(t *testing.T) {
	e := &GoEngine{}
	if !e.sliceContains([]string{"a", "b", "c"}, "b") {
		t.Error("expected 'b' to be found")
	}
	if e.sliceContains([]string{"a", "b", "c"}, "d") {
		t.Error("expected 'd' to not be found")
	}
}

func TestUniqueFormats(t *testing.T) {
	e := &GoEngine{}
	formats := []string{"tar.gz", "zip", "tar.gz", "zip", "deb"}
	unique := e.uniqueFormats(formats)
	if len(unique) != 3 {
		t.Errorf("expected 3 unique formats, got %d", len(unique))
	}
}
