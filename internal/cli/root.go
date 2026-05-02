package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refinery",
	Short: "Refinery: Multi-ecosystem artifact orchestrator",
	Long: `Refinery simplifies building, packaging, and distributing software artifacts
across different languages and platforms.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
