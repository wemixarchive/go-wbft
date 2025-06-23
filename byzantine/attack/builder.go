package attack

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// attackBuilder implements AttackBuilder interface
type attackBuilder struct {
	attackType types.AttackType
	config     *types.AttackConfig
	sequence   uint64
	round      uint64
	targets    []common.Address
	options    map[string]interface{}
}

var _ types.AttackBuilder = (*attackBuilder)(nil)

// NewAttackBuilder creates a new attack builder
func NewAttackBuilder() types.AttackBuilder {
	return &attackBuilder{}
}

// WithType sets the attack type
func (b *attackBuilder) WithType(attackType types.AttackType) types.AttackBuilder {
	b.attackType = attackType
	return b
}

// WithConfig sets the entire attack config
func (b *attackBuilder) WithConfig(config types.AttackConfig) types.AttackBuilder {
	b.config = &config
	// Extract values from config
	b.attackType = config.Type
	b.sequence = config.Sequence
	b.round = config.Round
	b.targets = config.GetTargetAddresses()
	b.options = config.Params
	return b
}

// WithTiming sets the sequence and round for attack execution
func (b *attackBuilder) WithTiming(sequence, round uint64) types.AttackBuilder {
	b.sequence = sequence
	b.round = round
	return b
}

// WithTargets sets the target validators
func (b *attackBuilder) WithTargets(targets []common.Address) types.AttackBuilder {
	b.targets = targets
	return b
}

// WithOptions sets additional options for the attack
func (b *attackBuilder) WithOptions(options map[string]interface{}) types.AttackBuilder {
	b.options = options
	return b
}

// Build creates the attack instance based on configuration
func (b *attackBuilder) Build() (types.Attack, error) {
	// Validate required fields
	if b.attackType == "" {
		return nil, errors.New("attack type is required")
	}

	if b.sequence == 0 {
		return nil, errors.New("sequence is required")
	}

	if b.config == nil {
		// TODO:
		// should refactoring
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

	// Create attack based on type
	switch b.attackType {
	case types.AttackTypeDoublePrepare:
		return b.buildDoublePrepareAttack()
	case types.AttackTypeDoubleCommit:
		return b.buildDoubleCommitAttack()
	case types.AttackTypeSilentProposer:
		return b.buildSilentProposerAttack()
	case types.AttackTypeSilentValidator:
		return b.buildSilentValidatorAttack()
	case types.AttackTypeTamperedHeader:
		return b.buildTamperedHeaderAttack()
	case types.AttackTypeFakeTransaction:
		return b.buildFakeTransactionAttack()
	default:
		return nil, fmt.Errorf("unsupported attack type: %d", b.attackType)
	}
}

// Reset clears the builder state
func (b *attackBuilder) Reset() types.AttackBuilder {
	b.attackType = ""
	b.config = nil
	b.sequence = 0
	b.round = 0
	b.targets = nil
	b.options = nil
	return b
}

// buildDoublePrepareAttack creates a double prepare attack
func (b *attackBuilder) buildDoublePrepareAttack() (types.Attack, error) {
	return &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
		delay:            b.getOptionUint64("delay", 0),
		withValidMessage: b.getOptionBool("withValidMessage", false),
	}, nil
}

// buildDoubleCommitAttack creates a double commit attack
func (b *attackBuilder) buildDoubleCommitAttack() (types.Attack, error) {
	return &doubleCommitAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
	}, nil
}

// buildSilentProposerAttack creates a silent proposer attack
func (b *attackBuilder) buildSilentProposerAttack() (types.Attack, error) {
	return &silentProposerAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
	}, nil
}

// buildSilentValidatorAttack creates a silent validator attack
func (b *attackBuilder) buildSilentValidatorAttack() (types.Attack, error) {
	return &silentValidatorAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
		messageCode: b.getOptionUint64("messageCode", uint64(types.MessageCodePrepare)),
	}, nil
}

// buildTamperedHeaderAttack creates a tampered header attack
func (b *attackBuilder) buildTamperedHeaderAttack() (types.Attack, error) {
	tamperFields, ok := b.options["tamperFields"].([]TamperField)
	if !ok {
		tamperFields = []TamperField{}
	}

	return &tamperedHeaderAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
		tamperFields: tamperFields,
	}, nil
}

// buildFakeTransactionAttack creates a fake transaction attack
func (b *attackBuilder) buildFakeTransactionAttack() (types.Attack, error) {
	return &fakeTransactionAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(string(b.attackType)),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
	}, nil
}

// Helper methods
func (b *attackBuilder) getOptionUint64(key string, defaultValue uint64) uint64 {
	if b.options == nil {
		return defaultValue
	}
	if val, ok := b.options[key].(uint64); ok {
		return val
	}
	return defaultValue
}

func (b *attackBuilder) getOptionBool(key string, defaultValue bool) bool {
	if b.options == nil {
		return defaultValue
	}
	if val, ok := b.options[key].(bool); ok {
		return val
	}
	return defaultValue
}

func generateAttackID(attackType string) string {
	return fmt.Sprintf("%s-%s", attackType, uuid.New().String()[:8])
}

// Base attack structure
type baseAttack struct {
	id       string
	sequence uint64
	round    uint64
	targets  []common.Address
	options  map[string]interface{}
}

func (a *baseAttack) ID() string {
	return a.id
}

func (a *baseAttack) RequiresData() []types.DataRequirement {
	// Default: no data requirements
	return []types.DataRequirement{}
}

// Concrete attack implementations
type doublePrepareAttack struct {
	baseAttack
	delay            uint64
	withValidMessage bool
}

func (a *doublePrepareAttack) ID() string {
	return a.id
}

func (a *doublePrepareAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *doublePrepareAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(types.MessageCodePrepare)
}

func (a *doublePrepareAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type doubleCommitAttack struct {
	baseAttack
}

func (a *doubleCommitAttack) ID() string {
	return a.id
}

func (a *doubleCommitAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *doubleCommitAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(types.MessageCodeCommit)
}

func (a *doubleCommitAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type silentProposerAttack struct {
	baseAttack
}

func (a *silentProposerAttack) ID() string {
	return a.id
}

func (a *silentProposerAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *silentProposerAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(types.MessageCodePrePrepare)
}

func (a *silentProposerAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type silentValidatorAttack struct {
	baseAttack
	messageCode uint64
}

func (a *silentValidatorAttack) ID() string {
	return a.id
}

func (a *silentValidatorAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *silentValidatorAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == a.messageCode
}

func (a *silentValidatorAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type tamperedHeaderAttack struct {
	baseAttack
	tamperFields []TamperField
}

func (a *tamperedHeaderAttack) ID() string {
	return a.id
}

func (a *tamperedHeaderAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *tamperedHeaderAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(types.MessageCodePrePrepare)
}

func (a *tamperedHeaderAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

func (a *tamperedHeaderAttack) RequiresData() []types.DataRequirement {
	return []types.DataRequirement{
		{
			Type: "blocks",
			Filter: types.DataFilter{
				FromSequence: a.sequence - 10,
				ToSequence:   a.sequence - 1,
			},
			MaxRecords: 10,
		},
	}
}

type fakeTransactionAttack struct {
	baseAttack
}

func (a *fakeTransactionAttack) ID() string {
	return a.id
}

func (a *fakeTransactionAttack) Config() types.AttackConfig {
	return types.AttackConfig{}
}

func (a *fakeTransactionAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(types.MessageCodePrePrepare)
}

func (a *fakeTransactionAttack) Execute(ctx types.AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}
