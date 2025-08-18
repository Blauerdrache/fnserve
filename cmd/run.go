package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [function]",
	Short: "Run a function once with an event",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		functionPath := args[0]
		fmt.Printf("Running function: %s (stub)\n", functionPath)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
