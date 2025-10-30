package config

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/njayp/chimera/proxy"
)

type Config interface {
	Clients() proxy.Clients
}

type Watcher[T Config] struct {
	sync.RWMutex
	clients proxy.Clients
}

func NewWatcher[T Config](ctx context.Context, path string) (*Watcher[T], error) {
	w := &Watcher[T]{}
	w.updateClients(path)
	return w, w.start(ctx, path)
}

func (w *Watcher[T]) start(ctx context.Context, path string) error {
	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if event.Has(fsnotify.Write) {
					w.updateClients(path)
				}
				if event.Has(fsnotify.Create) {
					w.updateClients(path)
				}
				if event.Has(fsnotify.Remove) {
					w.updateClients(path)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					continue
				}
				slog.Error(err.Error())
			case <-ctx.Done():
				err := watcher.Close()
				if err != nil {
					slog.Error("failed to close file watcher", "error", err)
				}
				return
			}
		}
	}()

	// Add file to watch
	return watcher.Add(path)
}

// Clients returns the current set of clients.
func (w *Watcher[T]) Clients() proxy.Clients {
	w.RLock()
	defer w.RUnlock()
	return w.clients
}

func (w *Watcher[T]) updateClients(path string) {
	w.Lock()
	defer w.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read config file", "error", err)
		w.clients = nil
		return
	}

	var config T
	if err := json.Unmarshal(data, config); err != nil {
		slog.Error("failed to parse JSON config", "error", err)
		w.clients = nil
		return
	}

	w.clients = config.Clients()
}
