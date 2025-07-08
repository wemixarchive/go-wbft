package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	}
}

// LoadByzantineConfig loads Byzantine configuration from the command line and config file
func LoadByzantineConfig(ctx *cli.Context, nodeConfig *node.Config) (*types.ByzantineConfig, error) {
	config := DefaultByzantineConfig()

	if ctx.IsSet(ByzantineEnabledFlag.Name) {
		config.Enabled = ctx.Bool(ByzantineEnabledFlag.Name)
		return config, fmt.Errorf("[BYZ] Byzantine is disabled")
	}

	// First load from a config file if specified
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

	return config, nil
}

// resolveConfigPath resolves the config file path
func resolveConfigPath(configFile string, dataDir string) string {
	if filepath.IsAbs(configFile) {
		return configFile
	}

	// Check the current working directory first
	if cwd, err := os.Getwd(); err == nil {
		cwdPath := filepath.Join(cwd, configFile)
		if _, err = os.Stat(cwdPath); err == nil {
			log.Info("Using Byzantine config from current directory", "path", cwdPath)
			return cwdPath
		}
	}

	// Fall back to the data directory
	dataPath := filepath.Join(dataDir, configFile)
	log.Info("Using Byzantine config from data directory", "path", dataPath)
	return dataPath
}
