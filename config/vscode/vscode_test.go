package vscode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClients_MixedServers(t *testing.T) {
	config := `{
		"servers": {
			"stdio-server": {
				"type": "stdio",
				"command": "/usr/bin/stdio",
				"args": ["start"]
			},
			"http-server": {
				"type": "http",
				"url": "http://example.com"
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}

	if _, ok := clients["stdio-server"]; !ok {
		t.Error("expected 'stdio-server' client")
	}

	if _, ok := clients["http-server"]; !ok {
		t.Error("expected 'http-server' client")
	}
}

func TestClients_EmptyConfig(t *testing.T) {
	config := `{
		"servers": {}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(clients) != 0 {
		t.Fatalf("expected 0 clients, got %d", len(clients))
	}
}

func TestClients_InvalidJSON(t *testing.T) {
	config := `{
		"servers": {
			"test": {
				"type": "stdio"
			}
		}
	` // missing closing brace

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	_, err := Clients(tmpfile)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestClients_FileNotFound(t *testing.T) {
	_, err := Clients("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestClients_UnsupportedServerType(t *testing.T) {
	config := `{
		"servers": {
			"unknown-server": {
				"type": "websocket",
				"url": "ws://example.com"
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(clients) != 0 {
		t.Fatalf("expected 0 clients for unsupported type, got %d", len(clients))
	}
}

func TestClients_Function(t *testing.T) {
	config := Config{
		Servers: map[string]Server{
			"stdio1": {
				Type:    "stdio",
				Command: "/usr/bin/stdio1",
				Args:    []string{"arg1"},
				Env:     map[string]string{"KEY": "value"},
			},
			"http1": {
				Type:    "http",
				URL:     "http://example.com",
				Headers: map[string]string{"Auth": "token"},
			},
			"unknown": {
				Type: "grpc",
			},
		},
	}

	result := clients(config)

	if len(result) != 2 {
		t.Fatalf("expected 2 clients (unknown type ignored), got %d", len(result))
	}

	if _, ok := result["stdio1"]; !ok {
		t.Error("expected 'stdio1' client")
	}

	if _, ok := result["http1"]; !ok {
		t.Error("expected 'http1' client")
	}

	if _, ok := result["unknown"]; ok {
		t.Error("did not expect 'unknown' client with unsupported type")
	}
}

// Helper function to create temporary config files
func createTempConfig(t *testing.T, content string) (string, func()) {
	t.Helper()

	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "config.json")

	if err := os.WriteFile(tmpfile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	cleanup := func() {
		// t.TempDir() automatically cleans up
	}

	return tmpfile, cleanup
}
