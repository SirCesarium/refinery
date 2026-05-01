package github

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"gopkg.in/yaml.v3"
)

const (
	ActionCheckout         = "actions/checkout@v6"
	ActionSetupGo          = "actions/setup-go@v6"
	ActionUploadArtifact   = "actions/upload-artifact@v7"
	ActionDownloadArtifact = "actions/download-artifact@v8"
	ActionGHRelease        = "softprops/action-gh-release@v3"
	ActionRustToolchain    = "actions-rust-lang/setup-rust-toolchain@v1"
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

func NewProvider(filename string) (*GithubProvider, error) {
	return &GithubProvider{filename: filename}, nil
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
			{Name: "Setup Rust", Run: ActionRustToolchain},
		}
	case "go":
		return []pipeline.Step{
			{Name: "Setup Go", Run: ActionSetupGo},
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
	uniqueMatrix := make(map[string]bool)

	for _, aName := range artifactNames {
		art := cfg.Artifacts[aName]
		for osName, tCfg := range art.Targets {
			runsOn := "ubuntu-latest"
			switch osName {
			case "windows":
				runsOn = "windows-latest"
			case "darwin":
				runsOn = "macos-latest"
			}

			for _, arch := range tCfg.Archs {
				abis := tCfg.ABIs
				if len(abis) == 0 {
					abis = []string{""}
				}

				for _, abi := range abis {
					key := fmt.Sprintf("%s-%s-%s-%s", aName, osName, arch, abi)
					if uniqueMatrix[key] {
						continue
					}
					uniqueMatrix[key] = true

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

	buildSteps := []Step{
		{Name: "Checkout", Uses: ActionCheckout},
		{
			Name: "Setup Go",
			Uses: ActionSetupGo,
			With: map[string]any{"go-version": "stable", "cache": true},
		},
	}

	for _, s := range p.GetSetupSteps(cfg.Project.Lang) {
		if s.Run != "" && s.Name != "Setup Go" {
			buildSteps = append(buildSteps, Step{
				Name: s.Name,
				Uses: s.Run,
				With: map[string]any{"cache": true},
			})
		}
	}

	buildSteps = append(buildSteps, Step{
		Name: "Install Refinery",
		Run:  "go install github.com/SirCesarium/refinery/cmd/refinery@main",
	})

	buildSteps = append(buildSteps, Step{
		Name:  "Add GOBIN to PATH",
		Run:   "echo \"$(go env GOPATH)/bin\" >> $GITHUB_PATH",
		Shell: "bash",
	})

	buildSteps = append(buildSteps, Step{
		Name: "Build Artifact",
		Uses: "SirCesarium/refinery@main",
		With: map[string]any{
			"artifact": "${{ matrix.artifact }}",
			"os":       "${{ matrix.os }}",
			"arch":     "${{ matrix.arch }}",
			"abi":      "${{ matrix.abi }}",
		},
	})

	buildSteps = append(buildSteps, Step{
		Name: "Upload",
		Uses: ActionUploadArtifact,
		With: map[string]any{
			"name":              "bin-${{ matrix.artifact }}-${{ matrix.os }}-${{ matrix.arch }}${{ matrix.abi && format('-{0}', matrix.abi) }}",
			"path":              "dist/*",
			"if-no-files-found": "error",
			"compression-level": 0,
		},
	})

	jobs := map[string]Job{
		"build": {
			Name:   "Build ${{ matrix.artifact }} (${{ matrix.os }}-${{ matrix.arch }})",
			RunsOn: "${{ matrix.runs_on }}",
			Strategy: &Strategy{
				FailFast: false,
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
					Name: "Download",
					Uses: ActionDownloadArtifact,
					With: map[string]any{
						"path":           "./artifacts",
						"merge-multiple": true,
					},
				},
				{
					Name: "Publish",
					Uses: ActionGHRelease,
					With: map[string]any{
						"files":                   "./artifacts/*",
						"fail_on_unmatched_files": true,
					},
					Env: map[string]string{
						"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
					},
				},
			},
		},
	}

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
