package config

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

	servers, err := VSCode(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers.StdioServers) != 1 {
		t.Errorf("expected 1 stdio server, got %d", len(servers.StdioServers))
	}

	if len(servers.HTTPServers) != 1 {
		t.Errorf("expected 1 http server, got %d", len(servers.HTTPServers))
	}

	stdioServer := servers.StdioServers["test-stdio"]
	if stdioServer.Command != "echo" {
		t.Errorf("expected command 'echo', got '%s'", stdioServer.Command)
	}

	if len(stdioServer.Args) != 1 || stdioServer.Args[0] != "hello" {
		t.Errorf("expected args ['hello'], got %v", stdioServer.Args)
	}

	if len(stdioServer.Env) != 1 || stdioServer.Env[0] != "FOO=bar" {
		t.Errorf("expected env ['FOO=bar'], got %v", stdioServer.Env)
	}

	httpServer := servers.HTTPServers["test-http"]
	if httpServer.URL != "http://localhost:8080" {
		t.Errorf("expected url 'http://localhost:8080', got '%s'", httpServer.URL)
	}

	if httpServer.Headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header 'Bearer token', got '%s'", httpServer.Headers["Authorization"])
	}
}

func TestVSCode_EmptyServers(t *testing.T) {
	content := `{"servers": {}}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := VSCode(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers.StdioServers) != 0 {
		t.Errorf("expected 0 stdio servers, got %d", len(servers.StdioServers))
	}

	if len(servers.HTTPServers) != 0 {
		t.Errorf("expected 0 http servers, got %d", len(servers.HTTPServers))
	}
}

func TestVSCode_InvalidJSON(t *testing.T) {
	content := `{"servers": {invalid json}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	_, err := VSCode(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestVSCode_FileNotFound(t *testing.T) {
	_, err := VSCode("/nonexistent/file.json")
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

	_, err := VSCode(tmpFile)
	if err == nil {
		t.Fatal("expected error for unsupported server type, got nil")
	}
}

func TestVSCode_MultipleEnvVars(t *testing.T) {
	content := `{
		"servers": {
			"test": {
				"type": "stdio",
				"command": "test",
				"env": {
					"VAR1": "value1",
					"VAR2": "value2",
					"VAR3": "value3"
				}
			}
		}
	}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := VSCode(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stdioServer := servers.StdioServers["test"]
	if len(stdioServer.Env) != 3 {
		t.Errorf("expected 3 env vars, got %d", len(stdioServer.Env))
	}

	// Check that all env vars are present (order doesn't matter)
	envMap := make(map[string]bool)
	for _, env := range stdioServer.Env {
		envMap[env] = true
	}

	expectedEnvs := []string{"VAR1=value1", "VAR2=value2", "VAR3=value3"}
	for _, expected := range expectedEnvs {
		if !envMap[expected] {
			t.Errorf("expected env var '%s' not found", expected)
		}
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

	servers, err := VSCode(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(servers.StdioServers) != 1 {
		t.Errorf("expected 1 stdio server, got %d", len(servers.StdioServers))
	}
}

func TestVSCode_ReturnsCorrectType(t *testing.T) {
	content := `{"servers": {}}`

	tmpFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tmpFile) }()

	servers, err := VSCode(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_ = servers
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
