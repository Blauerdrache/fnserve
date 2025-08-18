package cmd

import (
	"time"

	"github.com/homecloudhq/fnserve/dev"
	"github.com/spf13/cobra"
)

var (
	devPort           int
	devConcurrency    int
	devRequestTimeout time.Duration
)

var devCmd = &cobra.Command{
	Use:   "dev [directory]",
	Short: "Run in development mode with hot-reload",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		d := dev.DevServer{
			Dir:         dir,
			Port:        devPort,
			Concurrency: devConcurrency,
			Timeout:     devRequestTimeout,
		}
		return d.Start()
	},
}

func init() {
	devCmd.Flags().IntVar(&devPort, "port", 8080, "Port to listen on")
	devCmd.Flags().IntVar(&devConcurrency, "concurrency", 10, "Maximum number of concurrent function executions")
	devCmd.Flags().DurationVar(&devRequestTimeout, "timeout", 30*time.Second, "Request timeout duration (e.g. 30s, 1m)")
	rootCmd.AddCommand(devCmd)
}
