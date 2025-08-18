# Runtime Support in FnServe

This document details how FnServe manages different runtime environments for executing functions.

## Supported Runtimes

FnServe currently supports two runtime environments out of the box:

1. **Python Runtime**: For Python function files (`.py`)
2. **Go Runtime**: For compiled Go binaries

Each runtime is responsible for executing functions in the appropriate environment and handling the communication of events and context information.

## Runtime Interface

All runtimes implement the common `Runtime` interface:

```go
type Runtime interface {
	Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error)
}
```

This interface provides a unified way to execute functions regardless of the underlying runtime environment.

## Python Runtime

### Detection and Execution

The Python runtime is used for files with a `.py` extension. It executes Python scripts using the `python3` interpreter.

```go
if strings.HasSuffix(functionPath, ".py") {
    rt = &runtime.PythonRuntime{}
}
```

### Implementation

The Python runtime passes the event data via stdin and the context via an environment variable:

```go
func (r *PythonRuntime) Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error) {
    // Marshal context to JSON
    ctxJSON, err := json.Marshal(fnCtx)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal context: %w", err)
    }

    // Set up command with timeout from context
    cmd := exec.CommandContext(ctx, "python3", functionPath)
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

    // Execute with timeout handling
    // ...

    return out.Bytes(), nil
}
```

### Python Function Structure

A Python function must:

1. Read event data from stdin
2. Parse context from `FN_CONTEXT` environment variable
3. Process the event
4. Write JSON output to stdout

```python
import sys, json, os

def handler(event, context):
    # Function logic here
    return {"result": "some data"}

if __name__ == "__main__":
    # Read event from stdin
    event = json.load(sys.stdin)
    
    # Get context from environment
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
    
    # Call handler and print JSON result
    result = handler(event, ctx)
    print(json.dumps(result))
```

## Go Runtime

### Detection and Execution

The Go runtime is used for files without a `.py` extension. It executes compiled Go binaries directly.

```go
if !strings.HasSuffix(functionPath, ".py") {
    rt = &runtime.GoRuntime{}
}
```

### Implementation

Similar to the Python runtime, the Go runtime passes the event via stdin and the context via an environment variable:

```go
func (r *GoRuntime) Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error) {
    // Marshal context to JSON
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

    // Execute with timeout handling
    // ...

    return out.Bytes(), nil
}
```

### Go Function Structure

A Go function must:

1. Read event data from stdin
2. Parse context from `FN_CONTEXT` environment variable
3. Process the event
4. Write JSON output to stdout

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "time"
)

type Event struct {
    Name string `json:"name"`
}

type Context struct {
    RequestID  string                 `json:"request_id"`
    Timestamp  time.Time              `json:"timestamp"`
    Deadline   time.Duration          `json:"deadline"`
    Parameters map[string]string      `json:"parameters"`
    Env        map[string]string      `json:"env"`
    Tracing    map[string]interface{} `json:"tracing"`
}

func main() {
    // Read input from stdin
    body, err := io.ReadAll(os.Stdin)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
        os.Exit(1)
    }

    // Parse the event
    var event Event
    if err := json.Unmarshal(body, &event); err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing event: %v\n", err)
        os.Exit(1)
    }

    // Get the context from environment
    var ctx Context
    if contextJSON := os.Getenv("FN_CONTEXT"); contextJSON != "" {
        json.Unmarshal([]byte(contextJSON), &ctx)
    }

    // Function logic here
    response := map[string]interface{}{
        "message": fmt.Sprintf("Hello, %s!", event.Name),
        "request_id": ctx.RequestID,
    }

    // Output the response as JSON
    output, _ := json.Marshal(response)
    fmt.Println(string(output))
}
```

## Timeout and Error Handling

Both runtimes implement timeout and error handling:

```go
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
        return nil, fmt.Errorf("%s error: %s", runtimeName, out.String())
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
```

## Extending with New Runtimes

To add support for additional runtimes:

1. Create a new runtime type that implements the `Runtime` interface
2. Add detection logic in the server to select the appropriate runtime
3. Implement the `Execute` method for the new runtime

For example, to add Node.js support:

```go
type NodeRuntime struct{}

func (r *NodeRuntime) Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error) {
    // Implementation similar to Python/Go runtimes
    // Using node to execute JavaScript files
    cmd := exec.CommandContext(ctx, "node", functionPath)
    // Rest of implementation...
}
```

Then update the runtime detection in the server:

```go
// Choose runtime
var rt runtime.Runtime
if strings.HasSuffix(functionPath, ".py") {
    rt = &runtime.PythonRuntime{}
} else if strings.HasSuffix(functionPath, ".js") {
    rt = &runtime.NodeRuntime{}
} else {
    rt = &runtime.GoRuntime{}
}
```

## Best Practices for Functions

### Python Functions

1. Use proper error handling with try/except blocks
2. Return JSON-serializable responses
3. Keep dependencies minimal to reduce startup time
4. Handle missing context values gracefully

### Go Functions

1. Compile your Go functions before deploying them
2. Use proper error handling
3. Return JSON-serializable responses
4. Keep the binary size small to reduce startup time

## Environment Requirements

### Python Runtime

- Python 3.6+ installed on the system
- Required Python packages installed

### Go Runtime

- Functions must be compiled for the target platform
- No additional runtime dependencies needed
