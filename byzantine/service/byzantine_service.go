package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/node"

	"github.com/ethereum/go-ethereum/byzantine/adapter"
	byzantineapi "github.com/ethereum/go-ethereum/byzantine/api"
	"github.com/ethereum/go-ethereum/byzantine/manager"
	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

// ByzantineService implements the ByzantineService interface
type ByzantineService struct {
	mu             sync.RWMutex
	config         *types.ByzantineConfig
	configLoader   *ConfigLoader
	status         types.ServiceStatus
	attackManager  types.AttackManager
	eventPublisher types.EventPublisher
	hookAdapter    types.HookAdapter
	attackRegistry *registry.AttackRegistry
	paramRegistry  *registry.ParameterParserRegistry

	ctx    context.Context
	cancel context.CancelFunc

	consensusHook types.ConsensusHook
}

var _ node.Lifecycle = (*ByzantineService)(nil)

// NewByzantineService creates a new Byzantine service
func NewByzantineService(config *types.ByzantineConfig) (*ByzantineService, error) {
	// Create components
	attackRegistry := registry.DefaultRegistry
	attackManager := manager.NewAttackManager(attackRegistry)
	eventPublisher := adapter.NewEventObserver()
	configLoader := NewConfigLoader()

	service := &ByzantineService{
		config:         config,
		attackManager:  attackManager,
		eventPublisher: eventPublisher,
		attackRegistry: attackRegistry,
		configLoader:   configLoader,
	}

	// Create a hook adapter
	service.hookAdapter = adapter.NewHookAdapter(service, eventPublisher)
	service.consensusHook = NewConsensusHook(service.attackManager, service.eventPublisher)

	return service, nil
}

// Start starts the service
func (s *ByzantineService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status.Running {
		return fmt.Errorf("service already running")
	}

	// Create context
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Update status
	now := time.Now()
	s.status = types.ServiceStatus{
		Running:   true,
		StartedAt: &now,
	}

	// Subscribe to events after context is created
	s.subscribeToEvents()

	// Load attacks from config
	if err := s.loadAttacksFromConfig(); err != nil {
		// Clean up on error
		s.cancel()
		s.ctx = nil
		s.cancel = nil
		s.status.Running = false
		s.status.StartedAt = nil
		return fmt.Errorf("failed to load attacks from config: %w", err)
	}

	return nil
}

// Stop stops the service
func (s *ByzantineService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.status.Running {
		return fmt.Errorf("service not running")
	}

	// Cancel context
	if s.cancel != nil {
		s.cancel()
		s.ctx = nil
		s.cancel = nil
	}

	// Update status
	s.status.Running = false
	s.status.StartedAt = nil

	return nil
}

// GetStatus returns the service status
func (s *ByzantineService) GetStatus() types.ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := s.status

	// Get attack counts
	attackList := s.attackManager.ListAttacks()
	for _, attack := range attackList {
		switch attack.GetConfig().Status {
		case types.AttackStatusActive, types.AttackStatusPending:
			status.ActiveAttacks++
		case types.AttackStatusExecuted:
			status.ExecutedAttacks++
		case types.AttackStatusFailed:
			status.FailedAttacks++
		}
	}

	status.StoredMessages = 0

	return status
}

// Configure configures the service
func (s *ByzantineService) Configure(config *types.ByzantineConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// Reload attacks if needed
	if s.status.Running {
		return s.loadAttacksFromConfig()
	}

	return nil
}

// RegisterAttack registers a new attack
func (s *ByzantineService) RegisterAttack(config types.AttackConfig) (string, error) {
	// Set initial status
	config.Status = types.AttackStatusPending
	config.CreatedAt = time.Now()

	// Create attack instance
	attack, err := s.attackRegistry.CreateAttack(config)
	if err != nil {
		return "", fmt.Errorf("failed to create attack: %w", err)
	}

	// Register with manager
	if err = s.attackManager.RegisterAttack(attack); err != nil {
		log.Error("Failed to register attack", "name", config.Name, "error", err)
		return "", fmt.Errorf("failed to register attack: %w", err)
	}

	uid := attack.GetUID()
	log.Debug("[BYZ] Attack registered successfully",
		"name", config.Name,
		"uid", uid,
		"type", config.Type,
		"enabled", config.Enabled,
		"code", config.Parameters["code"],
		"seq_start", config.SequenceStart,
		"seq_end", config.SequenceEnd,
		"max_executions", config.MaxExecutionCount)

	return uid, nil
}

// CancelAttack cancels an attack
func (s *ByzantineService) CancelAttack(uid string) error {
	return s.attackManager.UnregisterAttack(uid)
}

// ListAttacks lists all attacks
func (s *ByzantineService) ListAttacks() []types.AttackConfig {
	attackList := s.attackManager.ListAttacks()
	configs := make([]types.AttackConfig, len(attackList))

	for i, attack := range attackList {
		configs[i] = attack.GetConfig()
	}

	return configs
}

// GetHookAdapter returns the hook adapter
func (s *ByzantineService) GetHookAdapter() types.HookAdapter {
	return s.hookAdapter
}

// subscribeToEvents subscribes to consensus events
func (s *ByzantineService) subscribeToEvents() {
	// Skip if context not initialized
	if s.ctx == nil {
		log.Warn("Cannot subscribe to events: context not initialized")
		return
	}

	// Subscribe to all event types
	eventTypes := []types.EventType{
		types.EventTypeMessageReceived,
		types.EventTypeMessageSent,
		types.EventTypeRoundChange,
		types.EventTypeProposalCreated,
		types.EventTypeBlockCommitted,
	}

	for _, eventType := range eventTypes {
		err := s.eventPublisher.Subscribe(eventType, func(event types.Event) error {
			// Process event through attack manager
			ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
			defer cancel()

			err := s.attackManager.ProcessEvent(ctx, event)

			return err
		})
		if err != nil {
			log.Warn("Failed to subscribe to events: ", err)
		}
	}
}

// loadAttacksFromConfig loads attacks from configuration
func (s *ByzantineService) loadAttacksFromConfig() error {
	log.Debug("[BYZ] Loading byzantine configuration", "cnt", len(s.config.Attacks))
	for _, attackConfig := range s.config.Attacks {
		if _, err := s.RegisterAttack(attackConfig); err != nil {
			return fmt.Errorf("failed to register attack %s: %w", attackConfig.Name, err)
		}
	}
	return nil
}

// APIs returns the collection of RPC services the byzantine service offers
func (s *ByzantineService) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "byzantine",
			Version:   "1.0",
			Service:   byzantineapi.NewPublicByzantineAPI(s),
			Public:    true,
		},
	}
}

// GetAttackManager returns the attack manager (interface for API)
func (s *ByzantineService) GetAttackManager() types.AttackManager {
	return s.attackManager
}

// GetConsensusHook returns the consensus hook for integration
func (s *ByzantineService) GetConsensusHook() types.ConsensusHook {
	return s.consensusHook
}
