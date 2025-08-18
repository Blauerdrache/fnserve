import sys, json, os, time

def handler(event, context):
    # Extract name from event or use default
    name = event.get("name", "World")
    
    # Access context information
    request_id = context.get("request_id", "unknown")
    timestamp = context.get("timestamp", time.time())
    
    # Access tracing info
    tracing = context.get("tracing", {})
    trace_id = tracing.get("trace_id", "unknown")
    span_id = tracing.get("span_id", "unknown")
    
    # Get parameters from query string
    params = context.get("parameters", {})
    
    # Access environment variables
    env_vars = context.get("env", {})
    
    # Create response with context info
    response = {
        "message": f"Hello {name}!",
        "request_info": {
            "id": request_id,
            "timestamp": timestamp,
        },
        "tracing": {
            "trace_id": trace_id,
            "span_id": span_id
        }
    }
    
    # Include parameters if present
    if params:
        response["parameters"] = params
    
    # Add selective environment info if present (careful with security)
    if "X-API-Key" in env_vars:
        response["auth"] = "API key provided"
    
    return response

if __name__ == "__main__":
    # Load event from stdin
    event = json.load(sys.stdin)
    
    # Try to get context from environment variable
    ctx = {}
    if "FN_CONTEXT" in os.environ:
        try:
            ctx = json.loads(os.environ["FN_CONTEXT"])
        except:
            pass
            
    result = handler(event, ctx)
    print(json.dumps(result))
