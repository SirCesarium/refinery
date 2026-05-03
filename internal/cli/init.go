// Package cli provides command-line interface commands for Refinery.
package cli

import (
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/SirCesarium/refinery/internal/ui"
	"github.com/spf13/cobra"
)

// Force overwrite of existing refinery.toml.
var force bool

// initCmd initializes a new Refinery project.
var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new refinery project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ui.Section("Initialization")
		// Get current directory as default project name.
		workingDir, err := os.Getwd()
		if err != nil {
			ui.Fatal(err, "Failed to determine current working directory. Check your permissions.")
		}

		projectName := filepath.Base(workingDir)
		if len(args) > 0 {
			projectName = args[0]
		}

		// Check for existing config unless force flag is set.
		if !force {
			if _, err := os.Stat("refinery.toml"); err == nil {
				ui.Fatal(nil, "refinery.toml already exists. Use --force to overwrite if you are sure.")
			}
		}

		cfg := config.Default(projectName)

		if err := cfg.Write("refinery.toml"); err != nil {
			ui.Fatal(err, "Failed to write 'refinery.toml'. Ensure you have write permissions in this directory.")
		}

		ui.Success("Successfully initialized refinery project: %s", projectName)
		ui.Info("You can now edit 'refinery.toml' to define your artifacts.")
	},
}

func init() {
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of existing refinery.toml")
	rootCmd.AddCommand(initCmd)
}
