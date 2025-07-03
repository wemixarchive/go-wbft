package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/service"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
)

// DefaultByzantineConfig returns default Byzantine configuration
func DefaultByzantineConfig() *types.ByzantineConfig {
	return &types.ByzantineConfig{
		Enabled: true,
		Attacks: []types.AttackConfig{},
		StorageConfig: types.StorageConfig{
			MessageRetention: 24 * time.Hour,
			HistoryRetention: 7 * 24 * time.Hour,
			MaxStorageSize:   1 << 30, // 1GB
			PruneInterval:    time.Hour,
		},
		Monitoring: types.MonitoringConfig{
			Enabled:         false,
			MetricsPort:     9090,
			LogLevel:        "info",
			AlertThresholds: make(map[string]int),
		},
	}
}

// LoadByzantineConfig loads Byzantine configuration from command line and config file
func LoadByzantineConfig(ctx *cli.Context, nodeConfig *node.Config) (*types.ByzantineConfig, error) {
	config := DefaultByzantineConfig()

	if ctx.IsSet(ByzantineEnabledFlag.Name) {
		config.Enabled = ctx.Bool(ByzantineEnabledFlag.Name)
		return config, fmt.Errorf("[BYZ] Byzantine is disabled")
	}

	// First load from config file if specified
	if ctx.IsSet(ByzantineConfigFileFlag.Name) {
		configFile := ctx.String(ByzantineConfigFileFlag.Name)
		if configFile != "" {
			configPath := resolveConfigPath(configFile, nodeConfig.DataDir)
			configLoader := service.NewConfigLoader()
			parsedConfig, err := configLoader.LoadConfig(configPath)
			if err != nil {
				return nil, fmt.Errorf("[BYZ] failed to load Byzantine config from '%s': %v\n"+
					"Hint: For relative paths, files are searched in:\n"+
					"  1. Current working directory\n"+
					"  2. Geth data directory (%s)", configPath, err, nodeConfig.DataDir)
			}
			return parsedConfig, nil
		}
	}

	// Apply defaults for zero values
	if config.StorageConfig.MessageRetention == 0 {
		config.StorageConfig.MessageRetention = 24 * time.Hour
	}
	if config.StorageConfig.HistoryRetention == 0 {
		config.StorageConfig.HistoryRetention = 7 * 24 * time.Hour
	}
	if config.StorageConfig.PruneInterval == 0 {
		config.StorageConfig.PruneInterval = time.Hour
	}

	return config, nil
}

// resolveConfigPath resolves the config file path
func resolveConfigPath(configFile string, dataDir string) string {
	if filepath.IsAbs(configFile) {
		return configFile
	}

	// Check current working directory first
	if cwd, err := os.Getwd(); err == nil {
		cwdPath := filepath.Join(cwd, configFile)
		if _, err = os.Stat(cwdPath); err == nil {
			log.Info("Using Byzantine config from current directory", "path", cwdPath)
			return cwdPath
		}
	}

	// Fall back to data directory
	dataPath := filepath.Join(dataDir, configFile)
	log.Info("Using Byzantine config from data directory", "path", dataPath)
	return dataPath
}

// loadConfigFromFile loads configuration from JSON file with custom parsing
func loadConfigFromFile(path string, config *types.ByzantineConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Use intermediate struct for parsing durations
	var rawConfig struct {
		Enabled bool                 `json:"enabled"`
		Attacks []types.AttackConfig `json:"attacks,omitempty"`
		Storage struct {
			MessageRetention string `json:"message_retention"`
			HistoryRetention string `json:"history_retention"`
			MaxStorageSize   int64  `json:"max_storage_size"`
			PruneInterval    string `json:"prune_interval"`
		} `json:"storage"`
		Monitoring types.MonitoringConfig `json:"monitoring"`
	}

	if err = json.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Copy simple fields
	config.Enabled = rawConfig.Enabled
	config.Attacks = rawConfig.Attacks
	config.Monitoring = rawConfig.Monitoring

	// Parse duration strings
	if rawConfig.Storage.MessageRetention != "" {
		duration, err := time.ParseDuration(rawConfig.Storage.MessageRetention)
		if err != nil {
			return fmt.Errorf("invalid message_retention duration: %w", err)
		}
		config.StorageConfig.MessageRetention = duration
	}

	if rawConfig.Storage.HistoryRetention != "" {
		duration, err := time.ParseDuration(rawConfig.Storage.HistoryRetention)
		if err != nil {
			return fmt.Errorf("invalid history_retention duration: %w", err)
		}
		config.StorageConfig.HistoryRetention = duration
	}

	if rawConfig.Storage.PruneInterval != "" {
		duration, err := time.ParseDuration(rawConfig.Storage.PruneInterval)
		if err != nil {
			return fmt.Errorf("invalid prune_interval duration: %w", err)
		}
		config.StorageConfig.PruneInterval = duration
	}

	config.StorageConfig.MaxStorageSize = rawConfig.Storage.MaxStorageSize

	return nil
}
