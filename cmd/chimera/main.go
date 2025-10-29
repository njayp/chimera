// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"flag"
	"fmt"

	"github.com/njayp/chimera/pkg/config"
	"github.com/njayp/chimera/pkg/proxy"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	config, err := config.VSCode(*configPath)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	panic((proxy.Run(config, ":8080")))
}
