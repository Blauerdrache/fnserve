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
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id"`
	TraceInfo  map[string]interface{} `json:"trace_info,omitempty"`
	Parameters map[string]string      `json:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
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
		if err := json.Unmarshal([]byte(contextJSON), &ctx); err != nil {
			// Context parse error is not fatal
			fmt.Fprintf(os.Stderr, "Warning: Error parsing context: %v\n", err)
		}
	}

	// Get name or use default
	name := event.Name
	if name == "" {
		name = "World"
	}

	// Create the response
	response := Response{
		Message:    fmt.Sprintf("Hello, %s from Go!", name),
		RequestID:  ctx.RequestID,
		Timestamp:  time.Now(),
		Parameters: ctx.Parameters,
		TraceInfo:  ctx.Tracing,
	}

	// Output the response as JSON
	output, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
