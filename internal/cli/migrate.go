package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/pipeline/github"
	"github.com/spf13/cobra"
)

const (
	ActionCheckout  = "actions/checkout@v6"
	ActionSetupGo   = "actions/setup-go@v6"
	ActionSetupRust = "actions-rust-lang/setup-rust-toolchain@v1"
	ActionUpload    = "actions/upload-artifact@v7"
	ActionDownload  = "actions/download-artifact@v8"
	ActionRelease   = "softprops/action-gh-release@v3"
	RefineryAction  = "SirCesarium/refinery@main"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [provider]",
	Short: "Generate CI/CD workflows using a specific provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		var provider pipeline.CIProvider
		switch providerName {
		case "github":
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

			jobs := make(map[string]github.Job)
			jobs["build"] = github.Job{
				Name:   "Build ${{ matrix.artifact }} (${{ matrix.os }}-${{ matrix.arch }}${{ matrix.abi && format('-{0}', matrix.abi) }})",
				RunsOn: "${{ matrix.runs_on }}",
				Strategy: &github.Strategy{
					FailFast: false,
					Matrix: map[string]any{
						"include": include,
					},
				},
				Steps: []github.Step{
					{
						Name: "Checkout",
						Uses: ActionCheckout,
					},
					{
						Name: "Setup Go",
						Uses: ActionSetupGo,
						With: map[string]any{"go-version": "stable", "cache": true},
					},
					{
						Name: "Setup Rust",
						Uses: ActionSetupRust,
						With: map[string]any{"cache": true},
					},
					{
						Name: "Install Refinery",
						Run:  "go install github.com/SirCesarium/refinery/cmd/refinery@main",
						Env: map[string]string{
							"GOPROXY":   "https://proxy.golang.org,direct",
							"GOPRIVATE": "github.com/SirCesarium/*",
						},
					},
					{
						Name:  "Add GOBIN to PATH",
						Run:   "echo \"$(go env GOPATH)/bin\" >> $GITHUB_PATH",
						Shell: "bash",
					},
					{
						Name: "Build Artifact",
						Uses: RefineryAction,
						With: map[string]any{
							"artifact": "${{ matrix.artifact }}",
							"os":       "${{ matrix.os }}",
							"arch":     "${{ matrix.arch }}",
							"abi":      "${{ matrix.abi }}",
						},
					},
					{
						Name: "Upload",
						Uses: ActionUpload,
						With: map[string]any{
							"name":              "bin-${{ matrix.artifact }}-${{ matrix.os }}-${{ matrix.arch }}${{ matrix.abi && format('-{0}', matrix.abi) }}",
							"path":              "dist/*",
							"if-no-files-found": "error",
							"compression-level": 0,
						},
					},
				},
			}

			jobs["release"] = github.Job{
				Name:   "Release Artifacts",
				Needs:  []string{"build"},
				RunsOn: "ubuntu-latest",
				If:     "startsWith(github.ref, 'refs/tags/')",
				Steps: []github.Step{
					{
						Name: "Download",
						Uses: ActionDownload,
						With: map[string]any{
							"path":           "./artifacts",
							"merge-multiple": true,
						},
					},
					{
						Name: "Publish",
						Uses: ActionRelease,
						With: map[string]any{
							"files":                   "./artifacts/*",
							"fail_on_unmatched_files": true,
						},
						Env: map[string]string{
							"GITHUB_TOKEN": "${{ secrets.GITHUB_TOKEN }}",
						},
					},
				},
			}

			wf := github.Workflow{
				Name: "Refinery Build Pipeline",
				On: github.On{
					Push: &github.Event{
						Tags: []string{"v*"},
					},
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
			provider, _ = github.NewProvider(wf, "refinery-build")
		default:
			fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerName)
			os.Exit(1)
		}

		gen := pipeline.NewGenerator(provider)
		data, err := gen.Generate(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating pipeline: %v\n", err)
			os.Exit(1)
		}

		outputPath := gen.Filename()
		dir := filepath.Dir(outputPath)
		os.MkdirAll(dir, 0755)
		os.WriteFile(outputPath, data, 0644)
		fmt.Printf("Workflows generated successfully using %s provider at %s\n", provider.Name(), outputPath)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
