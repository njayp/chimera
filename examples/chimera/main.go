// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/njayp/chimera/config/watcher"
	"github.com/njayp/chimera/proxy"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	path := flag.String("config", ".vscode/mcp.json", "path to configuration file")
	flag.Parse()

	ctx := context.Background()
	watcher, err := watcher.NewVSCodeWatcher(ctx, *path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	handler := proxy.Handler(watcher.Clients)

	// Start HTTP server
	addr := ":8080"
	log.Printf("Starting aggregating MCP HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
