package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/homecloudhq/fnserve/server"
)

var (
	port           int
	maxConcurrency int
	requestTimeout time.Duration
	workerPoolSize int
)

var serveCmd = &cobra.Command{
	Use:   "serve [directory]",
	Short: "Serve a directory of functions as HTTP endpoints",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]

		// Create server with configuration
		s := server.NewServer(dir)

		// Apply configuration from flags
		s.Config.MaxConcurrentRequests = maxConcurrency
		s.Config.RequestTimeout = requestTimeout
		s.Config.WorkerPoolSize = workerPoolSize

		return s.Start(port)
	},
}

func init() {
	serveCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	serveCmd.Flags().IntVar(&maxConcurrency, "concurrency", 100, "Maximum number of concurrent function executions")
	serveCmd.Flags().DurationVar(&requestTimeout, "timeout", 30*time.Second, "Request timeout duration (e.g. 30s, 1m)")
	serveCmd.Flags().IntVar(&workerPoolSize, "workers", 10, "Size of the worker pool")

	rootCmd.AddCommand(serveCmd)
}
