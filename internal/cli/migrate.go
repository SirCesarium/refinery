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
	Short: "Generate CI/CD workflows",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var provider pipeline.CIProvider
		switch providerName {
		case "github":
			// Ahora manejamos el error que devuelve NewProvider
			p, err := github.NewProvider("refinery-build")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing provider: %v\n", err)
				os.Exit(1)
			}
			provider = p
		default:
			fmt.Fprintf(os.Stderr, "Unsupported provider: %s\n", providerName)
			os.Exit(1)
		}

		gen := pipeline.NewGenerator(provider)
		data, err := gen.Generate(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Generation failed: %v\n", err)
			os.Exit(1)
		}

		outputPath := gen.Filename()
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Workflow generated: %s\n", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
