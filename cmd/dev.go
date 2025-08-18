package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev [directory]",
	Short: "Run in development mode with hot-reload",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		fmt.Printf("Dev mode started in: %s (stub)\n", dir)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
