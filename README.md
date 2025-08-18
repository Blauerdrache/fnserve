# FnServe

**Serverless, without the cloud.**

FnServe is a self-hosted, open-source serverless runtime. It brings AWS Lambda–style functions to your own machine or infrastructure, without Kubernetes, vendor lock-in, or hidden costs.

FnServe is part of the [HomeCloud](https://github.com/homecloudhq/homecloud) ecosystem, but designed to run standalone as well. You can use it independently, or integrate it into your HomeCloud stack when ready.

## Why FnServe?

* **Lightweight** — a single Go binary, no orchestration required
* **Language-agnostic** — write functions in Go, Python, or anything that speaks JSON
* **Local-first** — run on your laptop, homelab, or edge server
* **Practical** — built for small projects and real-world use, not demos

## Features

* Run functions once with JSON input/output
* Serve functions as HTTP endpoints
* Context injection (request ID, timestamp, deadlines)
* Works with Go and Python out of the box

## Quick Start

### Run a function

```bash
fnserve run ./functions/hello --event event.json
```

### Serve functions as HTTP endpoints

```bash
fnserve serve ./functions --port 8080
```

Functions in `./functions` will be available at:

```
POST /<function-name>
```

## Function Contract

* Input: JSON (stdin or HTTP body)
* Output: JSON (stdout or HTTP response)
* Must complete within configured timeout

## Roadmap (MVP)

* CLI: `run` and `serve` modes
* Go and Python runtime support
* Clear function contract
* HTTP serving with JSON I/O

## License

AGPL-3.0
