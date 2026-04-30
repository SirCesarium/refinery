package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new refinery project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("refinery initialized")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
