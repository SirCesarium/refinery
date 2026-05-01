package cli

import (
	"fmt"
	"os"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/engine"
	"github.com/spf13/cobra"
)

var (
	artName string
	osName  string
	arch    string
	abi     string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an artifact based on configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("refinery.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		art, ok := cfg.Artifacts[artName]
		if !ok {
			fmt.Fprintf(os.Stderr, "Artifact %s not found in config\n", artName)
			os.Exit(1)
		}

		registry := engine.NewRegistry()
		registry.Register(&engine.RustEngine{})

		eng := registry.Get(cfg.Project.Lang)
		if eng == nil {
			fmt.Fprintf(os.Stderr, "No engine found for language: %s\n", cfg.Project.Lang)
			os.Exit(1)
		}

		opts := engine.BuildOptions{
			ArtifactName: artName,
			OS:           osName,
			Arch:         arch,
			ABI:          abi,
		}

		if err := eng.Build(cfg, art, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Build completed successfully")
	},
}

func init() {
	buildCmd.Flags().StringVar(&artName, "artifact", "", "Artifact name to build")
	buildCmd.Flags().StringVar(&osName, "os", "", "Target OS")
	buildCmd.Flags().StringVar(&arch, "arch", "", "Target Architecture")
	buildCmd.Flags().StringVar(&abi, "abi", "", "Target ABI")

	buildCmd.MarkFlagRequired("artifact")
	buildCmd.MarkFlagRequired("os")
	buildCmd.MarkFlagRequired("arch")

	rootCmd.AddCommand(buildCmd)
}
