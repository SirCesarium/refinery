package github

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
)

func TestNewProviderFunction(t *testing.T) {
	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Filename() != ".github/workflows/build.yml" {
		t.Errorf("expected '.github/workflows/build.yml', got '%s'", p.Filename())
	}

	p, err = NewProvider("")
	if err == nil {
		t.Error("expected error for empty filename")
	}
	if p != nil {
		t.Error("expected nil provider for empty filename")
	}
}

func TestGenerateFunction(t *testing.T) {
	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust", "cargo-deb"},
	}

	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
		Artifacts: map[string]*config.ArtifactConfig{
			"test": {
				Type: "binary",
				Targets: map[string]config.TargetConfig{
					"linux-x64": {
						OS:    "linux",
						Archs: []string{"x86_64"},
					},
				},
			},
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Generate panicked: %v", r)
		}
	}()

	if _, err := p.Generate(cfg, mockEng); err != nil {
		t.Errorf("Generate failed: %v", err)
	}
}

type mockBuildEngineForGithub struct {
	requirements []string
}

func (m *mockBuildEngineForGithub) ID() string                        { return "mock" }
func (m *mockBuildEngineForGithub) Prepare(cfg *config.Config) error  { return nil }
func (m *mockBuildEngineForGithub) Validate(cfg *config.Config) error { return nil }
func (m *mockBuildEngineForGithub) Build(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions) error {
	return nil
}
func (m *mockBuildEngineForGithub) GetCIRequirements(cfg *config.Config) []string {
	return m.requirements
}
func (m *mockBuildEngineForGithub) Package(cfg *config.Config, art *config.ArtifactConfig, opts engine.BuildOptions, format string) error {
	return nil
}
func (m *mockBuildEngineForGithub) GetSupportedArchs(os string) []string {
	return []string{"x86_64", "i686", "aarch64"}
}

func TestGetBuildStepsFunction(t *testing.T) {
	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust", "cargo-deb"},
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	steps := p.getBuildSteps(mockEng, cfg)
	if len(steps) == 0 {
		t.Error("expected non-empty build steps")
	}
}

func TestAddCIRequirementStepsFunction(t *testing.T) {
	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust", "cargo-deb"},
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	steps := []Step{
		{Name: "Checkout", Uses: "actions/checkout@v4"},
	}

	newSteps := p.addCIRequirementSteps(steps, mockEng, cfg)
	if len(newSteps) <= len(steps) {
		t.Error("expected more steps after adding CI requirements")
	}
}

func TestGetBuildArtifactStepsFunction(t *testing.T) {
	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust", "cargo-deb"},
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
		Artifacts: map[string]*config.ArtifactConfig{
			"test": {
				Type: "binary",
				Targets: map[string]config.TargetConfig{
					"linux-x64": {
						OS:    "linux",
						Archs: []string{"x86_64"},
					},
				},
			},
		},
	}

	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	steps := p.getBuildArtifactStep(mockEng, cfg)
	if len(steps) == 0 {
		t.Error("expected non-empty steps")
	}
}
