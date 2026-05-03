package github

import (
	"testing"

	"github.com/SirCesarium/refinery/internal/config"
)

// TestGetRunsOn tests OS to runner label mapping.
func TestGetRunsOn(t *testing.T) {
	p := &GithubProvider{}

	if label := p.getRunsOn("linux"); label != "ubuntu-latest" {
		t.Errorf("expected 'ubuntu-latest', got '%s'", label)
	}
	if label := p.getRunsOn("windows"); label != "windows-latest" {
		t.Errorf("expected 'windows-latest', got '%s'", label)
	}
	if label := p.getRunsOn("darwin"); label != "macos-latest" {
		t.Errorf("expected 'macos-latest', got '%s'", label)
	}
}

// TestSortedArtifactNames tests artifact name sorting.
func TestSortedArtifactNames(t *testing.T) {
	p := &GithubProvider{}
	cfg := &config.Config{
		Artifacts: map[string]*config.ArtifactConfig{
			"zebra": {},
			"apple": {},
			"mango": {},
		},
	}

	names := p.sortedArtifactNames(cfg)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	if names[0] != "apple" || names[1] != "mango" || names[2] != "zebra" {
		t.Errorf("expected sorted names [apple, mango, zebra], got %v", names)
	}
}

// TestAssembleJobs tests job assembly logic.
func TestAssembleJobs(t *testing.T) {
	p := &GithubProvider{}
	include := []map[string]any{
		{"artifact": "test", "os": "linux", "arch": "x86_64", "runs_on": "ubuntu-latest"},
	}
	steps := []Step{
		{Name: "Test Step", Run: "echo test"},
	}

	jobs := p.assembleJobs(include, steps)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
	if _, ok := jobs["build"]; !ok {
		t.Error("expected 'build' job")
	}
	if _, ok := jobs["release"]; !ok {
		t.Error("expected 'release' job")
	}
}

// TestBuildMatrix tests matrix generation.
func TestBuildMatrix(t *testing.T) {
	p := &GithubProvider{}
	cfg := &config.Config{
		Artifacts: map[string]*config.ArtifactConfig{
			"test-art": {
				Targets: map[string]config.TargetConfig{
					"linux-x64": {OS: "linux", Archs: []string{"x86_64"}},
				},
			},
		},
	}

	include := p.buildMatrix(cfg)
	if len(include) != 1 {
		t.Fatalf("expected 1 matrix entry, got %d", len(include))
	}
	if include[0]["os"] != "linux" {
		t.Errorf("expected os 'linux', got '%v'", include[0]["os"])
	}
}

// TestNewProvider tests provider initialization.
func TestNewProvider(t *testing.T) {
	p, err := NewProvider(".github/workflows/build.yml")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Name() != "github" {
		t.Errorf("expected 'github', got '%s'", p.Name())
	}
}

// TestProviderName tests Name method.
func TestProviderName(t *testing.T) {
	p, _ := NewProvider(".github/workflows/build.yml")
	if p.Name() != "github" {
		t.Errorf("expected 'github', got '%s'", p.Name())
	}
}

// TestProviderFilename tests Filename method.
func TestProviderFilename(t *testing.T) {
	filename := "custom"
	p, _ := NewProvider(filename)
	expected := ".github/workflows/" + filename + ".yml"
	if p.Filename() != expected {
		t.Errorf("expected '%s', got '%s'", expected, p.Filename())
	}
}

func TestProviderGenerate(t *testing.T) {
	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust"},
	}

	if _, err := p.Generate(cfg, mockEng); err != nil {
		t.Errorf("Generate failed: %v", err)
	}
}

func TestGetBuildSteps(t *testing.T) {
	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust"},
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	steps := p.getBuildSteps(mockEng, cfg)
	if len(steps) == 0 {
		t.Error("expected non-empty build steps")
	}
}

func TestAddCIRequirementSteps(t *testing.T) {
	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust", "cargo-deb"},
	}
	cfg := &config.Config{
		Project: config.Project{
			Name: "test-project",
			Lang: "rust",
		},
	}

	steps := []Step{
		{Name: "Checkout", Uses: "actions/checkout@v4"},
	}

	newSteps := p.addCIRequirementSteps(steps, mockEng, cfg)
	if len(newSteps) <= len(steps) {
		t.Error("expected more steps after adding CI requirements")
	}
}

func TestGetBuildArtifactSteps(t *testing.T) {
	p, err := NewProvider("build")
	if err != nil {
		t.Fatalf("NewProvider returned error: %v", err)
	}

	mockEng := &mockBuildEngineForGithub{
		requirements: []string{"rust"},
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

	steps := p.getBuildArtifactStep(mockEng, cfg)
	if len(steps) == 0 {
		t.Error("expected non-empty steps")
	}
}
