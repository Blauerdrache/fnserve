package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Serve a directory of functions as HTTP endpoints",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		fmt.Printf("Serving functions from: %s (stub)\n", dir)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
