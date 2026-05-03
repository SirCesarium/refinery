package config

import (
	"testing"
)

// TestHooksResolveAll tests hook resolution with version and abi.
func TestHooksResolveAll(t *testing.T) {
	h := Hooks{
		PreBuild:  []string{"echo {version}", "echo {abi}"},
		PostBuild: []string{"echo {artifact}"},
	}

	resolved := h.ResolveAll("my-art", "linux", "x86_64", "1.0.0", "gnu", "/bin/my-art")

	if len(resolved.PreBuild) != 2 {
		t.Fatalf("expected 2 pre-build hooks, got %d", len(resolved.PreBuild))
	}
	if resolved.PreBuild[0] != "echo 1.0.0" {
		t.Errorf("expected 'echo 1.0.0', got '%s'", resolved.PreBuild[0])
	}
	if resolved.PreBuild[1] != "echo gnu" {
		t.Errorf("expected 'echo gnu', got '%s'", resolved.PreBuild[1])
	}
	if len(resolved.PostBuild) != 1 {
		t.Fatalf("expected 1 post-build hook, got %d", len(resolved.PostBuild))
	}
}

// TestNamingResolveEdgeCases tests edge cases in naming resolution.
func TestNamingResolveEdgeCases(t *testing.T) {
	n := NamingConfig{
		Binary:  "{artifact}-{os}-{arch}{abi}",
		Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
	}

	// Test with empty abi
	result := n.Resolve(n.Binary, "myapp", "linux", "amd64", "1.0.0", "", "exe")
	expected := "myapp-linux-amd64"
	if result != expected {
		t.Errorf("Binary resolution failed: got '%s', want '%s'", result, expected)
	}

	// Test with abi
	result = n.Resolve(n.Binary, "myapp", "linux", "amd64", "1.0.0", "musl", "exe")
	expected = "myapp-linux-amd64-musl"
	if result != expected {
		t.Errorf("Binary resolution with abi failed: got '%s', want '%s'", result, expected)
	}
}
