package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

// ByzantineConfig holds all configuration for Byzantine module
type ByzantineConfig struct {
	Enabled    bool                 `json:"enabled"`
	ConfigFile string               `json:"configFile,omitempty"`
	LogLevel   string               `json:"logLevel"`
	Attacks    []types.AttackConfig `json:"attacks,omitempty"`
}

// Validate validates the Byzantine configuration
func (c *ByzantineConfig) Validate() error {
	// Validate log level (enhanced validation)
	if err := c.validateLogLevel(); err != nil {
		return err
	}

	// Validate attack configurations
	for i, attack := range c.Attacks {
		if attack.Name == "" {
			return errors.New("attack name cannot be empty")
		}
		if attack.Type == "" {
			return errors.New("attack type cannot be empty")
		}

		// Validate attack type
		validTypes := []string{"silent", "tampered", "fake", "omit", "roleSpoofed", "replay", "flood"}
		isValidType := false
		for _, validType := range validTypes {
			if attack.ConvertTypeToString() == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return errors.New("invalid attack type: " + attack.ConvertTypeToString())
		}

		// Warning for missing sequence/round (not an error)
		if attack.Sequence == 0 && attack.Round == 0 {
			log.Warn("Attack configured without sequence or round", "index", i, "name", attack.Name)
		}
	}

	return nil
}

// validateLogLevel performs comprehensive log level validation
func (c *ByzantineConfig) validateLogLevel() error {
	// Check if log level is empty
	if c.LogLevel == "" {
		return errors.New("log level cannot be empty")
	}

	// Convert to lowercase for case-insensitive comparison
	logLevel := strings.ToLower(strings.TrimSpace(c.LogLevel))

	// Validate against supported log levels
	validLevels := map[string]bool{
		"trace": true,
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[logLevel] {
		return errors.New("invalid log level '" + c.LogLevel + "', supported levels: trace, debug, info, warn, error, crit")
	}

	// Normalize the log level to lowercase (optional: for consistency)
	c.LogLevel = logLevel

	// Verify that the log level can be converted to slog.Level
	// This ensures consistency between validation and GetLogLevel()
	level := c.GetLogLevel()
	if level == slog.LevelInfo && logLevel != "info" {
		// If GetLogLevel() returns default (info) but input wasn't "info",
		// it might indicate a conversion issue
		log.Warn("Log level conversion might have issues", "input", c.LogLevel, "converted", level)
	}

	return nil
}

// GetLogLevel returns the configured log level as slog.Level
func (c *ByzantineConfig) GetLogLevel() slog.Level {
	// Normalize log level for consistent comparison
	logLevel := strings.ToLower(strings.TrimSpace(c.LogLevel))

	switch logLevel {
	case "trace":
		return log.LevelTrace
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		// Return default level for any invalid input
		return log.LevelInfo
	}
}

// DefaultByzantineConfig returns default Byzantine configuration
func DefaultByzantineConfig() *ByzantineConfig {
	return &ByzantineConfig{
		Enabled:  true,
		LogLevel: "info",
	}
}

// LoadByzantineConfig loads Byzantine configuration from command line and config file
func LoadByzantineConfig(ctx *cli.Context, nodeConfig *node.Config) (*ByzantineConfig, error) {
	config := DefaultByzantineConfig()

	// First load from config file if specified (this will override defaults)
	if ctx.IsSet(ByzantineConfigFileFlag.Name) {
		configFile := ctx.String(ByzantineConfigFileFlag.Name)
		if configFile != "" {
			configPath := configFile
			if !filepath.IsAbs(configPath) {
				// Improved relative path handling
				// First, check by current working directory
				if cwd, err := os.Getwd(); err == nil {
					cwdPath := filepath.Join(cwd, configFile)
					if _, err := os.Stat(cwdPath); err == nil {
						configPath = cwdPath
						log.Info("Using Byzantine config from current directory", "path", configPath)
					} else {
						// 2. If not in current directory, check by DataDir
						configPath = filepath.Join(nodeConfig.DataDir, configFile)
						log.Info("Using Byzantine config from data directory", "path", configPath)
					}
				} else {
					// Use DataDir the old-fashioned way when os.Getwd() fails
					configPath = filepath.Join(nodeConfig.DataDir, configFile)
				}
			}

			if err := loadConfigFromFile(configPath, config); err != nil {
				return nil, fmt.Errorf("failed to load Byzantine config from '%s': %v\n"+
					"Hint: For relative paths, files are searched in:\n"+
					"  1. Current working directory\n"+
					"  2. Geth data directory (%s)", configPath, err, nodeConfig.DataDir)
			}
			config.ConfigFile = configPath
		}
	}

	// Then override with command line flags (CLI takes precedence over file)
	if ctx.IsSet(ByzantineEnabledFlag.Name) {
		config.Enabled = ctx.Bool(ByzantineEnabledFlag.Name)
	}

	if ctx.IsSet(ByzantineLogLevelFlag.Name) {
		config.LogLevel = ctx.String(ByzantineLogLevelFlag.Name)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadConfigFromFile loads configuration from JSON file
func loadConfigFromFile(path string, config *ByzantineConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse JSON
	var fileConfig ByzantineConfig
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return err
	}

	// Merge all non-zero/non-default values from file
	// Note: CLI flags will override these values later in LoadByzantineConfig

	if fileConfig.Enabled {
		config.Enabled = fileConfig.Enabled
	}

	if fileConfig.LogLevel != "" {
		config.LogLevel = fileConfig.LogLevel
	}

	// APIEnabled is a bool, so we need to check if it was explicitly set in the file
	// We can determine this by checking if the field exists in the raw JSON
	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return err
	}

	if len(fileConfig.Attacks) > 0 {
		config.Attacks = fileConfig.Attacks
	}

	return nil
}
