package cli

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine/rust"
)

func TestValidateRustEngine(t *testing.T) {
	eng := &rust.RustEngine{}
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

	cargoContent := `[package]
name = "test-project"
version = "0.1.0"
edition = "2024"
`
	if err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	if err := validateRustEngine(eng, cfg); err != nil {
		t.Errorf("validateRustEngine failed: %v", err)
	}
}

func TestMergePackages(t *testing.T) {
	global := []string{"deb", "rpm"}
	target := []string{"tar.gz", "deb"}

	result := mergePackages(global, target)

	if len(result) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(result))
	}

	seen := make(map[string]bool)
	for _, p := range result {
		if seen[p] {
			t.Errorf("duplicate package found: %s", p)
		}
		seen[p] = true
	}
}

func TestFindTargetPackages(t *testing.T) {
	art := &config.ArtifactConfig{
		Packages: []string{"deb", "rpm"},
		Targets: map[string]config.TargetConfig{
			"linux-x64": {
				OS:    "linux",
				Archs: []string{"x86_64"},
			},
		},
	}

	result := findTargetPackages(art, "linux", "x86_64", "")
	if len(result) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(result))
	}

	result = findTargetPackages(art, "windows", "x86_64", "")
	if result != nil {
		t.Errorf("expected nil for non-matching target, got %v", result)
	}
}

func TestMergePackagesEdgeCases(t *testing.T) {
	result := mergePackages(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}

	global := []string{"deb", "deb", "rpm"}
	target := []string{"deb", "tar.gz"}
	result = mergePackages(global, target)
	if len(result) != 3 {
		t.Errorf("expected 3 unique packages, got %d", len(result))
	}

	result = mergePackages(nil, []string{"deb"})
	if len(result) != 1 {
		t.Errorf("expected 1 package, got %v", result)
	}
}

func TestFindTargetPackagesEdgeCases(t *testing.T) {
	art := &config.ArtifactConfig{
		Packages: []string{"deb"},
	}
	result := findTargetPackages(art, "linux", "x86_64", "")
	if result != nil {
		t.Errorf("expected nil for no targets, got %v", result)
	}

	art = &config.ArtifactConfig{
		Packages: []string{"deb"},
		Targets: map[string]config.TargetConfig{
			"linux-x64": {
				OS:    "linux",
				Archs: []string{"x86_64"},
			},
		},
	}
	result = findTargetPackages(art, "linux", "x86_64", "")
	if len(result) != 1 {
		t.Errorf("expected 1 package from target, got %v", result)
	}
}

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected 'b' to be found in slice")
	}
	if contains([]string{"a", "b", "c"}, "d") {
		t.Error("expected 'd' to not be found in slice")
	}
	if contains(nil, "a") {
		t.Error("expected nil slice to return false")
	}
}

func TestFindTargetPackagesWithABI(t *testing.T) {
	art := &config.ArtifactConfig{
		Packages: []string{"deb"},
		Targets: map[string]config.TargetConfig{
			"linux-x64": {
				OS:    "linux",
				Archs: []string{"x86_64"},
				ABIs:  []string{"gnu", "musl"},
			},
		},
	}

	result := findTargetPackages(art, "linux", "x86_64", "gnu")
	if len(result) != 1 {
		t.Errorf("expected 1 package for linux-x86_64-gnu, got %v", result)
	}

	result = findTargetPackages(art, "linux", "x86_64", "musl")
	if len(result) != 1 {
		t.Errorf("expected 1 package for linux-x86_64-musl, got %v", result)
	}

	result = findTargetPackages(art, "linux", "x86_64", "unknown")
	if result != nil {
		t.Errorf("expected nil for unknown ABI, got %v", result)
	}
}

func TestFindTargetPackagesNoABI(t *testing.T) {
	art := &config.ArtifactConfig{
		Packages: []string{"deb"},
		Targets: map[string]config.TargetConfig{
			"linux-x64": {
				OS:    "linux",
				Archs: []string{"x86_64"},
			},
		},
	}

	result := findTargetPackages(art, "linux", "x86_64", "")
	if len(result) != 1 {
		t.Errorf("expected 1 package when no ABI specified, got %v", result)
	}

	art.Targets["linux-x64"] = config.TargetConfig{
		OS:    "linux",
		Archs: []string{"x86_64"},
		ABIs:  []string{"gnu"},
	}
	result = findTargetPackages(art, "linux", "x86_64", "")
	if len(result) != 1 {
		t.Errorf("expected 1 package when ABI not specified but target has ABIs, got %v", result)
	}
}
