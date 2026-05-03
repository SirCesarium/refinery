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
