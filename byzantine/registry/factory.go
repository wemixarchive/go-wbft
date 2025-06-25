package registry

import (
	"github.com/ethereum/go-ethereum/common"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
)

// BaseAttack provides common functionality for all attacks
type BaseAttack struct {
	config types.AttackConfig
	mu     sync.RWMutex
}

// NewBaseAttack creates a new base attack
func NewBaseAttack(config types.AttackConfig) *BaseAttack {
	return &BaseAttack{
		config: config,
	}
}

// GetUID returns the unique identifier
func (a *BaseAttack) GetUID() uint64 {
	return a.config.UID
}

// GetType returns the attack type
func (a *BaseAttack) GetType() types.AttackType {
	return a.config.Type
}

// GetConfig returns the attack configuration
func (a *BaseAttack) GetConfig() types.AttackConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// SetStatus updates the attack status
func (a *BaseAttack) SetStatus(status types.AttackStatus) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config.Status = status
}

// ParseTargets parses target addresses from config
func ParseTargets(config types.AttackConfig) ([]common.Address, error) {
	if len(config.Targets) == 0 {
		return nil, nil
	}

	targets := make([]common.Address, len(config.Targets))
	copy(targets, config.Targets)

	return targets, nil
}

// GetParameter retrieves a parameter value from config
func GetParameter(config types.AttackConfig, key string, defaultValue interface{}) interface{} {
	if value, exists := config.Parameters[key]; exists {
		return value
	}
	return defaultValue
}

// GetBoolParameter retrieves a boolean parameter
func GetBoolParameter(config types.AttackConfig, key string, defaultValue bool) bool {
	value := GetParameter(config, key, defaultValue)
	if boolValue, ok := value.(bool); ok {
		return boolValue
	}
	return defaultValue
}

// GetUint64Parameter retrieves a uint64 parameter
func GetUint64Parameter(config types.AttackConfig, key string, defaultValue uint64) uint64 {
	value := GetParameter(config, key, defaultValue)
	switch v := value.(type) {
	case uint64:
		return v
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	default:
		return defaultValue
	}
}
