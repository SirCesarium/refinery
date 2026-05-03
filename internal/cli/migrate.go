package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/app"
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine/rust"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/ui"
	"github.com/spf13/cobra"
)

var dryRun bool

var migrateCmd = &cobra.Command{
	Use:   "migrate [provider]",
	Short: "Generate CI/CD workflows with validation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.Section("Migration")

		// Extract provider name from command arguments.
		providerName := args[0]

		// Load refinery configuration from 'refinery.toml'.
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			ui.Fatal(err, "Could not load 'refinery.toml'. Run 'refinery init' first.")
		}

		// Get engine for the project language.
		engineRegistry := app.NewDefaultEngineRegistry()
		eng := engineRegistry.Get(cfg.Project.Lang)
		if eng == nil {
			ui.Fatal(nil, "Unsupported language: "+cfg.Project.Lang)
		}

		// Validate configuration against the engine.
		if err := eng.Validate(cfg); err != nil {
			ui.Fatal(err, "Configuration validation failed. Fix refinery.toml before migrating.")
		}

		// Run Rust-specific validation if applicable.
		if cfg.Project.Lang == "rust" {
			if rustEngine, ok := eng.(*rust.RustEngine); ok {
				ui.Info("Running Rust-specific validation...")
				if err := rustEngine.ValidateRustSpecific(cfg); err != nil {
					ui.Fatal(err, "Rust validation failed. Check Cargo.toml matches refinery.toml")
				}
			}
		}

		// Initialize provider registry and get the requested provider.
		providerRegistry, err := app.NewDefaultProviderRegistry()
		if err != nil {
			ui.Fatal(err, "Failed to initialize provider registry.")
		}

		provider := providerRegistry.Get(providerName)
		if provider == nil {
			ui.Fatal(nil, "Unsupported provider: "+providerName+". Supported: github")
		}

		// Generate CI workflow data.
		ui.Info("Generating workflow for %s...", providerName)
		gen := pipeline.NewGenerator(provider, eng)
		data, err := gen.Generate(cfg)
		if err != nil {
			ui.Fatal(err, "CI workflow generation failed. Check your artifact configuration.")
		}

		// Handle dry run: print workflow and exit.
		if dryRun {
			ui.Info("Dry run - generated workflow:")
			fmt.Println(string(data))
			ui.Success("Dry run completed. No files were written.")
			return
		}

		// Determine output path and ensure directory exists.
		outputPath := gen.Filename()
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			ui.Fatal(err, "Failed to create directory structure for: "+outputPath)
		}

		// Write workflow file if content has changed.
		if existing, err := os.ReadFile(outputPath); err == nil && string(existing) == string(data) {
			ui.Info("Workflow unchanged, skipping write.")
		} else {
			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				ui.Fatal(err, "Failed to write workflow file: "+outputPath)
			}
			ui.Success("Workflow generated: %s", outputPath)
		}
		ui.Info("Review the workflow file before pushing to avoid CI failures.")
	},
}

func init() {
	migrateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview generated workflow without writing")
	rootCmd.AddCommand(migrateCmd)
}
