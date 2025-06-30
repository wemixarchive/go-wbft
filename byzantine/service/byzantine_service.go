package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/node"

	"github.com/ethereum/go-ethereum/byzantine/adapter"
	byzantineapi "github.com/ethereum/go-ethereum/byzantine/api"
	"github.com/ethereum/go-ethereum/byzantine/attacks"
	"github.com/ethereum/go-ethereum/byzantine/manager"
	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/storage"
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
	messageStorage types.MessageStorage
	historyStorage types.HistoryStorage
	eventPublisher types.EventPublisher
	hookAdapter    types.HookAdapter
	attackRegistry *registry.AttackRegistry
	paramRegistry  *registry.ParameterParserRegistry

	ctx    context.Context
	cancel context.CancelFunc

	// Metrics
	metrics atomic.Value // types.Metrics

	consensusHook types.ConsensusHook
}

var _ node.Lifecycle = (*ByzantineService)(nil)

// storageProviderImpl implements StorageProvider
type storageProviderImpl struct {
	messageStorage types.MessageStorage
}

func (s *storageProviderImpl) GetMessageStorage() types.MessageStorage {
	return s.messageStorage
}

// NewByzantineService creates a new Byzantine service
func NewByzantineService(config *types.ByzantineConfig) (*ByzantineService, error) {
	log.Warn("storage config : ", config.StorageConfig)
	// Create components
	messageStorage := storage.NewInMemoryMessageStorage(config.StorageConfig)
	historyStorage := storage.NewInMemoryHistoryStorage(config.StorageConfig)
	attackRegistry := registry.DefaultRegistry
	attackManager := manager.NewAttackManager(attackRegistry, historyStorage)
	eventPublisher := adapter.NewEventObserver()
	configLoader := NewConfigLoader()

	service := &ByzantineService{
		config:         config,
		attackManager:  attackManager,
		messageStorage: messageStorage,
		historyStorage: historyStorage,
		eventPublisher: eventPublisher,
		attackRegistry: attackRegistry,
		configLoader:   configLoader,
	}

	// Create hook adapter
	service.hookAdapter = adapter.NewHookAdapter(service, eventPublisher, messageStorage)
	service.consensusHook = NewConsensusHook(service.attackManager, service.eventPublisher)

	// Initialize metrics
	service.metrics.Store(types.Metrics{})

	// Set storage provider for replay attack
	provider := &storageProviderImpl{
		messageStorage: messageStorage,
	}
	// TODO: refactoring storage provider to be set globally
	attacks.SetStorageProvider(provider)

	// 4. Register attacks with proper parameter handling
	//for _, attackConfig := range config.Attacks {
	//	// Create attack instance with type-specific parameters
	//	attack, err := createAttackWithParsedParams(attackConfig, attackRegistry)
	//	if err != nil {
	//		log.Error("Failed to create attack", "name", attackConfig.Name, "error", err)
	//		continue
	//	}
	//
	//	// Register the attack
	//	if err := attackManager.RegisterAttack(attack); err != nil {
	//		log.Error("Failed to register attack", "name", attackConfig.Name, "error", err)
	//		continue
	//	}
	//}

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

	// Start background tasks
	go s.metricsCollector()

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

	// Get stored messages count
	if messages, err := s.messageStorage.GetRecentMessages(0); err == nil {
		status.StoredMessages = len(messages)
	}

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

// GetMetrics returns service metrics
func (s *ByzantineService) GetMetrics() types.Metrics {
	return s.metrics.Load().(types.Metrics)
}

// RegisterAttack registers a new attack
func (s *ByzantineService) RegisterAttack(config types.AttackConfig) (string, error) {
	//// Generate UID if not provided
	//if config.UID == "" {
	//	config.UID = s.attackManager.(*manager.AttackManager).GenerateUID(config.Type, config.Code, config.Sequence, config.Round)
	//}

	// Set initial status
	config.Status = types.AttackStatusPending
	config.CreatedAt = time.Now()

	// Create attack instance with type-specific parameters
	//attack, err := createAttackWithParsedParams(config, s.attackRegistry)
	//if err != nil {
	//	log.Error("Failed to create attack", "name", config.Name, "error", err)
	//	return "", fmt.Errorf("failed to create attack: %w", err)
	//}

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

	// Update metrics
	s.updateMetrics(func(m *types.Metrics) {
		m.AttacksRegistered++
	})

	return config.UID, nil
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

// GetAttackHistory gets attack history
func (s *ByzantineService) GetAttackHistory(uid string) ([]types.AttackResult, error) {
	return s.historyStorage.GetAttackHistory(uid)
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

			// Update metrics
			s.updateMetrics(func(m *types.Metrics) {
				m.EventsProcessed++
				if err != nil {
					m.AttacksFailed++
				}
			})

			return err
		})
		if err != nil {
			log.Warn("Failed to subscribe to events: ", err)
		}
	}
}

// loadAttacksFromConfig loads attacks from configuration
func (s *ByzantineService) loadAttacksFromConfig() error {
	for _, attackConfig := range s.config.Attacks {
		if _, err := s.RegisterAttack(attackConfig); err != nil {
			return fmt.Errorf("failed to register attack %s: %w", attackConfig.Name, err)
		}
		log.Info("Attack registered successfully",
			"name", attackConfig.Name,
			"uid", attackConfig.UID,
			"type", attackConfig.Type)
	}
	return nil
}

// updateMetrics updates metrics atomically
func (s *ByzantineService) updateMetrics(fn func(*types.Metrics)) {
	metrics := s.metrics.Load().(types.Metrics)
	fn(&metrics)
	s.metrics.Store(metrics)
}

// metricsCollector collects metrics periodically
func (s *ByzantineService) metricsCollector() {
	if s.ctx == nil {
		log.Error("metricCollector started without context")
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.updateMetrics(func(m *types.Metrics) {
				m.Uptime = int64(time.Since(startTime).Seconds())

				// Get storage size (simplified)
				if messages, err := s.messageStorage.GetRecentMessages(0); err == nil {
					m.MessagesStored = int64(len(messages))
				}
			})
		}
	}
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

// GetMessageStorage returns the message storage (interface for API)
func (s *ByzantineService) GetMessageStorage() types.MessageStorage {
	return s.messageStorage
}

// GetHistoryStorage returns the history storage (interface for API)
func (s *ByzantineService) GetHistoryStorage() types.HistoryStorage {
	return s.historyStorage
}

// GetConsensusHook returns the consensus hook for integration
func (s *ByzantineService) GetConsensusHook() types.ConsensusHook {
	return s.consensusHook
}

//
//// createAttackWithParsedParams creates an attack instance with properly parsed parameters
//func createAttackWithParsedParams(config types.AttackConfig, registry *registry.AttackRegistry) (types.Attack, error) {
//	// Get parameter parser
//	paramRegistry := registry.NewParameterParserRegistry()
//
//	// Parse parameters for the specific attack type
//	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
//	if err != nil {
//		return nil, fmt.Errorf("failed to parse parameters: %w", err)
//	}
//
//	// Create attack instance based on type with parsed parameters
//	switch config.Type {
//	case types.AttackTypeSilentMessage:
//		params := parsedParams.(*types.SilentAttackParams)
//		return createSilentAttack(config, params, registry)
//
//	case types.AttackTypeTamperedMessage:
//		params := parsedParams.(*types.TamperAttackParams)
//		return createTamperAttack(config, params, registry)
//
//	case types.AttackTypeFakeMessage:
//		params := parsedParams.(*types.FakeAttackParams)
//		return createFakeAttack(config, params, registry)
//
//	case types.AttackTypeOmitMessage:
//		params := parsedParams.(*types.OmitAttackParams)
//		return createOmitAttack(config, params, registry)
//
//	case types.AttackTypeRoleSpoofed:
//		params := parsedParams.(*types.RoleSpoofAttackParams)
//		return createRoleSpoofAttack(config, params, registry)
//
//	case types.AttackTypeReplay:
//		params := parsedParams.(*types.ReplayAttackParams)
//		return createReplayAttack(config, params, registry)
//
//	default:
//		return nil, fmt.Errorf("unknown attack type: %s", config.Type)
//	}
//}
//
//// Example attack creation functions (to be implemented in actual attacks)
//func createSilentAttack(config types.AttackConfig, params *types.SilentAttackParams, attackRegistry *registry.AttackRegistry) (types.Attack, error) {
//	// Create attack with parsed parameters
//	attack := &attacks.SilentMessageAttack{
//		BaseAttack: types.NewBaseAttack(config),
//		direction:  types.MessageDirection(params.Direction),
//		targets:    params.Targets,
//	}
//	return attack, nil
//}
//
//func createTamperAttack(config types.AttackConfig, params *types.TamperAttackParams, registry *registry.AttackRegistry) (types.Attack, error) {
//	attack := &TamperAttack{
//		BaseAttack:       types.NewBaseAttack(config),
//		tamperFields:     params.TamperFields,
//		withValidMessage: params.WithValidMessage,
//		delay:            params.Delay,
//		targets:          params.Targets,
//	}
//	return attack, nil
//}
//
//func createFakeAttack(config types.AttackConfig, params *types.FakeAttackParams, registry *registry.AttackRegistry) (types.Attack, error) {
//	attack := &FakeAttack{
//		BaseAttack:  types.NewBaseAttack(config),
//		fakeMessage: params.FakeMessage,
//		targets:     params.Targets,
//	}
//	return attack, nil
//}
//
//func createOmitAttack(config types.AttackConfig, params *types.OmitAttackParams, registry *registry.AttackRegistry) (types.Attack, error) {
//	attack := &OmitAttack{
//		BaseAttack: types.NewBaseAttack(config),
//		cmd:        params.Cmd,
//		cnt:        params.Cnt,
//		targets:    params.Targets,
//	}
//	return attack, nil
//}
//
//func createRoleSpoofAttack(config types.AttackConfig, params *types.RoleSpoofAttackParams, registry *registry.AttackRegistry) (types.Attack, error) {
//	attack := &RoleSpoofAttack{
//		BaseAttack:  types.NewBaseAttack(config),
//		fakeMessage: params.FakeMessage,
//		targets:     params.Targets,
//	}
//	return attack, nil
//}
//
//func createReplayAttack(config types.AttackConfig, params *types.ReplayAttackParams, registry *registry.AttackRegistry) (types.Attack, error) {
//	attack := &ReplayAttack{
//		BaseAttack:      types.NewBaseAttack(config),
//		oriSequence:     params.OriSequence,
//		oriRound:        params.OriRound,
//		useOriginalView: params.UseOriginalView,
//		targets:         params.Targets,
//	}
//	return attack, nil
//}

// Migration helper for existing code
//func MigrateToImprovedConfig(oldConfigPath string) error {
//	// 1. Load with improved loader
//	loader := NewConfigLoader()
//	config, err := loader.LoadConfig(oldConfigPath)
//	if err != nil {
//		return fmt.Errorf("failed to load config: %w", err)
//	}
//
//	// 2. Validate all attacks
//	for i, attack := range config.Attacks {
//		log.Info("Validating attack configuration",
//			"index", i,
//			"name", attack.Name,
//			"uid", attack.UID,
//			"type", attack.Type)
//
//		// Ensure parameters are properly parsed
//		paramRegistry := registry.NewParameterParserRegistry()
//		_, err := paramRegistry.ParseParameters(attack.Type, attack.Parameters)
//		if err != nil {
//			log.Error("Attack parameter validation failed",
//				"name", attack.Name,
//				"error", err)
//		}
//	}
//
//	// 3. Save validated config
//	newPath := oldConfigPath + ".validated"
//	if err := loader.SaveConfig(config, newPath); err != nil {
//		return fmt.Errorf("failed to save validated config: %w", err)
//	}
//
//	log.Info("Configuration migrated successfully",
//		"original", oldConfigPath,
//		"validated", newPath)
//
//	return nil
//}
