package attacks

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

var (
	ErrAttackTypeExists   = errors.New("attacks type already registered")
	ErrAttackTypeNotFound = errors.New("attacks type not found")
	ErrInvalidAttackType  = errors.New("invalid attacks type")
)

// AttackFactory creates attacks instances
type AttackFactory func(config types.AttackConfig) (types.Attack, error)

// Registry manages attacks type registration and creation
type Registry struct {
	factories map[types.AttackType]AttackFactory
	mu        sync.RWMutex
	logger    log.Logger
}

// NewRegistry creates a new attacks registry
func NewRegistry(logger log.Logger) *Registry {
	return &Registry{
		factories: make(map[types.AttackType]AttackFactory),
		logger:    logger,
	}
}

// Register registers an attacks factory
func (r *Registry) Register(attackType types.AttackType, factory AttackFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[attackType]; exists {
		return ErrAttackTypeExists
	}

	r.factories[attackType] = factory
	r.logger.Info("Registered attacks type", "type", attackType)
	return nil
}

// Unregister removes an attacks factory
func (r *Registry) Unregister(attackType types.AttackType) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[attackType]; !exists {
		return ErrAttackTypeNotFound
	}

	delete(r.factories, attackType)
	r.logger.Info("Unregistered attacks type", "type", attackType)
	return nil
}

// CreateAttack creates an attacks instance
func (r *Registry) CreateAttack(config *types.AttackConfig) (types.Attack, error) {
	r.mu.RLock()
	factory, exists := r.factories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrAttackTypeNotFound, config.Type)
	}

	attack, err := factory(*config)
	if err != nil {
		return nil, fmt.Errorf("failed to create attacks: %w", err)
	}

	r.logger.Debug("Created attacks instance",
		"type", config.Type,
		"name", config.Name,
		"sequence", config.Sequence,
		"round", config.Round)

	return attack, nil
}

// GetRegisteredTypes returns all registered attacks types
func (r *Registry) GetRegisteredTypes() []types.AttackType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]types.AttackType, 0, len(r.factories))
	for attackType := range r.factories {
		types = append(types, attackType)
	}

	return types
}

// IsRegistered checks if an attacks type is registered
func (r *Registry) IsRegistered(attackType types.AttackType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.factories[attackType]
	return exists
}

// Clear removes all registered attacks types
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.factories = make(map[types.AttackType]AttackFactory)
	r.logger.Info("Cleared attacks registry")
}

// DefaultRegistry is the global attacks registry
var DefaultRegistry *Registry

// InitDefaultRegistry initializes the default registry
func InitDefaultRegistry(logger log.Logger) {
	DefaultRegistry = NewRegistry(logger)

	// Register built-in attacks types
	registerBuiltinAttacks()
}

// registerBuiltinAttacks registers all built-in attacks types
func registerBuiltinAttacks() {
	// Register SilentAttack
	DefaultRegistry.Register(types.AttackTypeSilentMessage, SilentAttackFactory)

	// Register TamperAttack
	DefaultRegistry.Register(types.AttackTypeTamperedMessage, TamperAttackFactory)

	// Register FakeAttack
	DefaultRegistry.Register(types.AttackTypeFakeMessage, FakeAttackFactory)

	// Register OmitAttack
	DefaultRegistry.Register(types.AttackTypeOmitMessage, OmitAttackFactory)

	// Register RoleSpoofAttack
	DefaultRegistry.Register(types.AttackTypeRoleSpoofed, RoleSpoofAttackFactory)

	// Register ReplayAttack
	DefaultRegistry.Register(types.AttackTypeReplay, ReplayAttackFactory)
}
