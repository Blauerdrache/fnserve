# FnServe

**Serverless, without the cloud.**

FnServe is a self-hosted, open-source serverless runtime. It brings AWS Lambda–style functions to your own machine or infrastructure, without Kubernetes, vendor lock-in, or hidden costs.

FnServe is part of the [HomeCloud](https://github.com/homecloudhq/homecloud) ecosystem, but designed to run standalone as well. You can use it independently, or integrate it into your HomeCloud stack when ready.

## Why FnServe?

* **Lightweight** — a single Go binary, no orchestration required
* **Language-agnostic** — write functions in Go, Python, or anything that speaks JSON
* **Local-first** — run on your laptop, homelab, or edge server
* **Practical** — built for small projects and real-world use, not demos
* **Scalable** — built-in concurrency controls and monitoring
* **Developer-friendly** — hot reload in development mode

## Features

* Run functions once with JSON input/output
* Serve functions as HTTP endpoints
* Development mode with hot reload
* Advanced context injection (request ID, timestamp, deadlines, env vars, parameters)
* Tracing support with trace IDs, span IDs, and parent IDs
* Concurrency controls for high-scale deployments
* Health and statistics endpoints
* Works with Go and Python out of the box
* Graceful shutdown and error handling
* Function timeouts and cancellation

## Installation

### From source

```bash
git clone https://github.com/homecloudhq/fnserve.git
cd fnserve
go build -o fnserve .
```

### Using Go install

```bash
go install github.com/homecloudhq/fnserve@latest
```

## Quick Start

### Run a function once

```bash
# Using a JSON string
fnserve run ./functions/hello.py --event '{"name": "World"}'

# Using a JSON file
fnserve run ./functions/hello.py --event event.json

# Using stdin
echo '{"name": "World"}' | fnserve run ./functions/hello.py
```

### Serve functions as HTTP endpoints

```bash
# Basic usage
fnserve serve ./functions --port 8080

# With concurrency controls
fnserve serve ./functions --port 8080 --concurrency 100 --timeout 30s
```

### Development mode with hot reload

```bash
fnserve dev ./functions --port 8080 --concurrency 10 --timeout 10s
```

Functions in `./functions` will be available at:

```
POST /<function-name>
```

## CLI Commands

### Root Command

```bash
fnserve [command]
```

Available Commands:
- `run`: Run a function once with an event
- `serve`: Serve a directory of functions as HTTP endpoints
- `dev`: Run in development mode with hot-reload

### Run Command

```bash
fnserve run [function] [flags]
```

Flags:
- `--event string`: Event JSON to pass to function (can be a JSON string or a file path)

### Serve Command

```bash
fnserve serve [directory] [flags]
```

Flags:
- `--port int`: Port to listen on (default 8080)
- `--concurrency int`: Maximum number of concurrent function executions (default 100)
- `--timeout duration`: Request timeout duration (default 30s)
- `--workers int`: Size of the worker pool (default 10)

### Dev Command

```bash
fnserve dev [directory] [flags]
```

Flags:
- `--port int`: Port to listen on (default 8080)
- `--concurrency int`: Maximum number of concurrent function executions (default 10)
- `--timeout duration`: Request timeout duration (default 30s)

## Function Contract

### Input/Output

* Input: JSON (stdin or HTTP body)
* Output: JSON (stdout or HTTP response)
* Must complete within configured timeout

### Context

Functions receive a context object with the following structure:

```json
{
  "request_id": "req-uuid",
  "timestamp": "2025-08-18T12:00:00Z",
  "deadline": 30000000000,
  "parameters": {
    "param1": "value1",
    "param2": "value2"
  },
  "env": {
    "X-API-Key": "api-key-value",
    "Authorization": "auth-header"
  },
  "tracing": {
    "trace_id": "trace-uuid",
    "span_id": "span-uuid",
    "parent_id": "parent-span-id"
  }
}
```

## Creating Functions

### Python Functions

```python
import sys, json, os

def handler(event, context):
    # Access context information
    request_id = context.get("request_id", "unknown")
    trace_id = context.get("tracing", {}).get("trace_id", "unknown")
    
    # Access request parameters
    params = context.get("parameters", {})
    
    return {
        "message": f"Hello {event.get('name', 'World')}",
        "request_id": request_id,
        "trace_id": trace_id,
        "params": params
    }

if __name__ == "__main__":
    # Load event from stdin
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        event = {}
    
    # Get context from environment variable
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result))
```

### Go Functions

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

type Response struct {
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
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

	// Create response
	response := Response{
		Message:   fmt.Sprintf("Hello, %s!", event.Name),
		RequestID: ctx.RequestID,
		Timestamp: time.Now(),
	}

	// Output JSON response
	output, _ := json.Marshal(response)
	fmt.Println(string(output))
}
```

## HTTP API

### Function Endpoints

Each function in your functions directory is exposed at:

```
POST /<function-name>
```

#### Request

- Method: POST
- Body: JSON payload
- Headers:
  - `X-Trace-ID`: Optional trace ID for distributed tracing
  - `X-Parent-Span`: Optional parent span ID for tracing
  - `X-API-Key`: Optional API key (passed to function)
  - `Authorization`: Optional auth header (passed to function)

#### Query Parameters

All query parameters are passed to the function context:

```
POST /hello?param1=value1&param2=value2
```

### Health Check Endpoint

```
GET /health
```

Response:
```json
{"status":"ok"}
```

### Statistics Endpoint

```
GET /stats
```

Response:
```json
{
  "active_requests": 2,
  "total_requests": 150,
  "success_requests": 145,
  "failed_requests": 3,
  "avg_execution_ms": 45
}
```

## Concurrency and Scaling

FnServe provides built-in concurrency controls:

- **Concurrent Requests Limit**: Control the maximum number of concurrent function executions
- **Request Timeout**: Set execution time limits per function
- **Worker Pool**: Configure the worker pool size for function execution
- **Graceful Shutdown**: Properly handle in-flight requests during shutdown

## Tracing and Monitoring

- **Request IDs**: Every function execution gets a unique ID
- **Trace and Span IDs**: Support for distributed tracing
- **Execution Statistics**: Track success, failure, and timing metrics

## Examples

See the [examples directory](./examples) for more detailed usage examples.

## Roadmap

- [ ] Support for more runtimes (Node.js, Ruby, etc.)
- [ ] Local function deployment
- [ ] Function versioning
- [ ] Environment management
- [ ] Authentication and authorization
- [ ] Function logs and metrics storage

## License

AGPL-3.0
