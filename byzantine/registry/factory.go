package registry

import (
	"errors"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// BaseAttack provides common functionality for all attacks
type BaseAttack struct {
	uid    string
	config types.AttackConfig
	status types.AttackStatus
	mu     sync.RWMutex
}

// NewBaseAttack creates a new base attack
func NewBaseAttack(config types.AttackConfig) *BaseAttack {
	return &BaseAttack{
		uid:    config.UID,
		config: config,
		status: config.Status,
	}
}

// SetConfig updates the attack configuration
func (b *BaseAttack) SetConfig(config types.AttackConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = config
}

// GetConfig returns the attack configuration
func (b *BaseAttack) GetConfig() types.AttackConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// GetUID returns the unique identifier
func (b *BaseAttack) GetUID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.config.UID
}

// GetType returns the attack type
func (b *BaseAttack) GetType() types.AttackType {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config.Type
}

func (b *BaseAttack) GetStatus() types.AttackStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config.Status
}

// SetStatus updates the attack status
func (b *BaseAttack) SetStatus(status types.AttackStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.status = status
	b.config.Status = status
}

// UpdateParameters updates attack parameters
func (b *BaseAttack) UpdateParameters(params map[string]interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if params == nil {
		return errors.New("parameters cannot be nil")
	}

	b.config.Parameters = params
	// ParsedParameters should be handled by specific attack implementations

	return nil
}

// CanExecute checks if an attack can be executed at a given sequence
func (b *BaseAttack) CanExecute(sequence uint64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.config.IsInSequenceRange(sequence) && b.config.CanExecute()
}

// ParseTargets parses target addresses from config
func ParseTargets(config types.AttackConfig) ([]common.Address, error) {
	if len(config.Parameters["targets"].([]common.Address)) == 0 {
		return nil, nil
	}

	targets := make([]common.Address, len(config.Parameters["targets"].([]common.Address)))
	copy(targets, config.Parameters["targets"].([]common.Address))

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
