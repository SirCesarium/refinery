package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/pipeline/github"
	"github.com/spf13/cobra"
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
			jobs := make(map[string]github.Job)
			for aName, art := range cfg.Artifacts {
				var include []map[string]any
				for tName, tCfg := range art.Targets {
					runsOn := "ubuntu-latest"
					switch tName {
					case "windows":
						runsOn = "windows-latest"
					case "darwin":
						runsOn = "macos-latest"
					}

					for _, targetArch := range tCfg.Archs {
						for _, targetAbi := range tCfg.ABIs {
							m := map[string]any{
								"os":      tName,
								"arch":    targetArch,
								"runs_on": runsOn,
							}
							if targetAbi != "" {
								m["abi"] = targetAbi
							}
							include = append(include, m)
						}
					}
				}

				jobs["build-"+aName] = github.Job{
					Name:   "Build " + aName,
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
							Uses: "actions/checkout@v6",
						},
						{
							Name: "Setup Go",
							Uses: "actions/setup-go@v6",
							With: map[string]any{
								"go-version": "stable",
							},
						},
						{
							Name: "Setup Rust",
							Uses: "actions-rust-lang/setup-rust-toolchain@v1",
						},
						{
							Name: "Install Refinery",
							Run:  "go install github.com/SirCesarium/refinery/cmd/refinery@main",
						},
						{
							Name:  "Build",
							Shell: "bash",
							Run:   fmt.Sprintf(`ABI_FLAG=""; if [ -n "${{ matrix.abi }}" ]; then ABI_FLAG="--abi ${{ matrix.abi }}"; fi; refinery build --artifact %s --os ${{ matrix.os }} --arch ${{ matrix.arch }} $ABI_FLAG`, aName),
						},
					},
				}
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
