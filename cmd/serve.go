package cmd

import (
	"github.com/spf13/cobra"

	"github.com/homecloudhq/fnserve/server"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Serve a directory of functions as HTTP endpoints",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		s := server.Server{Dir: dir}
		return s.Start(port)
	},
}

func init() {
	serveCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
