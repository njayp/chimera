// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/njayp/chimera/config"
	"github.com/njayp/chimera/proxy"
)

func main() {
	panic(run())
}

func run() error {
	configPath := flag.String("config", ".vscode/mcp.json", "path to configuration file")
	flag.Parse()

	config, err := config.VSCode(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	handler := proxy.Handler(config)

	// Start HTTP server
	addr := ":8080"
	log.Printf("Starting aggregating MCP HTTP server on %s", addr)
	return http.ListenAndServe(addr, handler)
}
