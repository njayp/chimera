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
	ToClients() proxy.Clients
}

// Watcher watches a configuration file for changes and reloads it.
// T should not be a pointer.
type Watcher[T Config] struct {
	sync.RWMutex
	path string
	// clients are stored so they can be reused
	clients proxy.Clients
}

// New creates a new Watcher.
func New[T Config](ctx context.Context, path string) (*Watcher[T], error) {
	w := &Watcher[T]{
		path: path,
	}
	w.update()
	return w, w.start(ctx)
}

// NewVSCodeWatcher creates a new Watcher for VSCode MCP configuration files.
func NewVSCodeWatcher(ctx context.Context, path string) (*Watcher[vscode.Config], error) {
	return New[vscode.Config](ctx, path)
}

func (w *Watcher[T]) start(ctx context.Context) error {
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
					w.update()
				}
				if event.Has(fsnotify.Create) {
					w.update()
				}
				if event.Has(fsnotify.Remove) {
					w.update()
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
	return watcher.Add(w.path)
}

// Clients returns the current set of clients.
func (w *Watcher[T]) Clients() proxy.Clients {
	w.RLock()
	defer w.RUnlock()
	return w.clients
}

func (w *Watcher[T]) update() {
	data, err := os.ReadFile(w.path)
	if err != nil {
		slog.Error("failed to read config file", "error", err)
		return
	}

	config := new(T)
	if err := json.Unmarshal(data, config); err != nil {
		slog.Error("failed to parse JSON config", "error", err)
		return
	}

	w.Lock()
	defer w.Unlock()
	w.clients = (*config).ToClients()
}
