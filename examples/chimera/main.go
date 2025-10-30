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
	port := flag.String("port", "8080", "HTTP server port")
	flag.Parse()

	ctx := context.Background()
	watcher, err := watcher.NewVSCodeWatcher(ctx, *path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Start HTTP server
	addr := ":" + *port
	log.Printf("Starting reverse-proxy MCP HTTP server on address %q", addr)
	handler := proxy.Handler(watcher.Clients)
	return http.ListenAndServe(addr, handler)
}
