// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/njayp/chimera/pkg/aggregator"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	// Load configuration from file
	cfg, err := aggregator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create HTTP handler that creates a new aggregating server per session
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		agg := aggregator.New()

		// Connect to all stdio servers
		if err := agg.ConnectToStdioServers(req.Context(), cfg.StdioServers); err != nil {
			log.Printf("Failed to connect to stdio servers: %v", err)
			return nil
		}

		// Connect to all HTTP servers
		if err := agg.ConnectToHTTPServers(req.Context(), cfg.HTTPServers); err != nil {
			log.Printf("Failed to connect to HTTP servers: %v", err)
			return nil
		}

		return agg.MCPServer()
	}, nil)

	// Start HTTP server
	log.Printf("Starting aggregating MCP HTTP server on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, handler); err != nil {
		log.Fatal(err)
	}
}
