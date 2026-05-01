package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SirCesarium/refinery/internal/config"
	"github.com/spf13/cobra"
)

var force bool

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new refinery project",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}

		projectName := filepath.Base(workingDir)
		if len(args) > 0 {
			projectName = args[0]
		}

		if !force {
			if _, err := os.Stat("refinery.toml"); err == nil {
				fmt.Println("Error: refinery.toml already exists. Use --force to overwrite.")
				return
			}
		}

		cfg := config.Default(projectName)

		if err := cfg.Write("refinery.toml"); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully initialized refinery project: %s\n", projectName)
	},
}

func init() {
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite of existing refinery.toml")
	rootCmd.AddCommand(initCmd)
}
