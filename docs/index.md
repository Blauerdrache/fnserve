# FnServe Documentation

Welcome to the FnServe documentation! FnServe is a lightweight, self-hosted serverless runtime that brings AWS Lambda–style functions to your own machine or infrastructure.

## Table of Contents

1. [Getting Started](../README.md) - Overview and basic usage
2. [Context and Concurrency](context-and-concurrency.md) - Details on context handling and concurrency features
3. [Runtime Support](runtime-support.md) - Information about runtime environments
4. [Testing and Development](testing-and-development.md) - Guide for testing and development

## Quick Links

- [Project Repository](https://github.com/homecloudhq/fnserve)
- [Examples](../examples/)

## Command Reference

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

## Getting Support

If you encounter any issues or have questions, please [open an issue](https://github.com/homecloudhq/fnserve/issues) on our GitHub repository.
