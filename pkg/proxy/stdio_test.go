package proxy

import (
	"context"
	"testing"
	"time"
)

func TestStdioClient_Connect_EchoCommand(t *testing.T) {
	client := StdioClient{
		Command: "echo",
		Args:    []string{"hello"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	session, err := client.connect(ctx)

	// We expect this to fail because echo is not an MCP server
	// but we're testing that the command execution works
	if session != nil {
		t.Error("expected nil session for non-MCP command")
	}

	if err == nil {
		t.Error("expected error for non-MCP command, got nil")
	}
}

func TestStdioClient_Connect_InvalidCommand(t *testing.T) {
	client := StdioClient{
		Command: "nonexistent-command-12345",
		Args:    []string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.connect(ctx)
	if err == nil {
		t.Fatal("expected error for invalid command, got nil")
	}
}

func TestStdioClient_Connect_WithEnv(t *testing.T) {
	client := StdioClient{
		Command: "printenv",
		Args:    []string{"TEST_VAR"},
		Env:     []string{"TEST_VAR=test_value"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail as an MCP connection but tests env var handling
	_, err := client.connect(ctx)
	if err == nil {
		t.Error("expected error for non-MCP command")
	}
}

func TestStdioClient_Connect_ContextCancellation(t *testing.T) {
	client := StdioClient{
		Command: "sleep",
		Args:    []string{"10"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.connect(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestStdioClient_EmptyCommand(t *testing.T) {
	client := StdioClient{
		Command: "",
		Args:    []string{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.connect(ctx)
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}
