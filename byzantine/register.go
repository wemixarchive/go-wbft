package byzantine

import (
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/cmd"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
)

// Register registers Byzantine service using the provided config
// This is called from cmd/geth/config.go after LoadByzantineConfig
func Register(ctx *cli.Context, stack *node.Node, backend ethapi.Backend, eth *eth.Ethereum) error {
	// Load Byzantine configuration
	config, err := cmd.LoadByzantineConfig(ctx, stack.Config())
	if err != nil {
		return fmt.Errorf("byzantine: failed to load configuration: %w", err)
	}
	log.Info("===[Byzantine]=== Byzantine config", "config", config)

	// Skip registration if Byzantine is disabled
	if !config.Enabled {
		log.Info("Byzantine module disabled")
		return nil
	}

	log.Info("Registering Byzantine module",
		"enabled", config.Enabled,
		"configFile", config.ConfigFile,
		"logLevel", config.LogLevel,
		"attacks", len(config.Attacks))

	// Create and register Byzantine service
	service, err := NewByzantineService(config, backend, eth)
	if err != nil {
		return fmt.Errorf("byzantine: failed to create service: %w", err)
	}

	// Register service lifecycle
	stack.RegisterLifecycle(service)

	// Register APIs
	apis := service.APIs()
	if len(apis) > 0 {
		stack.RegisterAPIs(apis)
		log.Info("Byzantine APIs registered", "count", len(apis))
	}

	log.Info("Byzantine module registered successfully")
	return nil
}

// RegisterFlags adds Byzantine-specific flags to the command
func RegisterFlags(app *cli.App) {
	byzantineCmdFlags := cmd.ByzantineCommandFlags()
	app.Flags = append(app.Flags, byzantineCmdFlags...)
}

// RegisterByzantineCommands registers Byzantine commands to the app
func RegisterByzantineCommands(app *cli.App) {
	byzantineCmd := cmd.ByzantineCommand()
	app.Commands = append(app.Commands, byzantineCmd)
}
