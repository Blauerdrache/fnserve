import sys, json, os, time

def handler(event, context):
    # Use context information in the response
    name = event.get("name", "World")
    
    response = {
        "message": f"Hello {name}",
        "request_id": context.get("request_id", "unknown"),
        "timestamp": context.get("timestamp", time.time()),
        "tracing": {
            "trace_id": context.get("tracing", {}).get("trace_id", "unknown"),
            "span_id": context.get("tracing", {}).get("span_id", "unknown")
        }
    }
    
    # Add any parameters passed via query string
    if "parameters" in context and context["parameters"]:
        response["params"] = context["parameters"]
        
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
