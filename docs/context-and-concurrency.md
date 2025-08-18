# Context and Concurrency in FnServe

This document provides detailed information about the context handling and concurrency features in FnServe.

## Context Handling

FnServe provides rich context information to functions, similar to AWS Lambda's context object. This context includes request metadata, tracing information, and environment variables.

### Context Object Structure

```go
type Context struct {
    RequestID  string            `json:"request_id"`
    Timestamp  time.Time         `json:"timestamp"`
    Deadline   time.Duration     `json:"deadline"`
    Parameters map[string]string `json:"parameters"`
    Env        map[string]string `json:"env"`
    Tracing    TracingInfo       `json:"tracing"`
}

type TracingInfo struct {
    TraceID  string `json:"trace_id"`
    SpanID   string `json:"span_id"`
    ParentID string `json:"parent_id"`
}
```

### How Context is Passed

For both Python and Go functions, the context is:

1. Marshaled to JSON
2. Passed via environment variable `FN_CONTEXT`
3. Additionally, individual environment variables from the `Env` map are set directly

### Context in HTTP Requests

When functions are called via HTTP:

- Query parameters are added to the `Parameters` map
- Headers like `X-API-Key`, `Authorization`, and `X-Forwarded-For` are added to the `Env` map
- Tracing headers are used to populate the `Tracing` field

### Context in CLI Execution

When using the `run` command:

- A unique request ID is generated
- Tracing information is automatically created
- Deadline is set to the default 30 seconds

## Concurrency Control

FnServe implements several mechanisms for concurrency control and scaling.

### Semaphore-Based Concurrency Limiting

The server uses a semaphore channel to limit the number of concurrent function executions:

```go
// Initialize concurrency control
s.semaphore = make(chan struct{}, s.Config.MaxConcurrentRequests)

// In request handler
select {
case s.semaphore <- struct{}{}:
    // Got the semaphore, continue
    defer func() { <-s.semaphore }()
default:
    // Too many requests
    w.WriteHeader(429)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"error":"too many requests"}`))
    return
}
```

### Timeouts and Cancellation

Function execution has multiple timeout and cancellation mechanisms:

1. **Context Cancellation**: Using Go's `context.Context` for cancellation propagation
2. **Execution Timeout**: Based on the function's deadline
3. **HTTP Server Timeouts**: Read, write, and idle timeouts

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(r.Context(), s.Config.RequestTimeout)
defer cancel()

// In runtime execution
select {
case err := <-done:
    // Function completed
case <-timeout:
    // Timeout occurred
    cmd.Process.Kill()
    return nil, fmt.Errorf("function execution timed out after %v", fnCtx.Deadline)
case <-ctx.Done():
    // Context was canceled
    cmd.Process.Kill()
    return nil, ctx.Err()
}
```

### Worker Pool

A configurable worker pool helps manage concurrent execution:

```go
s.Config.WorkerPoolSize = workerPoolSize // Configure pool size
```

## Statistics and Monitoring

FnServe tracks execution statistics to provide insights into function performance.

### Statistics Collected

- Active requests count
- Total requests processed
- Successful requests count
- Failed requests count
- Total execution time
- Average execution time

### Statistics Implementation

```go
type Stats struct {
    sync.Mutex
    ActiveRequests   int
    TotalRequests    int64
    SuccessRequests  int64
    FailedRequests   int64
    TotalExecutionMs int64
}

func (s *Server) recordSuccess(startTime time.Time) {
    duration := time.Since(startTime)
    
    s.stats.Lock()
    defer s.stats.Unlock()
    
    s.stats.ActiveRequests--
    s.stats.SuccessRequests++
    s.stats.TotalExecutionMs += duration.Milliseconds()
}

func (s *Server) recordFailure(startTime time.Time) {
    duration := time.Since(startTime)
    
    s.stats.Lock()
    defer s.stats.Unlock()
    
    s.stats.ActiveRequests--
    s.stats.FailedRequests++
    s.stats.TotalExecutionMs += duration.Milliseconds()
}
```

## Development Mode

The development mode provides hot-reload capability for functions.

### How Hot Reload Works

1. Uses the `fsnotify` library to watch for file changes
2. When a function file changes, the server gracefully restarts
3. Maintains the same configuration during restart

```go
if event.Op&fsnotify.Write == fsnotify.Write {
    fmt.Printf("[reload] %s modified, reloading server...\n", event.Name)
    startServer()
    <-serverReady
}
```

## Best Practices

### Function Development

1. **Error Handling**: Always handle errors in your functions and return appropriate error responses
2. **Context Usage**: Access context values with fallbacks for missing values
3. **Timeouts**: Be aware of configured timeouts and ensure your function completes within the deadline

### Server Configuration

1. **Concurrency**: Set concurrency limits appropriate for your hardware
2. **Timeouts**: Configure timeouts based on expected function execution time
3. **Worker Pool**: Adjust worker pool size based on the number of available CPU cores

### Performance Optimization

1. **Keep Functions Small**: Smaller functions are faster to start and execute
2. **Minimize Dependencies**: Fewer dependencies mean faster startup times
3. **Use Appropriate Memory Limits**: Right-size memory allocation for your functions

## Command Line Configuration

### Serve Mode

```bash
fnserve serve ./functions \
    --port 8080 \
    --concurrency 100 \
    --timeout 30s \
    --workers 10
```

### Development Mode

```bash
fnserve dev ./functions \
    --port 8080 \
    --concurrency 10 \
    --timeout 10s
```

## Advanced Usage Examples

### Function with Custom Timeout

```python
import sys, json, os, time

def handler(event, context):
    # Get deadline from context
    deadline = context.get("deadline", 30)
    
    # Simulate work
    work_time = min(deadline * 0.8, 10)  # Use 80% of deadline or max 10 seconds
    time.sleep(work_time)
    
    return {
        "message": f"Completed work in {work_time} seconds",
        "deadline": deadline,
        "request_id": context.get("request_id", "unknown")
    }

if __name__ == "__main__":
    event = json.load(sys.stdin)
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        ctx = json.loads(os.environ["FN_CONTEXT"])
    result = handler(event, ctx)
    print(json.dumps(result))
```
