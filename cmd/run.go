package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
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
			// Check if the provided string is a file path
			if _, err := os.Stat(eventJSON); err == nil {
				// It's a file, read its contents
				fileContent, err := os.ReadFile(eventJSON)
				if err != nil {
					return fmt.Errorf("failed to read event file: %w", err)
				}
				event = fileContent
			} else {
				// It's not a file, treat as raw JSON string
				event = []byte(eventJSON)
			}
		} else {
			// Only try to read from stdin if no event flag provided
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Data is being piped to stdin
				in, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				event = in
			} else {
				// No data from stdin and no event flag - use empty object
				event = []byte("{}")
			}
		}

		// Create context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		fnCtx := runtime.Context{
			RequestID:  "req-" + uuid.New().String(),
			Timestamp:  time.Now(),
			Deadline:   30 * time.Second,
			Parameters: map[string]string{},
			Env:        map[string]string{},
			Tracing: runtime.TracingInfo{
				TraceID: uuid.New().String(),
				SpanID:  uuid.New().String(),
			},
		}

		var r runtime.Runtime
		if hasExt(functionPath, ".py") {
			r = &runtime.PythonRuntime{}
		} else {
			r = &runtime.GoRuntime{}
		}

		result, err := r.Execute(ctx, functionPath, event, fnCtx)
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
