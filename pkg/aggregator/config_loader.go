package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds the complete configuration for the aggregating server.
type ServerConfig struct {
	Address    string                 `json:"address" yaml:"address"`       // HTTP server address (e.g., ":8080")
	MCPServers map[string]StdioConfig `json:"mcpServers" yaml:"mcpServers"` // List of stdio servers to aggregate
}

// LoadConfig loads configuration from a file. Supports JSON and YAML formats
// based on file extension.
func LoadConfig(path string) (*ServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ServerConfig

	// Determine format by file extension
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	default:
		// Unsupported extension
		return nil, fmt.Errorf("extension not supported: %s", ext)
	}

	// Set default address if not specified
	if cfg.Address == "" {
		cfg.Address = ":8080"
	}

	return &cfg, nil
}
