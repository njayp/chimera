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

	if len(cfg.MCPServers) == 0 {
		log.Fatal("No servers configured")
	}

	// Create HTTP handler that creates a new aggregating server per session
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		agg := aggregator.New()

		// Connect to all stdio servers for this HTTP session
		if err := agg.ConnectToStdioServers(req.Context(), cfg.MCPServers); err != nil {
			log.Printf("Failed to connect to stdio servers: %v", err)
			return nil
		}

		return agg.MCPServer()
	}, nil)

	// Start HTTP server
	log.Printf("Starting aggregating MCP HTTP server on %s", cfg.Address)
	log.Printf("Aggregating %d server(s)", len(cfg.MCPServers))
	for name, srv := range cfg.MCPServers {
		log.Printf("  - %s: %s", name, srv.Command)
	}

	if err := http.ListenAndServe(cfg.Address, handler); err != nil {
		log.Fatal(err)
	}
}
