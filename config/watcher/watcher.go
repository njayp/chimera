package watcher

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/njayp/chimera/config/vscode"
	"github.com/njayp/chimera/proxy"
)

// Config represents the file structure, and it must be able to produce Clients.
type Config interface {
	Clients() proxy.Clients
}

// Watcher watches a configuration file for changes and reloads it.
// T should not be a pointer.
type Watcher[T Config] struct {
	sync.RWMutex
	// clients are stored so they can be reused
	clients proxy.Clients
}

// NewVSCodeWatcher creates a new Watcher for VSCode MCP configuration files.
func NewVSCodeWatcher(ctx context.Context, path string) (*Watcher[vscode.Config], error) {
	w := &Watcher[vscode.Config]{}
	w.update(path)
	return w, w.start(ctx, path)
}

func (w *Watcher[T]) start(ctx context.Context, path string) error {
	// Create new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Listen for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}
				if event.Has(fsnotify.Write) {
					w.update(path)
				}
				if event.Has(fsnotify.Create) {
					w.update(path)
				}
				if event.Has(fsnotify.Remove) {
					w.update(path)
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

func (w *Watcher[T]) update(path string) {
	config := new(T)

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read config file", "error", err)
		return
	}

	if err := json.Unmarshal(data, config); err != nil {
		slog.Error("failed to parse JSON config", "error", err)
		return
	}

	w.Lock()
	defer w.Unlock()
	w.clients = (*config).Clients()
}
