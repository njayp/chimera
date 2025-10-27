package aggregator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the complete configuration for the aggregating server.
type Config struct {
	Address      string                 `json:"address" yaml:"address"` // HTTP server address (e.g., ":8080")
	StdioServers map[string]StdioConfig `json:"stdioServers" yaml:"stdioServers"`
	HTTPServers  map[string]HTTPConfig  `json:"httpServers" yaml:"httpServers"`
}

// LoadConfig loads configuration from a file. Supports JSON and YAML formats
// based on file extension.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config

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
