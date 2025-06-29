package registry

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// AttackFactory creates attack instances
type AttackFactory func(config types.AttackConfig) (types.Attack, error)

// AttackRegistry manages attack type registrations
type AttackRegistry struct {
	mu        sync.RWMutex
	factories map[types.AttackType]AttackFactory
}

// NewAttackRegistry creates a new attack registry
func NewAttackRegistry() *AttackRegistry {
	return &AttackRegistry{
		factories: make(map[types.AttackType]AttackFactory),
	}
}

// Register registers an attack factory
func (r *AttackRegistry) Register(attackType types.AttackType, factory AttackFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[attackType]; exists {
		return fmt.Errorf("attack type %s already registered", attackType)
	}

	r.factories[attackType] = factory
	return nil
}

// CreateAttack creates an attack instance
func (r *AttackRegistry) CreateAttack(config types.AttackConfig) (types.Attack, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("unknown attack type: %s", config.Type)
	}

	return factory(config)
}

// GetRegisteredTypes returns all registered attack types
func (r *AttackRegistry) GetRegisteredTypes() []types.AttackType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registeredTypeList := make([]types.AttackType, 0, len(r.factories))
	for attackType := range r.factories {
		registeredTypeList = append(registeredTypeList, attackType)
	}

	return registeredTypeList
}

// DefaultRegistry is the global attack registry
var DefaultRegistry = NewAttackRegistry()

// Register registers an attack factory to the default registry
func Register(attackType types.AttackType, factory AttackFactory) error {
	return DefaultRegistry.Register(attackType, factory)
}

// CreateAttack creates an attack instance using the default registry
func CreateAttack(config types.AttackConfig) (types.Attack, error) {
	return DefaultRegistry.CreateAttack(config)
}

func CreateAttackWithParam(config types.AttackConfig, params *types.SilentAttackParams) (types.Attack, error) {
	return DefaultRegistry.CreateAttack(config)
}
