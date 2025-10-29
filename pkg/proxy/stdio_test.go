package proxy

import (
	"context"
	"testing"
	"time"
)

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
