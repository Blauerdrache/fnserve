package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type GoRuntime struct{}

func (r *GoRuntime) Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error) {
	// Create a temporary file for context
	ctxJSON, err := json.Marshal(fnCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal context: %w", err)
	}

	// Set up command with context
	cmd := exec.CommandContext(ctx, functionPath)
	cmd.Stdin = bytes.NewReader(event)

	// Pass context as environment variable
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("FN_CONTEXT=%s", ctxJSON))

	// Add all environment variables from context
	for k, v := range fnCtx.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Set up command execution with timeout
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	// Apply deadline if specified
	var timeout <-chan time.Time
	if fnCtx.Deadline > 0 {
		timeout = time.After(fnCtx.Deadline)
	}

	// Wait for completion or timeout
	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("go error: %s", out.String())
		}
	case <-timeout:
		// Try to kill the process
		cmd.Process.Kill()
		return nil, fmt.Errorf("function execution timed out after %v", fnCtx.Deadline)
	case <-ctx.Done():
		// Context was canceled
		cmd.Process.Kill()
		return nil, ctx.Err()
	}

	return out.Bytes(), nil
}
