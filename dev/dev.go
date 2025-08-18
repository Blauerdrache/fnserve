package dev

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/homecloudhq/fnserve/server"
)

type DevServer struct {
	Dir  string
	Port int
}

func (d *DevServer) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	// Initial server start
	go func() {
		s := server.Server{Dir: d.Dir}
		go func() {
			if err := s.Start(d.Port); err != nil {
				log.Println("Server error:", err)
			}
		}()

		// Watch events
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf("[reload] %s modified, reloading...\n", event.Name)
					// For MVP, just print; full reload requires endpoint refresh
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
