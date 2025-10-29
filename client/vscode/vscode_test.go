package vscode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/njayp/chimera/client"
)

func TestClients_ValidStdioConfig(t *testing.T) {
	config := `{
		"servers": {
			"test-server": {
				"type": "stdio",
				"command": "/usr/bin/test",
				"args": ["arg1", "arg2"],
				"env": {
					"KEY1": "value1",
					"KEY2": "value2"
				}
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}

	c, ok := clients["test-server"]
	if !ok {
		t.Fatal("expected 'test-server' client")
	}

	stdio, ok := c.(client.Stdio)
	if !ok {
		t.Fatal("expected Stdio client")
	}

	if stdio.Command != "/usr/bin/test" {
		t.Errorf("expected command '/usr/bin/test', got '%s'", stdio.Command)
	}

	if len(stdio.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(stdio.Args))
	}

	if stdio.Args[0] != "arg1" || stdio.Args[1] != "arg2" {
		t.Errorf("expected args [arg1, arg2], got %v", stdio.Args)
	}

	if len(stdio.Env) != 2 {
		t.Fatalf("expected 2 env vars, got %d", len(stdio.Env))
	}

	expectedEnv := map[string]bool{
		"KEY1=value1": false,
		"KEY2=value2": false,
	}
	for _, env := range stdio.Env {
		if _, ok := expectedEnv[env]; ok {
			expectedEnv[env] = true
		}
	}
	for env, found := range expectedEnv {
		if !found {
			t.Errorf("expected env var '%s' not found", env)
		}
	}
}

func TestClients_ValidHTTPConfig(t *testing.T) {
	config := `{
		"servers": {
			"http-server": {
				"type": "http",
				"url": "http://example.com/mcp",
				"headers": {
					"Authorization": "Bearer token123",
					"X-Custom": "custom-value"
				}
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(clients))
	}

	c, ok := clients["http-server"]
	if !ok {
		t.Fatal("expected 'http-server' client")
	}

	httpClient, ok := c.(client.HTTP)
	if !ok {
		t.Fatal("expected HTTP client")
	}

	if httpClient.URL != "http://example.com/mcp" {
		t.Errorf("expected URL 'http://example.com/mcp', got '%s'", httpClient.URL)
	}

	if len(httpClient.Headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(httpClient.Headers))
	}

	if httpClient.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected Authorization header 'Bearer token123', got '%s'", httpClient.Headers["Authorization"])
	}

	if httpClient.Headers["X-Custom"] != "custom-value" {
		t.Errorf("expected X-Custom header 'custom-value', got '%s'", httpClient.Headers["X-Custom"])
	}
}

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

func TestClients_StdioWithoutOptionalFields(t *testing.T) {
	config := `{
		"servers": {
			"simple-server": {
				"type": "stdio",
				"command": "/usr/bin/simple"
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	c, ok := clients["simple-server"]
	if !ok {
		t.Fatal("expected 'simple-server' client")
	}

	stdio, ok := c.(client.Stdio)
	if !ok {
		t.Fatal("expected Stdio client")
	}

	if stdio.Command != "/usr/bin/simple" {
		t.Errorf("expected command '/usr/bin/simple', got '%s'", stdio.Command)
	}

	if len(stdio.Args) != 0 {
		t.Errorf("expected no args, got %d", len(stdio.Args))
	}

	if len(stdio.Env) != 0 {
		t.Errorf("expected no env vars, got %d", len(stdio.Env))
	}
}

func TestClients_HTTPWithoutHeaders(t *testing.T) {
	config := `{
		"servers": {
			"simple-http": {
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

	c, ok := clients["simple-http"]
	if !ok {
		t.Fatal("expected 'simple-http' client")
	}

	httpClient, ok := c.(client.HTTP)
	if !ok {
		t.Fatal("expected HTTP client")
	}

	if httpClient.URL != "http://example.com" {
		t.Errorf("expected URL 'http://example.com', got '%s'", httpClient.URL)
	}

	if len(httpClient.Headers) != 0 {
		t.Errorf("expected no headers, got %d", len(httpClient.Headers))
	}
}

func TestClients_ComplexEnvironment(t *testing.T) {
	config := `{
		"servers": {
			"env-test": {
				"type": "stdio",
				"command": "/usr/bin/env-test",
				"env": {
					"PATH": "/custom/path:/usr/bin",
					"API_KEY": "secret123",
					"DEBUG": "true",
					"EMPTY_VAR": ""
				}
			}
		}
	}`

	tmpfile, cleanup := createTempConfig(t, config)
	defer cleanup()

	clients, err := Clients(tmpfile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	c, ok := clients["env-test"]
	if !ok {
		t.Fatal("expected 'env-test' client")
	}

	stdio, ok := c.(client.Stdio)
	if !ok {
		t.Fatal("expected Stdio client")
	}

	if len(stdio.Env) != 4 {
		t.Fatalf("expected 4 env vars, got %d", len(stdio.Env))
	}

	expectedEnv := map[string]bool{
		"PATH=/custom/path:/usr/bin": false,
		"API_KEY=secret123":          false,
		"DEBUG=true":                 false,
		"EMPTY_VAR=":                 false,
	}
	for _, env := range stdio.Env {
		if _, ok := expectedEnv[env]; ok {
			expectedEnv[env] = true
		}
	}
	for env, found := range expectedEnv {
		if !found {
			t.Errorf("expected env var '%s' not found", env)
		}
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
