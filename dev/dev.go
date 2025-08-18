package dev

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/homecloudhq/fnserve/server"
)

type DevServer struct {
	Dir         string
	Port        int
	Concurrency int
	Timeout     time.Duration
}

func (d *DevServer) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	// Track current server to allow for graceful restarts
	var currentServer *http.Server
	serverReady := make(chan bool, 1)

	// Server start/restart function
	startServer := func() {
		// Cancel previous server if it exists
		if currentServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			log.Println("Shutting down previous server...")
			currentServer.Shutdown(ctx)
		}

		// Create new server with config
		s := server.NewServer(d.Dir)

		// Apply configuration
		if d.Concurrency > 0 {
			s.Config.MaxConcurrentRequests = d.Concurrency
		}
		if d.Timeout > 0 {
			s.Config.RequestTimeout = d.Timeout
		}

		// Create HTTP server
		currentServer = &http.Server{
			Addr: fmt.Sprintf(":%d", d.Port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Custom handler that wraps the function server
				s := server.NewServer(d.Dir)

				// Start the server
				s.Start(d.Port)
			}),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		go func() {
			log.Printf("Starting development server on port %d\n", d.Port)
			serverReady <- true
			if err := s.Start(d.Port); err != nil && err != http.ErrServerClosed {
				log.Println("Server error:", err)
			}
		}()
	}

	// Initial server start
	go func() {
		// Start the server
		startServer()
		<-serverReady

		// Watch for file changes
		for {
			select {
			case <-serverCtx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("[reload] %s modified, reloading server...\n", event.Name)
					startServer()
					<-serverReady
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	err = watcher.Add(d.Dir)
	if err != nil {
		return err
	}

	<-done
	return nil
}
