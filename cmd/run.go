package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/homecloudhq/fnserve/runtime"
)

var eventJSON string

var runCmd = &cobra.Command{
	Use:   "run [function]",
	Short: "Run a function once with an event",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		functionPath := args[0]

		// Load event JSON (from flag or stdin)
		var event []byte
		if eventJSON != "" {
			event = []byte(eventJSON)
		} else {
			in, _ := os.ReadFile("/dev/stdin")
			event = in
		}

		ctx := runtime.Context{
			RequestID: "req-1234",
			Timestamp: time.Now(),
			Deadline:  30 * time.Second,
		}

		var r runtime.Runtime
		if hasExt(functionPath, ".py") {
			r = &runtime.PythonRuntime{}
		} else {
			r = &runtime.GoRuntime{}
		}

		result, err := r.Execute(functionPath, event, ctx)
		if err != nil {
			return err
		}

		fmt.Println(string(result))
		return nil
	},
}

func hasExt(path, ext string) bool {
	return len(path) > len(ext) && path[len(path)-len(ext):] == ext
}

func init() {
	runCmd.Flags().StringVar(&eventJSON, "event", "", "Event JSON to pass to function")
	rootCmd.AddCommand(runCmd)
}
