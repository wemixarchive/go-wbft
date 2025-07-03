package byzantine

import (
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/wbft/backend"

	"github.com/ethereum/go-ethereum/byzantine/cmd"
	"github.com/ethereum/go-ethereum/byzantine/service"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
)

// Register registers Byzantine service using the provided config
func Register(ctx *cli.Context, stack *node.Node, backend ethapi.Backend, eth *eth.Ethereum) error {
	// Load Byzantine configuration
	config, err := cmd.LoadByzantineConfig(ctx, stack.Config())
	if err != nil {
		return fmt.Errorf("byzantine: failed to load configuration: %w", err)
	}

	// Skip registration if Byzantine is disabled
	if !config.Enabled {
		log.Warn("[BYZ] module disabled")
		return nil
	}

	log.Info("[BYZ] Registering Byzantine module",
		"enabled", config.Enabled,
		"attacks count", len(config.Attacks),
	)

	log.Info("=== Attack configurations ===")
	for i, attack := range config.Attacks {
		log.Info(fmt.Sprintf("→ [%d] uid=%s name=%s type=%s enabled=%v sequence=%d-%d round=%d parameters=%v",
			i,
			attack.UID,
			attack.Name,
			attack.Type,
			attack.Enabled,
			attack.SequenceStart,
			attack.SequenceEnd,
			attack.Round,
			attack.Parameters,
		))
	}

	log.Info("=== Storage configuration ===")
	log.Info(fmt.Sprintf("→ retention=%s history=%s max size=%d prune interval=%s",
		config.StorageConfig.MessageRetention,
		config.StorageConfig.HistoryRetention,
		config.StorageConfig.MaxStorageSize,
		config.StorageConfig.PruneInterval,
	))

	log.Info("=== Monitoring configuration ===")
	log.Info(fmt.Sprintf("→ enabled=%v port=%d log level=%s alert thresholds=%v",
		config.Monitoring.Enabled,
		config.Monitoring.MetricsPort,
		config.Monitoring.LogLevel,
		config.Monitoring.AlertThresholds,
	))
	// Create Byzantine service
	byzantineService, err := service.NewByzantineService(config)
	if err != nil {
		return fmt.Errorf("byzantine: failed to create service: %w", err)
	}

	// Register service lifecycle
	stack.RegisterLifecycle(byzantineService)

	// Register APIs
	apis := byzantineService.APIs()
	if len(apis) > 0 {
		stack.RegisterAPIs(apis)
		log.Info("[BYZ] Byzantine APIs registered", "count", len(apis))
	}

	log.Info("[BYZ] Byzantine module registered successfully")

	if eth != nil && eth.Engine() != nil {
		if err := IntegrateByzantineWithConsensus(byzantineService, eth.Engine()); err != nil {
			log.Warn("Failed to integrate Byzantine with consensus", "err", err)
		}
	}
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

// IntegrateByzantineWithConsensus connects Byzantine module with consensus engine
func IntegrateByzantineWithConsensus(service types.ByzantineService, consensusEngine consensus.Engine) error {
	// Type assertion to WBFT backend
	wbftBackend, ok := consensusEngine.(*backend.Backend)
	if ok {
		hook := service.GetConsensusHook()
		wbftBackend.SetByzantineHook(hook)

		log.Info("[BYZ] module integrated with consensus",
			"total_attacks", len(service.ListAttacks()),
			"active_attacks", service.GetStatus().ActiveAttacks)
	} else {
		log.Debug("Consensus is not WBFT, skipping Byzantine integration")
	}

	return nil
}
