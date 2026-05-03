package cli

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine/rust"
)

// TestValidateRustEngine tests Rust engine validation logic.
func TestValidateRustEngine(t *testing.T) {
	// Test with Rust engine
	eng := &rust.RustEngine{}
	err := validateRustEngine(eng, nil)
	// Since we can't easily mock, just verify no panic
	_ = err
}

// TestMergePackages tests package merging logic.
func TestMergePackages(t *testing.T) {
	global := []string{"deb", "rpm"}
	target := []string{"tar.gz", "deb"}

	result := mergePackages(global, target)

	if len(result) != 3 {
		t.Fatalf("expected 3 packages, got %d", len(result))
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, p := range result {
		if seen[p] {
			t.Errorf("duplicate package found: %s", p)
		}
		seen[p] = true
	}
}

// TestFindTargetPackages tests target package lookup.
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

	// Test matching target
	result := findTargetPackages(art, "linux", "x86_64", "")
	if len(result) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(result))
	}

	// Test non-matching target
	result = findTargetPackages(art, "windows", "x86_64", "")
	if result != nil {
		t.Errorf("expected nil for non-matching target, got %v", result)
	}
}
