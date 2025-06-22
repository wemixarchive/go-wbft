package byzantine

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// attackBuilder implements AttackBuilder interface
type attackBuilder struct {
	// Builder state
	attackType types.AttackType
	config     *types.AttackConfig
	sequence   uint64
	round      uint64
	targets    []common.Address
	options    map[string]interface{}
}

// NewAttackBuilder creates a new attack builder
func NewAttackBuilder() AttackBuilder {
	return &attackBuilder{}
}

// Reset clears the builder state
func (b *attackBuilder) Reset() AttackBuilder {
	b.attackType = ""
	b.config = nil
	b.sequence = 0
	b.round = 0
	b.targets = nil
	b.options = nil
	return b
}

// WithType sets the attack type
func (b *attackBuilder) WithType(attackType types.AttackType) AttackBuilder {
	b.attackType = attackType
	return b
}

// WithConfig sets the entire attack config
func (b *attackBuilder) WithConfig(config types.AttackConfig) AttackBuilder {
	b.config = &config
	// Extract values from config
	b.attackType = config.Type
	b.sequence = config.Sequence
	b.round = config.Round
	b.targets = config.GetTargetAddresses()
	b.options = config.Params
	return b
}

// WithTiming sets the sequence and round
func (b *attackBuilder) WithTiming(sequence, round uint64) AttackBuilder {
	b.sequence = sequence
	b.round = round
	return b
}

// WithTargets sets the target validators
func (b *attackBuilder) WithTargets(targets []common.Address) AttackBuilder {
	b.targets = targets
	return b
}

// WithOptions sets additional options
func (b *attackBuilder) WithOptions(options map[string]interface{}) AttackBuilder {
	b.options = options
	return b
}

// Build creates the attack instance
func (b *attackBuilder) Build() (Attack, error) {
	// Validate required fields
	if b.attackType == "" {
		return nil, errors.New("attack type is required")
	}

	// Create config if not provided
	if b.config == nil {
		b.config = &types.AttackConfig{
			Type:     b.attackType,
			Sequence: b.sequence,
			Round:    b.round,
			Params:   b.options,
			Enabled:  true,
		}

		// Convert addresses to strings
		if len(b.targets) > 0 {
			b.config.Targets = make([]string, len(b.targets))
			for i, addr := range b.targets {
				b.config.Targets[i] = addr.Hex()
			}
		}
	}

	// Generate unique ID
	id := fmt.Sprintf("%s-%s", b.attackType, uuid.New().String()[:8])

	// Create attack based on type
	switch b.attackType {
	case types.AttackTypeSilent:
		return &silentAttack{
			baseAttack: baseAttack{
				id:     id,
				config: *b.config,
			},
		}, nil

	case types.AttackTypeTamper:
		return &tamperAttack{
			baseAttack: baseAttack{
				id:     id,
				config: *b.config,
			},
		}, nil

	case types.AttackTypeFake:
		return &fakeAttack{
			baseAttack: baseAttack{
				id:     id,
				config: *b.config,
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported attack type: %s", b.attackType)
	}
}
