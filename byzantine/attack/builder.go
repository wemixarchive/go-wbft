package attack

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// attackBuilder implements AttackBuilder interface
type attackBuilder struct {
	attackType AttackType
	sequence   uint64
	round      uint64
	targets    []common.Address
	options    map[string]interface{}
}

// NewAttackBuilder creates a new attack builder
func NewAttackBuilder() AttackBuilder {
	return &attackBuilder{}
}

// WithType sets the attack type
func (b *attackBuilder) WithType(attackType AttackType) AttackBuilder {
	b.attackType = attackType
	return b
}

// WithTiming sets the sequence and round for attack execution
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

// WithOptions sets additional options for the attack
func (b *attackBuilder) WithOptions(options map[string]interface{}) AttackBuilder {
	b.options = options
	return b
}

// Build creates the attack instance based on configuration
func (b *attackBuilder) Build() (Attack, error) {
	// Validate required fields
	if b.attackType == 0 {
		return nil, errors.New("attack type is required")
	}

	if b.sequence == 0 {
		return nil, errors.New("sequence is required")
	}

	// Create attack based on type
	switch b.attackType {
	case AttackTypeDoublePrepare:
		return b.buildDoublePrepareAttack()
	case AttackTypeDoubleCommit:
		return b.buildDoubleCommitAttack()
	case AttackTypeSilentProposer:
		return b.buildSilentProposerAttack()
	case AttackTypeSilentValidator:
		return b.buildSilentValidatorAttack()
	case AttackTypeTamperedHeader:
		return b.buildTamperedHeaderAttack()
	case AttackTypeFakeTransaction:
		return b.buildFakeTransactionAttack()
	default:
		return nil, fmt.Errorf("unsupported attack type: %d", b.attackType)
	}
}

// buildDoublePrepareAttack creates a double prepare attack
func (b *attackBuilder) buildDoublePrepareAttack() (Attack, error) {
	return &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
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
func (b *attackBuilder) buildDoubleCommitAttack() (Attack, error) {
	return &doubleCommitAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
	}, nil
}

// buildSilentProposerAttack creates a silent proposer attack
func (b *attackBuilder) buildSilentProposerAttack() (Attack, error) {
	return &silentProposerAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
	}, nil
}

// buildSilentValidatorAttack creates a silent validator attack
func (b *attackBuilder) buildSilentValidatorAttack() (Attack, error) {
	return &silentValidatorAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
		messageCode: b.getOptionUint64("messageCode", uint64(MessageCodePrepare)),
	}, nil
}

// buildTamperedHeaderAttack creates a tampered header attack
func (b *attackBuilder) buildTamperedHeaderAttack() (Attack, error) {
	tamperFields, ok := b.options["tamperFields"].([]TamperField)
	if !ok {
		tamperFields = []TamperField{}
	}

	return &tamperedHeaderAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
			sequence: b.sequence,
			round:    b.round,
			targets:  b.targets,
			options:  b.options,
		},
		tamperFields: tamperFields,
	}, nil
}

// buildFakeTransactionAttack creates a fake transaction attack
func (b *attackBuilder) buildFakeTransactionAttack() (Attack, error) {
	return &fakeTransactionAttack{
		baseAttack: baseAttack{
			id:       generateAttackID(),
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

func generateAttackID() string {
	return "attack-" + uuid.New().String()
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

func (a *baseAttack) RequiresData() []DataRequirement {
	// Default: no data requirements
	return []DataRequirement{}
}

// Concrete attack implementations
type doublePrepareAttack struct {
	baseAttack
	delay            uint64
	withValidMessage bool
}

func (a *doublePrepareAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(MessageCodePrepare)
}

func (a *doublePrepareAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type doubleCommitAttack struct {
	baseAttack
}

func (a *doubleCommitAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(MessageCodeCommit)
}

func (a *doubleCommitAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type silentProposerAttack struct {
	baseAttack
}

func (a *silentProposerAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(MessageCodePrePrepare)
}

func (a *silentProposerAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type silentValidatorAttack struct {
	baseAttack
	messageCode uint64
}

func (a *silentValidatorAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == a.messageCode
}

func (a *silentValidatorAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

type tamperedHeaderAttack struct {
	baseAttack
	tamperFields []TamperField
}

func (a *tamperedHeaderAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(MessageCodePrePrepare)
}

func (a *tamperedHeaderAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}

func (a *tamperedHeaderAttack) RequiresData() []DataRequirement {
	return []DataRequirement{
		{
			Type: "blocks",
			Filter: DataFilter{
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

func (a *fakeTransactionAttack) ShouldExecute(sequence, round uint64, msgCode uint64) bool {
	return a.sequence == sequence && a.round == round && msgCode == uint64(MessageCodePrePrepare)
}

func (a *fakeTransactionAttack) Execute(ctx AttackContext) error {
	// Implementation will be added when integrating with QBFT hooks
	return nil
}
