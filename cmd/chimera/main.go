// Package main provides the entry point for the MCP aggregating HTTP server.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/njayp/chimera/pkg/proxy"
	"gopkg.in/yaml.v3"
)

func main() {
	configPath := flag.String("config", "config.json", "path to configuration file")
	flag.Parse()

	config, err := loadConfig(*configPath)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	panic((proxy.Run(config, ":8080")))
}

// loadConfig loads configuration from a file. Supports JSON and YAML formats
// based on file extension.
func loadConfig(path string) (proxy.Servers, error) {
	var config proxy.Servers

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	// Determine format by file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		// Unsupported extension
		return config, fmt.Errorf("extension not supported: %s", ext)
	}

	return config, nil
}
