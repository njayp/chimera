// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/njayp/chimera/config/watcher"
	"github.com/njayp/chimera/proxy"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// get port and path from env vars
	port, exists := os.LookupEnv("PORT")
	if !exists {
		port = "8080"
	}

	path, exists := os.LookupEnv("CONFIG_PATH")
	if !exists {
		path = ".vscode/mcp.json"
	}

	ctx := context.Background()
	watcher, err := watcher.NewVSCodeWatcher(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Start HTTP server
	addr := ":" + port
	log.Printf("Starting reverse-proxy MCP HTTP server on address %q", addr)
	handler := proxy.Handler(watcher.Clients)
	return http.ListenAndServe(addr, handler)
}
