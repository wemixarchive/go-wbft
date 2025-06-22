package byzantine

import (
	"fmt"
	"github.com/ethereum/go-ethereum/node"
	"sync"

	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/byzantine/cmd"
)

// ByzantineService implements the Byzantine service
type ByzantineService struct {
	config   *cmd.ByzantineConfig
	backend  ethapi.Backend
	ethereum *eth.Ethereum
	//manager *manager.ByzantineManager

	// Components
	attackManager  AttackManager
	eventCollector EventCollector

	// Service state
	running   bool
	runningMu sync.RWMutex
	stopCh    chan struct{}
}

// Ensure ByzantineService implements node.Lifecycle
var _ node.Lifecycle = (*ByzantineService)(nil)

// NewByzantineService creates a new Byzantine service
func NewByzantineService(config *cmd.ByzantineConfig, backend ethapi.Backend, ethereum *eth.Ethereum) (*ByzantineService, error) {
	if config == nil {
		return nil, fmt.Errorf("byzantine: configuration cannot be nil")
	}

	if backend == nil {
		return nil, fmt.Errorf("byzantine: backend cannot be nil")
	}

	service := &ByzantineService{
		config:   config,
		backend:  backend,
		ethereum: ethereum,
		stopCh:   make(chan struct{}),
	}

	if err := service.initialize(); err != nil {
		return nil, fmt.Errorf("byzantine: failed to initialize service: %w", err)
	}

	return service, nil
}

// initialize sets up service components
func (s *ByzantineService) initialize() error {
	// Create event collector
	s.eventCollector = NewEventCollector()

	// Create attack builder
	builder := NewAttackBuilder()

	// Create attack manager
	s.attackManager = NewAttackManager(s.eventCollector, builder)

	// Load attacks from configuration
	for _, attackCfg := range s.config.Attacks {
		if !attackCfg.Enabled {
			log.Debug("Skipping disabled attack", "name", attackCfg.Name)
			continue
		}

		if err := s.attackManager.LoadAttack(attackCfg); err != nil {
			log.Error("Failed to load attack from config",
				"name", attackCfg.Name,
				"type", attackCfg.ConvertTypeToString(),
				"error", err)
			// Continue loading other attacks
		} else {
			log.Info("Loaded attack from config",
				"name", attackCfg.Name,
				"type", attackCfg.ConvertTypeToString(),
				"sequence", attackCfg.Sequence,
				"round", attackCfg.Round)
		}
	}

	return nil
}

// Start implements node.Lifecycle
func (s *ByzantineService) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if s.running {
		return nil
	}

	log.Info("Starting Byzantine service", "logLevel", s.config.LogLevel)
	// Set log level
	//log.Root().SetHandler(log.LvlFilterHandler(s.config.GetLogLevel(), log.Root().GetHandler()))

	// Start event collector
	if err := s.eventCollector.Start(); err != nil {
		return fmt.Errorf("failed to start event collector: %w", err)
	}

	// Start attack manager
	if err := s.attackManager.Start(); err != nil {
		return fmt.Errorf("failed to start attack manager: %w", err)
	}

	// Register hooks with consensus engine if available
	if s.ethereum != nil {
		s.registerConsensusHooks()
	}

	s.running = true
	log.Info("Byzantine service started", "attacks", len(s.config.Attacks))
	return nil
}

// Stop implements node.Lifecycle
func (s *ByzantineService) Stop() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		return nil
	}

	log.Info("Stopping Byzantine service")

	// Signal stop
	close(s.stopCh)

	// Stop components
	if s.attackManager != nil {
		s.attackManager.Stop()
	}

	if s.eventCollector != nil {
		s.eventCollector.Stop()
	}

	// Unregister hooks
	if s.ethereum != nil {
		s.unregisterConsensusHooks()
	}

	s.running = false
	log.Info("Byzantine service stopped")
	return nil
}

// APIs returns the RPC API descriptors the Byzantine service offers
func (s *ByzantineService) APIs() []rpc.API {
	log.Info("[Byzantine] Byzantine APIs called")
	apis := []rpc.API{}
	apis = append(apis, rpc.API{
		Namespace:     "byzantine",
		Version:       "1.0",
		Service:       NewByzantineAPI(s),
		Public:        true,
		Authenticated: false,
	})
	return apis
}

// IsRunning returns whether the service is running
func (s *ByzantineService) IsRunning() bool {
	s.runningMu.RLock()
	defer s.runningMu.RUnlock()
	return s.running
}

// registerConsensusHooks registers Byzantine hooks with the consensus engine
func (s *ByzantineService) registerConsensusHooks() {
	// This is where you would register hooks with QBFT
	// Implementation depends on QBFT hook interface
	log.Debug("Registering consensus hooks")
}

// unregisterConsensusHooks removes Byzantine hooks from the consensus engine
func (s *ByzantineService) unregisterConsensusHooks() {
	log.Debug("Unregistering consensus hooks")
}
