import sys, json

def handler(event, context):
    return {"message": "Hello " + event["name"]}

if __name__ == "__main__":
    event = json.load(sys.stdin)
    ctx = {}  # context injection later
    result = handler(event, ctx)
    print(json.dumps(result))
