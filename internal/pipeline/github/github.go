package github

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"gopkg.in/yaml.v3"
)

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

func NewProvider(filename string) *GithubProvider {
	return &GithubProvider{filename: filename}
}

func (p *GithubProvider) Name() string {
	return "github"
}

func (p *GithubProvider) Filename() string {
	return filepath.Join(".github", "workflows", fmt.Sprintf("%s.yml", p.filename))
}

func (p *GithubProvider) GetSetupSteps(lang string) []pipeline.Step {
	switch lang {
	case "rust":
		return []pipeline.Step{
			{Name: "Setup Rust", Run: "actions-rust-lang/setup-rust-toolchain@v1"},
		}
	case "go":
		return []pipeline.Step{
			{Name: "Setup Go", Run: "actions/setup-go@v6"},
		}
	default:
		return nil
	}
}

func (p *GithubProvider) Generate(cfg *config.Config) ([]byte, error) {
	var artifactNames []string
	for name := range cfg.Artifacts {
		artifactNames = append(artifactNames, name)
	}
	sort.Strings(artifactNames)

	var include []map[string]any
	for _, aName := range artifactNames {
		art := cfg.Artifacts[aName]
		for osName, tCfg := range art.Targets {
			runsOn := "ubuntu-latest"
			if osName == "windows" {
				runsOn = "windows-latest"
			} else if osName == "darwin" {
				runsOn = "macos-latest"
			}

			for _, arch := range tCfg.Archs {
				for _, abi := range tCfg.ABIs {
					m := map[string]any{
						"artifact": aName,
						"os":       osName,
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

	steps := []Step{
		{Name: "Checkout", Uses: "actions/checkout@v6"},
	}

	for _, s := range p.GetSetupSteps(cfg.Project.Lang) {
		if s.Run != "" {
			steps = append(steps, Step{Name: s.Name, Uses: s.Run})
		}
	}

	steps = append(steps, Step{
		Name: "Install Refinery",
		Run:  "go install github.com/SirCesarium/refinery/cmd/refinery@main",
	})

	steps = append(steps, Step{
		Name: "Build",
		Uses: "SirCesarium/refinery@main",
		With: map[string]any{
			"artifact": "${{ matrix.artifact }}",
			"os":       "${{ matrix.os }}",
			"arch":     "${{ matrix.arch }}",
			"abi":      "${{ matrix.abi }}",
		},
	})

	wf := Workflow{
		Name: "Refinery Build",
		On: On{
			Push: &Event{Tags: []string{"v*"}},
		},
		Jobs: map[string]Job{
			"build": {
				RunsOn: "${{ matrix.runs_on }}",
				Strategy: &Strategy{
					FailFast: false,
					Matrix:   map[string]any{"include": include},
				},
				Steps: steps,
			},
		},
	}

	return yaml.Marshal(wf)
}
