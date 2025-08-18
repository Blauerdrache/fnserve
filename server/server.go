package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/homecloudhq/fnserve/runtime"
)

type ServerConfig struct {
	MaxConcurrentRequests int
	RequestTimeout        time.Duration
	WorkerPoolSize        int
}

type Server struct {
	Dir    string
	Config ServerConfig

	// For concurrency control
	semaphore    chan struct{}
	functionPool sync.Pool
	stats        Stats
}

type Stats struct {
	sync.Mutex
	ActiveRequests   int
	TotalRequests    int64
	SuccessRequests  int64
	FailedRequests   int64
	TotalExecutionMs int64
}

// NewServer creates a new server with default configuration
func NewServer(dir string) *Server {
	return &Server{
		Dir: dir,
		Config: ServerConfig{
			MaxConcurrentRequests: 100,
			RequestTimeout:        30 * time.Second,
			WorkerPoolSize:        10,
		},
		stats: Stats{},
	}
}

func (s *Server) Start(port int) error {
	// Initialize concurrency control
	s.semaphore = make(chan struct{}, s.Config.MaxConcurrentRequests)

	mux := http.NewServeMux()

	// Discover functions in directory
	files, err := os.ReadDir(s.Dir)
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	// Register health check and stats endpoint
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/stats", s.handleStats)

	// Register function endpoints
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		functionPath := filepath.Join(s.Dir, f.Name())

		// Register endpoint
		mux.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
			// Concurrency control - acquire semaphore or reject if too many requests
			select {
			case s.semaphore <- struct{}{}:
				// Got the semaphore, continue
				defer func() { <-s.semaphore }()
			default:
				// Too many requests
				w.WriteHeader(429)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error":"too many requests"}`))
				return
			}

			start := time.Now()
			reqID := uuid.NewString()
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = uuid.NewString()
			}

			// Update stats
			s.stats.Lock()
			s.stats.ActiveRequests++
			s.stats.TotalRequests++
			s.stats.Unlock()

			// Create context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), s.Config.RequestTimeout)
			defer cancel()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"invalid request body"}`, 400)
				s.recordFailure(start)
				return
			}

			// Extract parameters from query string and headers
			params := make(map[string]string)
			for key, values := range r.URL.Query() {
				if len(values) > 0 {
					params[key] = values[0]
				}
			}

			// Environment variables to pass to function
			env := make(map[string]string)
			for _, header := range []string{"X-API-Key", "Authorization", "X-Forwarded-For"} {
				if val := r.Header.Get(header); val != "" {
					env[header] = val
				}
			}

			fnCtx := runtime.Context{
				RequestID:  reqID,
				Timestamp:  time.Now(),
				Deadline:   s.Config.RequestTimeout,
				Parameters: params,
				Env:        env,
				Tracing: runtime.TracingInfo{
					TraceID:  traceID,
					SpanID:   uuid.NewString(),
					ParentID: r.Header.Get("X-Parent-Span"),
				},
			}

			// Choose runtime
			var rt runtime.Runtime
			if strings.HasSuffix(functionPath, ".py") {
				rt = &runtime.PythonRuntime{}
			} else {
				rt = &runtime.GoRuntime{}
			}

			// Execute function with context
			result, err := rt.Execute(ctx, functionPath, body, fnCtx)
			if err != nil {
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"error": %q}`, err.Error())
				s.recordFailure(start)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Request-ID", reqID)
			w.Header().Set("X-Trace-ID", traceID)
			w.Write(result)

			duration := time.Since(start)
			s.recordSuccess(start)

			fmt.Printf("[req=%s] %s %s (%s)\n", reqID, r.Method, r.URL.Path, duration)
		})
	}

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("FnServe listening on %s (max concurrent: %d)\n",
		server.Addr, s.Config.MaxConcurrentRequests)
	return server.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	s.stats.Lock()
	defer s.stats.Unlock()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"active_requests": %d,
		"total_requests": %d,
		"success_requests": %d,
		"failed_requests": %d,
		"avg_execution_ms": %d
	}`,
		s.stats.ActiveRequests,
		s.stats.TotalRequests,
		s.stats.SuccessRequests,
		s.stats.FailedRequests,
		s.calcAvgExecutionTime(),
	)
}

func (s *Server) recordSuccess(startTime time.Time) {
	duration := time.Since(startTime)

	s.stats.Lock()
	defer s.stats.Unlock()

	s.stats.ActiveRequests--
	s.stats.SuccessRequests++
	s.stats.TotalExecutionMs += duration.Milliseconds()
}

func (s *Server) recordFailure(startTime time.Time) {
	duration := time.Since(startTime)

	s.stats.Lock()
	defer s.stats.Unlock()

	s.stats.ActiveRequests--
	s.stats.FailedRequests++
	s.stats.TotalExecutionMs += duration.Milliseconds()
}

func (s *Server) calcAvgExecutionTime() int64 {
	total := s.stats.SuccessRequests + s.stats.FailedRequests
	if total == 0 {
		return 0
	}
	return s.stats.TotalExecutionMs / total
}
