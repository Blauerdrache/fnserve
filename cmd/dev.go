package cmd

import (
	"github.com/homecloudhq/fnserve/dev"
	"github.com/spf13/cobra"
)

var devPort int

var devCmd = &cobra.Command{
	Use:   "dev [directory]",
	Short: "Run in development mode with hot-reload",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		d := dev.DevServer{Dir: dir, Port: devPort}
		return d.Start()
	},
}

func init() {
	devCmd.Flags().IntVar(&devPort, "port", 8080, "Port to listen on")
	rootCmd.AddCommand(devCmd)
}
