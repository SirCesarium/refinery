package cli

import (
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/app"
	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/pipeline"
	"github.com/SirCesarium/refinery/internal/ui"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [provider]",
	Short: "Generate CI/CD workflows",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.Section("Migration")
		providerName := args[0]
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			ui.Fatal(err, "Could not load 'refinery.toml'. Run 'refinery init' first.")
		}

		engineRegistry := app.NewDefaultEngineRegistry()
		eng := engineRegistry.Get(cfg.Project.Lang)
		if eng == nil {
			ui.Fatal(nil, "Unsupported language: "+cfg.Project.Lang)
		}

		providerRegistry, err := app.NewDefaultProviderRegistry()
		if err != nil {
			ui.Fatal(err, "Failed to initialize provider registry.")
		}

		provider := providerRegistry.Get(providerName)
		if provider == nil {
			ui.Fatal(nil, "Unsupported provider: "+providerName+". Supported: github")
		}

		ui.Info("Generating workflow for %s...", providerName)
		gen := pipeline.NewGenerator(provider, eng)
		data, err := gen.Generate(cfg)
		if err != nil {
			ui.Fatal(err, "CI workflow generation failed. Check your artifact configuration.")
		}

		outputPath := gen.Filename()
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			ui.Fatal(err, "Failed to create directory structure for: "+outputPath)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			ui.Fatal(err, "Failed to write workflow file: "+outputPath)
		}

		ui.Success("Workflow generated: %s", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
