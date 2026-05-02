package github

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
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

func (p *GithubProvider) Generate(cfg *config.Config, eng engine.BuildEngine) ([]byte, error) {
	if err := eng.Validate(cfg); err != nil {
		return nil, err
	}

	var artifactNames []string
	for name := range cfg.Artifacts {
		artifactNames = append(artifactNames, name)
	}
	sort.Strings(artifactNames)

	var include []map[string]any
	uniqueMatrix := make(map[string]bool)

	for _, aName := range artifactNames {
		art := cfg.Artifacts[aName]
		for _, tCfg := range art.Targets {
			runsOn := "ubuntu-latest"
			switch tCfg.OS {
			case "windows":
				runsOn = "windows-latest"
			case "darwin":
				runsOn = "macos-latest"
			}

			for _, arch := range tCfg.Archs {
				abis := tCfg.ABIs
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

	buildSteps := []Step{
		{Name: "Checkout", Uses: ActionCheckout},
	}

	for _, req := range eng.GetCIRequirements(cfg) {
		switch {
		case req == "rust":
			buildSteps = append(buildSteps, Step{
				Name: "Setup Rust",
				Uses: ActionRustToolchain,
				With: map[string]any{"cache": true},
			})
		case req == "cross-linker:linux-aarch64":
			buildSteps = append(buildSteps, Step{
				Name: "Install ARM Linker",
				If:   "runner.os == 'Linux'",
				Run:  "sudo apt-get update && sudo apt-get install -y gcc-aarch64-linux-gnu",
			})
		case req == "pkg:musl-tools":
			buildSteps = append(buildSteps, Step{
				Name: "Install Musl Tools",
				If:   "runner.os == 'Linux'",
				Run:  "sudo apt-get update && sudo apt-get install -y musl-tools",
			})
		case req == "pkg:cargo-deb":
			buildSteps = append(buildSteps, Step{
				Name: "Install cargo-deb",
				If:   "runner.os == 'Linux'",
				Run:  "cargo install cargo-deb",
			})
		case req == "pkg:cargo-generate-rpm":
			buildSteps = append(buildSteps, Step{
				Name: "Install cargo-generate-rpm",
				If:   "runner.os == 'Linux'",
				Run:  "cargo install cargo-generate-rpm",
			})
		case req == "pkg:cargo-wix":
			buildSteps = append(buildSteps, Step{
				Name: "Install cargo-wix",
				If:   "runner.os == 'Windows'",
				Run:  "cargo install cargo-wix",
			})
		}
	}

	buildSteps = append(buildSteps, Step{
		Name: "Build Artifact",
		Uses: "SirCesarium/refinery@main",
		With: map[string]any{
			"artifact": "${{ matrix.artifact }}",
			"os":       "${{ matrix.os }}",
			"arch":     "${{ matrix.arch }}",
			"abi":      "${{ matrix.abi }}",
			"version":  cfg.RefineryVersion,
		},
	})

	buildSteps = append(buildSteps, Step{
		Name: "Upload",
		Uses: ActionUploadArtifact,
		With: map[string]any{
			"name":              "bin-${{ matrix.artifact }}-${{ matrix.os }}-${{ matrix.arch }}${{ matrix.abi && format('-{0}', matrix.abi) }}",
			"path":              cfg.OutputDir + "/*",
			"if-no-files-found": "error",
			"compression-level": 0,
		},
	})

	jobs := map[string]Job{
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
