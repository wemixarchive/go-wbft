package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// ConfigLoader loads Byzantine configuration
type ConfigLoader struct{}

// NewConfigLoader creates a new config loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

// LoadFromFile loads configuration from a JSON file
func (c *ConfigLoader) LoadFromFile(filename string) (*types.ByzantineConfig, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.ByzantineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := c.validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadFromJSON loads configuration from JSON data
func (c *ConfigLoader) LoadFromJSON(data []byte) (*types.ByzantineConfig, error) {
	var config types.ByzantineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := c.validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validate validates the configuration
func (c *ConfigLoader) validate(config *types.ByzantineConfig) error {
	if !config.Enabled {
		return nil // Skip validation if not enabled
	}

	// Validate storage config
	if config.StorageConfig.MessageRetention <= 0 {
		config.StorageConfig.MessageRetention = 24 * time.Hour
	}

	if config.StorageConfig.HistoryRetention <= 0 {
		config.StorageConfig.HistoryRetention = 7 * 24 * time.Hour
	}

	if config.StorageConfig.PruneInterval <= 0 {
		config.StorageConfig.PruneInterval = time.Hour
	}

	// Validate attacks
	for i, attack := range config.Attacks {
		if attack.Type == "" {
			return fmt.Errorf("attack[%d]: type is required", i)
		}

		if attack.Name == "" {
			attack.Name = fmt.Sprintf("%s_%d", attack.Type, i)
		}
	}

	return nil
}

// SaveToFile saves configuration to a JSON file
func (c *ConfigLoader) SaveToFile(config *types.ByzantineConfig, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
