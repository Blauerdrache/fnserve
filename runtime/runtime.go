package runtime

import (
	"context"
	"time"
)

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

type Runtime interface {
	Execute(ctx context.Context, functionPath string, event []byte, fnCtx Context) ([]byte, error)
}
