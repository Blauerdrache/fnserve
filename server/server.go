package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/homecloudhq/fnserve/runtime"
)

type Server struct {
	Dir string
}

func (s *Server) Start(port int) error {
	mux := http.NewServeMux()

	// Discover functions in directory
	files, err := os.ReadDir(s.Dir)
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		functionPath := filepath.Join(s.Dir, f.Name())

		// Register endpoint
		mux.HandleFunc("/"+name, func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := uuid.NewString()

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, `{"error":"invalid request body"}`, 400)
				return
			}

			ctx := runtime.Context{
				RequestID: reqID,
				Timestamp: time.Now(),
				Deadline:  30 * time.Second,
			}

			// Choose runtime
			var rt runtime.Runtime
			if strings.HasSuffix(functionPath, ".py") {
				rt = &runtime.PythonRuntime{}
			} else {
				rt = &runtime.GoRuntime{}
			}

			result, err := rt.Execute(functionPath, body, ctx)
			if err != nil {
				w.WriteHeader(500)
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"error": %q}`, err.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(result)

			duration := time.Since(start)
			fmt.Printf("[req=%s] %s %s (%s)\n", reqID, r.Method, r.URL.Path, duration)
		})
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("FnServe listening on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}
