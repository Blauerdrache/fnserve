# Testing and Development Guide for FnServe

This guide provides detailed instructions for testing and developing with FnServe, including examples of working with the context and concurrency features.

## Testing FnServe

### Testing with the `run` Command

The `run` command allows you to execute functions directly from the command line, which is useful for testing and development.

#### Basic Function Testing

```bash
# Create a test event file
mkdir -p test
echo '{"name": "FnServe Tester"}' > test/event.json

# Run a Python function with the event file
fnserve run ./functions/hello.py --event test/event.json

# Run a function with direct JSON input
fnserve run ./functions/hello.py --event '{"name": "Direct JSON"}'

# Pipe JSON to the function
echo '{"name": "Piped JSON"}' | fnserve run ./functions/hello.py
```

#### Testing Context Information

Create a test function that outputs the context:

```python
# context_test.py
import sys, json, os

def handler(event, context):
    # Return the entire context object
    return {
        "event": event,
        "context": context
    }

if __name__ == "__main__":
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        event = {}
    
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result, default=str))
```

Run the test function:

```bash
fnserve run ./functions/context_test.py --event '{"test": "data"}'
```

This will output the event and context data, allowing you to verify that the context is being correctly passed to the function.

### Testing with the `serve` Command

The `serve` command starts an HTTP server that exposes functions as endpoints.

#### Starting the Server

```bash
fnserve serve ./functions --port 8080 --concurrency 100 --timeout 30s
```

#### Testing HTTP Endpoints

Use curl to test the HTTP endpoints:

```bash
# Basic request
curl -X POST http://localhost:8080/hello -d '{"name": "World"}'

# With query parameters
curl -X POST "http://localhost:8080/hello?version=1.0&debug=true" -d '{"name": "World"}'

# With headers
curl -X POST http://localhost:8080/hello \
  -H "X-Trace-ID: test-trace-123" \
  -H "X-Parent-Span: test-span-456" \
  -H "X-API-Key: test-api-key" \
  -d '{"name": "World"}'
```

#### Testing Health and Stats Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Statistics
curl http://localhost:8080/stats
```

### Testing with the `dev` Command

The `dev` command is similar to `serve` but includes hot-reload capability for development.

#### Starting Development Mode

```bash
fnserve dev ./functions --port 8080 --concurrency 5 --timeout 10s
```

#### Testing Hot Reload

1. Start the dev server as above
2. Modify a function file
3. The server should detect the change and reload automatically
4. Test the modified function by making an HTTP request

```bash
# While the dev server is running, modify a function
echo "# Adding a comment" >> ./functions/hello.py

# The server should reload
# Then test the function
curl -X POST http://localhost:8080/hello -d '{"name": "Hot Reload Test"}'
```

## Testing Concurrency Features

### Testing Concurrency Limits

To test the concurrency limits, send multiple requests simultaneously:

```bash
# Create a test script
cat > test_concurrency.sh << 'EOF'
#!/bin/bash
for i in {1..20}; do
  curl -s -X POST http://localhost:8080/hello -d '{"name":"Concurrent'$i'"}' &
done
wait
EOF

chmod +x test_concurrency.sh
./test_concurrency.sh
```

If you've set the concurrency limit to a low value (e.g., 5), some requests should receive a 429 "Too Many Requests" response.

### Testing Request Timeouts

Create a function that sleeps longer than the timeout:

```python
# timeout_test.py
import sys, json, os, time

def handler(event, context):
    # Sleep for a specified duration
    sleep_time = event.get("sleep", 15)  # Default 15 seconds
    time.sleep(sleep_time)
    
    return {
        "message": f"Slept for {sleep_time} seconds",
        "request_id": context.get("request_id", "unknown")
    }

if __name__ == "__main__":
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        event = {}
    
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result))
```

Test the timeout:

```bash
# Start the server with a short timeout
fnserve serve ./functions --timeout 5s --port 8080

# Send a request that will timeout
curl -X POST http://localhost:8080/timeout_test -d '{"sleep": 10}'
```

The request should fail with a timeout error.

## Testing Tracing Features

### Creating Traced Requests

To test tracing, create a chain of requests with trace IDs:

```bash
# Generate a trace ID
TRACE_ID="trace-$(date +%s)"

# First request with trace ID
SPAN1_ID="span-1"
curl -X POST http://localhost:8080/hello \
  -H "X-Trace-ID: $TRACE_ID" \
  -H "X-Parent-Span: root" \
  -d '{"name": "Trace Test 1"}'

# Second request with same trace ID but new span ID
SPAN2_ID="span-2"
curl -X POST http://localhost:8080/hello \
  -H "X-Trace-ID: $TRACE_ID" \
  -H "X-Parent-Span: $SPAN1_ID" \
  -d '{"name": "Trace Test 2"}'
```

## Development Workflows

### Local Development Workflow

1. Start the development server:
   ```bash
   fnserve dev ./functions --port 8080
   ```

2. Create or modify functions in the `./functions` directory.

3. Test functions via HTTP:
   ```bash
   curl -X POST http://localhost:8080/function_name -d '{"key": "value"}'
   ```

4. Functions are automatically reloaded when files change.

### Creating a New Function

#### Python Function

```bash
cat > ./functions/new_function.py << 'EOF'
import sys, json, os

def handler(event, context):
    name = event.get("name", "World")
    return {"message": f"Hello, {name} from new function!"}

if __name__ == "__main__":
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        event = {}
    
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result))
EOF

# Test the new function
fnserve run ./functions/new_function.py --event '{"name": "Developer"}'
```

#### Go Function

```bash
# Create a new Go function
mkdir -p ./functions/go_func
cat > ./functions/go_func/main.go << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func main() {
	// Read input from stdin
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Parse the event
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing event: %v\n", err)
		os.Exit(1)
	}

	// Get name from event
	name, ok := event["name"].(string)
	if !ok {
		name = "World"
	}

	// Create response
	response := map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s from Go!", name),
	}

	// Output JSON response
	output, _ := json.Marshal(response)
	fmt.Println(string(output))
}
EOF

# Build the Go function
cd ./functions/go_func
go build -o ../go_function .
cd ../..

# Test the Go function
fnserve run ./functions/go_function --event '{"name": "Go Developer"}'
```

## Advanced Testing

### Load Testing

For load testing, you can use tools like `wrk` or `siege`:

```bash
# Install wrk
sudo apt-get install -y wrk

# Create a request file
echo '{"name": "Load Test"}' > test/payload.json

# Run load test (10 threads, 100 connections, 30 seconds)
wrk -t10 -c100 -d30s -s test/payload.json http://localhost:8080/hello
```

### Testing Statistics Aggregation

To verify that statistics are being aggregated correctly:

1. Start the server:
   ```bash
   fnserve serve ./functions --port 8080
   ```

2. Send multiple requests:
   ```bash
   for i in {1..50}; do
     curl -s -X POST http://localhost:8080/hello -d '{"name":"Test'$i'"}' > /dev/null
   done
   ```

3. Check the statistics:
   ```bash
   curl http://localhost:8080/stats
   ```

4. Send some requests that will fail:
   ```bash
   for i in {1..5}; do
     curl -s -X POST http://localhost:8080/nonexistent -d '{"name":"Test'$i'"}' > /dev/null
   done
   ```

5. Check the statistics again to verify the failed requests are counted:
   ```bash
   curl http://localhost:8080/stats
   ```

## Debugging Tips

### Debugging Python Functions

1. Add print statements for debugging:
   ```python
   import sys
   print("Debug info", file=sys.stderr)
   ```

2. Use the `run` command for direct debugging:
   ```bash
   fnserve run ./functions/problematic_function.py --event '{"debug": true}'
   ```

### Debugging Go Functions

1. Add print statements for debugging:
   ```go
   fmt.Fprintf(os.Stderr, "Debug info: %+v\n", someVar)
   ```

2. Use environment variables for configuration:
   ```bash
   DEBUG=true fnserve run ./functions/go_function --event '{"debug": true}'
   ```

### Inspecting Context

Create a function that dumps all context information:

```python
# debug_context.py
import sys, json, os

def handler(event, context):
    # Return everything we know
    return {
        "event": event,
        "context": context,
        "env": dict(os.environ),
    }

if __name__ == "__main__":
    try:
        event = json.load(sys.stdin)
    except json.JSONDecodeError:
        event = {}
    
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result, default=str))
```

Run this function to see all available context information:

```bash
fnserve run ./functions/debug_context.py --event '{"debug": true}'
```
