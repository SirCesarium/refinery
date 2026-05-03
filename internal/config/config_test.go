package config

import (
	"path/filepath"
	"testing"
)

// TestDefault verifies the default configuration values.
func TestDefault(t *testing.T) {
	cfg := Default("test-project")
	if cfg.Project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", cfg.Project.Name)
	}
	if cfg.Project.Lang != "rust" {
		t.Errorf("Expected lang 'rust', got '%s'", cfg.Project.Lang)
	}
	if cfg.OutputDir != "dist" {
		t.Errorf("Expected output dir 'dist', got '%s'", cfg.OutputDir)
	}
}

// TestValidate checks config validation with various cases.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				RefineryVersion: "latest",
				Project:         Project{Name: "test", Lang: "rust"},
				OutputDir:       "dist",
				Naming: NamingConfig{
					Binary:  "{artifact}-{os}-{arch}{abi}",
					Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
				},
				Artifacts: map[string]*ArtifactConfig{
					"test_art": {
						Type:   "bin",
						Source: "src/main.rs",
						Targets: map[string]TargetConfig{
							"linux": {OS: "linux", Archs: []string{"amd64"}},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			cfg: &Config{
				RefineryVersion: "latest",
				Project:         Project{Lang: "rust"},
				OutputDir:       "dist",
			},
			wantErr: true,
		},
		{
			name: "invalid artifact type",
			cfg: &Config{
				RefineryVersion: "latest",
				Project:         Project{Name: "test", Lang: "rust"},
				OutputDir:       "dist",
				Artifacts: map[string]*ArtifactConfig{
					"test_art": {
						Type:   "invalid",
						Source: "src/main.rs",
						Targets: map[string]TargetConfig{
							"linux": {OS: "linux", Archs: []string{"amd64"}},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWriteAndLoad ensures config can be written to TOML and read back.
func TestWriteAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	tomlPath := filepath.Join(tmpDir, "refinery.toml")

	cfg := Default("test-project")
	cfg.Artifacts["test"] = &ArtifactConfig{
		Type:   "bin",
		Source: "src/main.rs",
		Targets: map[string]TargetConfig{
			"linux": {OS: "linux", Archs: []string{"amd64"}},
		},
	}

	if err := cfg.Write(tomlPath); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	loaded, err := Load(tomlPath)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	if loaded.Project.Name != cfg.Project.Name {
		t.Errorf("Loaded name mismatch: got %s, want %s", loaded.Project.Name, cfg.Project.Name)
	}
}

// TestRemoveRedundantFields checks that non-essential fields are stripped.
func TestRemoveRedundantFields(t *testing.T) {
	cfg := Default("test")
	cfg.Project = Project{Name: "test", Lang: "rust"}
	cfg.RemoveRedundantFields()

	if cfg.Project.Name != "test" || cfg.Project.Lang != "rust" {
		t.Error("RemoveRedundantFields changed essential fields")
	}
}

// TestNamingResolution validates the template substitution logic.
func TestNamingResolution(t *testing.T) {
	n := NamingConfig{
		Binary:  "{artifact}-{os}-{arch}{abi}",
		Package: "{artifact}-{version}-{os}-{arch}{abi}.{ext}",
	}

	binary := n.Resolve(n.Binary, "myapp", "linux", "amd64", "1.0.0", "musl", "exe")
	expected := "myapp-linux-amd64-musl.exe"
	if binary != expected {
		t.Errorf("Binary resolution failed: got %s, want %s", binary, expected)
	}

	pkg := n.Resolve(n.Package, "myapp", "linux", "amd64", "1.0.0", "musl", "deb")
	expectedPkg := "myapp-1.0.0-linux-amd64-musl.deb"
	if pkg != expectedPkg {
		t.Errorf("Package resolution failed: got %s, want %s", pkg, expectedPkg)
	}
}
