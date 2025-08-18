# FnServe Examples

## Basic Usage

### Starting the Server with Concurrency Controls

```bash
# Start the server with 100 concurrent requests allowed
fnserve serve ./functions --port 8080 --concurrency 100 --timeout 30s

# Development mode with hot reload
fnserve dev ./functions --port 8080 --concurrency 10 --timeout 10s
```

### Running a Function with Context

```bash
# Run a Python function with event data
echo '{"name": "FnServe"}' | fnserve run ./functions/hello.py
```

## Creating Functions

### Python Function with Context Support

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
    event = json.load(sys.stdin)
    
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

### Invoking Functions via HTTP

```bash
# Basic request
curl -X POST http://localhost:8080/hello -d '{"name": "World"}'

# With query parameters
curl -X POST "http://localhost:8080/hello?version=1.0&debug=true" -d '{"name": "World"}'

# With tracing headers
curl -X POST http://localhost:8080/hello \
  -H "X-Trace-ID: trace-123" \
  -H "X-Parent-Span: span-456" \
  -d '{"name": "World"}'
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Server Statistics

```bash
curl http://localhost:8080/stats
```

Example response:
```json
{
  "active_requests": 2,
  "total_requests": 150,
  "success_requests": 145,
  "failed_requests": 3,
  "avg_execution_ms": 45
}
```
