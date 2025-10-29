package vscode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVSCode_ValidConfig(t *testing.T) {
	content := `{
		"servers": {
			"test-stdio": {
				"type": "stdio",
				"command": "echo",
				"args": ["hello"],
				"env": {
					"FOO": "bar"
				}
			},
			"test-http": {
				"type": "http",
				"url": "http://localhost:8080",
				"headers": {
					"Authorization": "Bearer token"
				}
			}
		}
	}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := Clients(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers) != 2 {
		t.Errorf("expected 1 stdio server, got %d", len(servers))
	}
}

func TestVSCode_InvalidJSON(t *testing.T) {
	content := `{"servers": {invalid json}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	_, err := Clients(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestVSCode_FileNotFound(t *testing.T) {
	_, err := Clients("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestVSCode_UnsupportedServerType(t *testing.T) {
	content := `{
		"servers": {
			"test": {
				"type": "websocket",
				"url": "ws://localhost:8080"
			}
		}
	}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := Clients(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers) != 0 {
		t.Errorf("expected 1 stdio server, got %d", len(servers))
	}
}

func TestVSCode_WithInputs(t *testing.T) {
	content := `{
		"inputs": [
			{
				"type": "promptString",
				"id": "apiKey",
				"description": "API Key",
				"password": true
			}
		],
		"servers": {
			"test": {
				"type": "stdio",
				"command": "test"
			}
		}
	}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := Clients(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("expected 1 stdio server, got %d", len(servers))
	}
}

func createTempFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tmpFile
}
