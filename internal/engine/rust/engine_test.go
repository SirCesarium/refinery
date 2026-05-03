package rust

import (
	"os"
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

// TestRustEngineValidate checks validation logic against a Cargo.toml.
func TestRustEngineValidate(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	cargoContent := `
[package]
name = "my-rust-app"
version = "0.1.0"

[[bin]]
name = "my-bin"
path = "src/main.rs"
`
	err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	e := &RustEngine{}
	cfg := &config.Config{
		Project: config.Project{Name: "my-rust-app", Lang: "rust"},
		Artifacts: map[string]*config.ArtifactConfig{
			"my-bin": {Type: "bin"},
		},
	}

	if err := e.Validate(cfg); err != nil {
		t.Errorf("expected validation to pass, got: %v", err)
	}

	cfg.Artifacts["non-existent"] = &config.ArtifactConfig{Type: "bin"}
	if err := e.Validate(cfg); err == nil {
		t.Error("expected validation to fail for non-existent artifact")
	}
}

// TestIsLibDefined checks library artifact detection in Cargo.toml.
func TestIsLibDefined(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create Cargo.toml with lib
	cargoContent := `
[package]
name = "my-package"
version = "0.1.0"

[lib]
name = "my-lib"
`
	if err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &RustEngine{}
	m, err := e.loadManifest()
	if err != nil {
		t.Fatal(err)
	}

	// Should match lib name
	if !e.isLibDefined("my-lib", m) {
		t.Error("expected my-lib to be defined")
	}
	// Should match package name with underscores
	if !e.isLibDefined("my_package", m) {
		t.Error("expected my_package to match package name")
	}
	// Should not match wrong name
	if e.isLibDefined("wrong-lib", m) {
		t.Error("expected wrong-lib to not be defined")
	}
}

// TestIsBinDefined checks binary artifact detection in Cargo.toml.
func TestIsBinDefined(t *testing.T) {
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create Cargo.toml with bins
	cargoContent := `
[package]
name = "my-package"
version = "0.1.0"

[[bin]]
name = "bin1"

[[bin]]
name = "bin2"
`
	if err := os.WriteFile("Cargo.toml", []byte(cargoContent), 0644); err != nil {
		t.Fatal(err)
	}

	e := &RustEngine{}
	m, err := e.loadManifest()
	if err != nil {
		t.Fatal(err)
	}

	// Should match binary name
	if !e.isBinDefined("bin1", m) {
		t.Error("expected bin1 to be defined")
	}
	// Should match package name
	if !e.isBinDefined("my-package", m) {
		t.Error("expected my-package to match")
	}
	// Should not match wrong name
	if e.isBinDefined("wrong-bin", m) {
		t.Error("expected wrong-bin to not be defined")
	}
}

// TestAddTargetRequirements checks CI requirements based on target config.
func TestAddTargetRequirements(t *testing.T) {
	e := &RustEngine{}

	// Test linux + aarch64
	reqs := []string{"rust"}
	art := &config.ArtifactConfig{
		Targets: map[string]config.TargetConfig{
			"linux-arm": {OS: "linux", Archs: []string{"aarch64"}},
		},
	}
	reqs = e.addTargetRequirements(reqs, art)

	found := false
	for _, r := range reqs {
		if r == "cross-linker:linux-aarch64" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected cross-linker:linux-aarch64 requirement")
	}
}

// TestAddPackageRequirements checks CI requirements based on package formats.
func TestAddPackageRequirements(t *testing.T) {
	e := &RustEngine{}

	// Test deb package
	reqs := []string{}
	art := &config.ArtifactConfig{
		Packages: []string{"deb"},
	}
	reqs = e.addPackageRequirements(reqs, art)

	found := false
	for _, r := range reqs {
		if r == "pkg:cargo-deb" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected pkg:cargo-deb requirement")
	}
}
