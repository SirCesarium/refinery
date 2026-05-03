// Package github implements the CIProvider interface for GitHub Actions workflows.
package github

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
	"gopkg.in/yaml.v3"
)

// Workflow represents a GitHub Actions workflow YAML structure.
type Workflow struct {
	Name        string            `yaml:"name"`
	On          On                `yaml:"on"`
	Permissions map[string]string `yaml:"permissions,omitempty"`
	Jobs        map[string]Job    `yaml:"jobs"`
}

type On struct {
	Push    *Event         `yaml:"push,omitempty"`
	Release map[string]any `yaml:"release,omitempty"`
}

type Event struct {
	Tags []string `yaml:"tags,omitempty"`
}

// Job defines a single job in the workflow.
type Job struct {
	Name     string    `yaml:"name,omitempty"`
	RunsOn   string    `yaml:"runs-on"`
	Needs    []string  `yaml:"needs,omitempty"`
	If       string    `yaml:"if,omitempty"`
	Strategy *Strategy `yaml:"strategy,omitempty"`
	Steps    []Step    `yaml:"steps"`
}

type Strategy struct {
	FailFast bool           `yaml:"fail-fast,omitempty"`
	Matrix   map[string]any `yaml:"matrix,omitempty"`
}

// Step represents a single step within a job.
type Step struct {
	Name  string            `yaml:"name,omitempty"`
	If    string            `yaml:"if,omitempty"`
	Uses  string            `yaml:"uses,omitempty"`
	With  map[string]any    `yaml:"with,omitempty"`
	Run   string            `yaml:"run,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
	Shell string            `yaml:"shell,omitempty"`
}

type GithubProvider struct {
	filename string
}

// NewProvider creates a new GitHub Actions workflow provider.
func NewProvider(filename string) (*GithubProvider, error) {
	if filename == "" {
		return nil, fmt.Errorf("workflow filename cannot be empty")
	}
	return &GithubProvider{filename: filename}, nil
}

func (p *GithubProvider) Name() string {
	return "github"
}

func (p *GithubProvider) Filename() string {
	return filepath.Join(".github", "workflows", fmt.Sprintf("%s.yml", p.filename))
}

// Generate creates a GitHub Actions workflow YAML for the given config and engine.
func (p *GithubProvider) Generate(cfg *config.Config, eng engine.BuildEngine) ([]byte, error) {
	if err := eng.Validate(cfg); err != nil {
		return nil, err
	}

	include := p.buildMatrix(cfg)
	buildSteps := p.getBuildSteps(eng, cfg)
	jobs := p.assembleJobs(include, buildSteps)

	wf := Workflow{
		Name: "Refinery Build",
		On: On{
			Push: &Event{Tags: []string{"v*"}},
			Release: map[string]any{
				"types": []string{"created"},
			},
		},
		Permissions: map[string]string{
			"contents": "write",
			"packages": "write",
		},
		Jobs: jobs,
	}

	return yaml.Marshal(wf)
}

// buildMatrix creates the matrix include array from config artifacts and targets.
func (p *GithubProvider) buildMatrix(cfg *config.Config) []map[string]any {
	var include []map[string]any
	uniqueMatrix := make(map[string]bool)

	for _, aName := range p.sortedArtifactNames(cfg) {
		art := cfg.Artifacts[aName]
		for _, tCfg := range art.Targets {
			runsOn := p.getRunsOn(tCfg.OS)
			for _, arch := range tCfg.Archs {
				abis := tCfg.ABIs
				if len(abis) == 0 {
					abis = []string{""}
				}
				for _, abi := range abis {
					key := fmt.Sprintf("%s-%s-%s-%s", aName, tCfg.OS, arch, abi)
					if uniqueMatrix[key] {
						continue
					}
					uniqueMatrix[key] = true

					m := map[string]any{
						"artifact": aName,
						"os":       tCfg.OS,
						"arch":     arch,
						"runs_on":  runsOn,
					}
					if abi != "" {
						m["abi"] = abi
					}
					include = append(include, m)
				}
			}
		}
	}
	return include
}

// sortedArtifactNames returns artifact names in sorted order.
func (p *GithubProvider) sortedArtifactNames(cfg *config.Config) []string {
	var names []string
	for name := range cfg.Artifacts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getRunsOn maps OS to GitHub Actions runner labels.
func (p *GithubProvider) getRunsOn(osName string) string {
	switch osName {
	case "windows":
		return "windows-latest"
	case "darwin":
		return "macos-latest"
	default:
		return "ubuntu-latest"
	}
}

// getBuildSteps assembles the list of steps for the build job.
func (p *GithubProvider) getBuildSteps(eng engine.BuildEngine, cfg *config.Config) []Step {
	steps := []Step{
		{Name: "Checkout", Uses: ActionCheckout},
	}
	steps = p.addCIRequirementSteps(steps, eng, cfg)

	// Build refinery first if enabled (needed for pre_build steps that use it)
	if cfg.BuildRefinery != nil && cfg.BuildRefinery.Enabled {
		steps = append(steps, Step{
			Name: "Build Refinery from Source",
			Run:  "go build -o ./refinery-local ./cmd/refinery",
		})
	}

	steps = p.addPreBuildSteps(steps, cfg)
	steps = append(steps, p.getBuildArtifactStep(cfg)...)
	steps = p.addPostBuildSteps(steps, cfg)
	return steps
}

// addCIRequirementSteps adds steps based on engine requirements.
func (p *GithubProvider) addCIRequirementSteps(steps []Step, eng engine.BuildEngine, cfg *config.Config) []Step {
	for _, req := range eng.GetCIRequirements(cfg) {
		switch {
		case req == "go":
			steps = append(steps, Step{
				Name: "Setup Go",
				Uses: ActionSetupGo,
				With: map[string]any{"go-version": "1.26.2", "cache": true},
			})
		case req == "pkg:go-bin-tools":
			steps = append(steps, Step{
				Name: "Install Go Bin Tools",
				If:   "runner.os == 'Linux'",
				Run:  "go install github.com/mh-cbon/go-bin-deb@latest && go install github.com/mh-cbon/go-bin-rpm@latest",
			})
		case req == "rust":
			steps = append(steps, Step{
				Name: "Setup Rust",
				Uses: ActionRustToolchain,
				With: map[string]any{"cache": true},
			})
		case req == "cross-linker:linux-aarch64":
			steps = append(steps, Step{
				Name: "Install ARM Linker",
				If:   "runner.os == 'Linux'",
				Run:  "sudo apt-get update && sudo apt-get install -y gcc-aarch64-linux-gnu",
			})
		case req == "pkg:musl-tools":
			steps = append(steps, Step{
				Name: "Install Musl Tools",
				If:   "runner.os == 'Linux'",
				Run:  "sudo apt-get update && sudo apt-get install -y musl-tools",
			})
		case req == "pkg:cargo-deb":
			steps = append(steps, Step{
				Name: "Install cargo-deb",
				If:   "runner.os == 'Linux'",
				Run:  "cargo install cargo-deb",
			})
		case req == "pkg:cargo-generate-rpm":
			steps = append(steps, Step{
				Name: "Install cargo-generate-rpm",
				If:   "runner.os == 'Linux'",
				Run:  "cargo install cargo-generate-rpm",
			})
		case req == "pkg:cargo-wix":
			steps = append(steps, Step{
				Name: "Install cargo-wix",
				If:   "runner.os == 'Windows'",
				Run:  "cargo install cargo-wix",
			})
		}
	}
	return steps
}

// getBuildArtifactStep returns the artifact build step.
func (p *GithubProvider) getBuildArtifactStep(cfg *config.Config) []Step {
	if cfg.BuildRefinery != nil && cfg.BuildRefinery.Enabled {
		return []Step{
			{
				Name: "Build Artifact using Local Refinery",
				Run:  "./refinery-local build --artifact ${{ matrix.artifact }} --os ${{ matrix.os }} --arch ${{ matrix.arch }}${{ matrix.abi != '' && format(' --abi {0}', matrix.abi) || '' }}",
			},
		}
	}

	return []Step{
		{
			Name: "Build Artifact",
			Uses: ActionRefinery,
			With: map[string]any{
				"artifact": "${{ matrix.artifact }}",
				"os":       "${{ matrix.os }}",
				"arch":     "${{ matrix.arch }}",
				"abi":      "${{ matrix.abi }}",
				"version":  cfg.RefineryVersion,
			},
		},
	}
}

// addPreBuildSteps adds steps from [[pre_build]] config.
func (p *GithubProvider) addPreBuildSteps(steps []Step, cfg *config.Config) []Step {
	for _, step := range cfg.PreBuild {
		steps = append(steps, p.createGithubStep(step, "Pre-Build"))
	}
	return steps
}

// addPostBuildSteps adds steps from [[post_build]] config.
func (p *GithubProvider) addPostBuildSteps(steps []Step, cfg *config.Config) []Step {
	for _, step := range cfg.PostBuild {
		steps = append(steps, p.createGithubStep(step, "Post-Build"))
	}
	return steps
}

func (p *GithubProvider) createGithubStep(step config.BuildStep, prefix string) Step {
	// Resolve action name to full action path if needed
	action := step.Action
	if action != "" && !strings.Contains(action, "/") && !strings.HasSuffix(action, ".yml") {
		action = fmt.Sprintf("./.github/actions/%s", action)
	}

	// Use action name as ID if not provided
	id := step.ID
	if id == "" && step.Action != "" {
		// Extract name from action path
		parts := strings.Split(step.Action, "/")
		name := parts[len(parts)-1]
		name = strings.TrimSuffix(name, ".yml")
		id = name
	}

	ghStep := Step{
		Name: fmt.Sprintf("%s: %s", prefix, id),
	}
	if action != "" {
		ghStep.Uses = action
		ghStep.With = step.With
	} else if len(step.Command) > 0 {
		ghStep.Run = strings.Join(step.Command, "\n")
	}

	if len(step.OS) > 0 {
		var conditions []string
		for _, osName := range step.OS {
			switch strings.ToLower(osName) {
			case "linux":
				conditions = append(conditions, "runner.os == 'Linux'")
			case "windows":
				conditions = append(conditions, "runner.os == 'Windows'")
			case "darwin", "macos":
				conditions = append(conditions, "runner.os == 'macOS'")
			}
		}
		if len(conditions) > 0 {
			ghStep.If = strings.Join(conditions, " || ")
		}
	}
	return ghStep
}

// assembleJobs creates the jobs map for the workflow.
func (p *GithubProvider) assembleJobs(include []map[string]any, buildSteps []Step) map[string]Job {
	return map[string]Job{
		"build": {
			Name:   "Build ${{ matrix.artifact }} (${{ matrix.os }}-${{ matrix.arch }})",
			RunsOn: "${{ matrix.runs_on }}",
			Strategy: &Strategy{
				FailFast: true,
				Matrix:   map[string]any{"include": include},
			},
			Steps: buildSteps,
		},
		"release": {
			Name:   "Release Artifacts",
			Needs:  []string{"build"},
			RunsOn: "ubuntu-latest",
			If:     "startsWith(github.ref, 'refs/tags/')",
			Steps: []Step{
				{
					Name: "Download Artifacts",
					Uses: ActionDownloadArtifact,
					With: map[string]any{
						"path":           "./artifacts",
						"merge-multiple": false,
					},
				},
				{
					Name: "List Artifacts",
					Run:  "find ./artifacts -type f | sort",
				},
				{
					Name: "Publish Release",
					Uses: ActionGHRelease,
					With: map[string]any{
						"files":                   "./artifacts/**/*",
						"fail_on_unmatched_files": true,
					},
					Env: map[string]string{
						"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
					},
				},
			},
		},
	}
}
