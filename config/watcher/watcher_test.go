package watcher

import (
	"os"
	"testing"
	"time"
)

func TestVSCodeWatcher(t *testing.T) {
	ctx := t.Context()

	tmpfile, err := os.CreateTemp("", "mcpconfig-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	watcher, err := NewVSCodeWatcher(ctx, tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	clients := watcher.Clients()
	if len(clients) != 0 {
		t.Fatalf("expected 0 clients, got %d", len(clients))
	}

	initialConfig := `{
		"servers": {
			"server1": {
				"type": "http",
				"url": "http://localhost:8080",
				"headers": {
					"Authorization": "Bearer token"
				}
			}
		}
	}`

	if _, err := tmpfile.Write([]byte(initialConfig)); err != nil {
		t.Fatal(err)
	}
	if err = tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Give some time for the watcher to pick up the change
	time.Sleep(time.Second)

	clients = watcher.Clients()
	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}

	if _, ok := clients["server1"]; !ok {
		t.Fatalf("expected client 'server1' to exist")
	}

	updatedConfig := `{
		"servers": {
			"server2": {
				"type": "stdio",
				"command": "dummy-command"
			}
		}
	}`

	if err := os.WriteFile(tmpfile.Name(), []byte(updatedConfig), 0o644); err != nil {
		t.Fatal(err)
	}

	// Give some time for the watcher to pick up the change
	time.Sleep(time.Second)

	clients = watcher.Clients()
	if len(clients) != 1 {
		t.Fatalf("expected 1 client after update, got %d", len(clients))
	}

	if _, ok := clients["server2"]; !ok {
		t.Fatalf("expected client 'server2' to exist after update")
	}
}
