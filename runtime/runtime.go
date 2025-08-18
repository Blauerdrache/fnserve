package runtime

import (
	"time"
)

type Context struct {
	RequestID string        `json:"request_id"`
	Timestamp time.Time     `json:"timestamp"`
	Deadline  time.Duration `json:"deadline"`
}

type Runtime interface {
	Execute(functionPath string, event []byte, ctx Context) ([]byte, error)
}
