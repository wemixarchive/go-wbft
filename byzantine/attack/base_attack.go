package attack

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// BaseAttack provides common functionality for all attacks
type BaseAttack struct {
	// Configuration
	name     string
	sequence uint64
	round    uint64
	targets  []common.Address

	// Condition
	condition types.Condition

	// State
	status    types.AttackStatus
	statusMu  sync.RWMutex
	startTime time.Time

	// Logger
	logger log.Logger
}

// NewBaseAttack creates a new base attack
func NewBaseAttack(name string, sequence, round uint64, targets []common.Address, logger log.Logger) *BaseAttack {
	return &BaseAttack{
		name:      name,
		sequence:  sequence,
		round:     round,
		targets:   targets,
		status:    types.AttackStatusPending,
		condition: nil, // TODO: should refactoring
		logger:    logger,
	}
}

// Name returns the attack name
func (b *BaseAttack) Name() string {
	return b.name
}

// Sequence returns the target sequence
func (b *BaseAttack) Sequence() uint64 {
	return b.sequence
}

// Round returns the target round
func (b *BaseAttack) Round() uint64 {
	return b.round
}

// Status returns the current status
func (b *BaseAttack) Status() types.AttackStatus {
	b.statusMu.RLock()
	defer b.statusMu.RUnlock()
	return b.status
}

// SetStatus updates the attack status
func (b *BaseAttack) SetStatus(status types.AttackStatus) {
	b.statusMu.Lock()
	defer b.statusMu.Unlock()

	oldStatus := b.status
	b.status = status

	b.logger.Debug("Attack status changed",
		"name", b.name,
		"old", oldStatus,
		"new", status)
}

// Start starts the attack
func (b *BaseAttack) Start() error {
	b.statusMu.Lock()
	defer b.statusMu.Unlock()

	if b.status != types.AttackStatusPending {
		return nil // Already started
	}

	b.status = types.AttackStatusActive
	b.startTime = time.Now()

	b.logger.Info("Attack started",
		"name", b.name,
		"sequence", b.sequence,
		"round", b.round)

	return nil
}

// Stop stops the attack
func (b *BaseAttack) Stop() error {
	b.statusMu.Lock()
	defer b.statusMu.Unlock()

	if b.status == types.AttackStatusStopped {
		return nil // Already stopped
	}

	b.status = types.AttackStatusStopped

	b.logger.Info("Attack stopped",
		"name", b.name,
		"duration", time.Since(b.startTime))

	return nil
}

// ShouldExecute checks if the attack should execute
func (b *BaseAttack) ShouldExecute(sequence, round uint64) bool {
	// Check status
	if b.Status() != types.AttackStatusActive {
		return false
	}

	// Check sequence and round
	if b.sequence != 0 && sequence != b.sequence {
		return false
	}

	if b.round != 0 && round != b.round {
		return false
	}

	// Evaluate condition
	ctx := &types.ConditionContext{
		Sequence: sequence,
		Round:    round,
		// Other fields will be filled by the caller
	}

	return b.condition.Evaluate(ctx)
}

// SetCondition sets the execution condition
func (b *BaseAttack) SetCondition(condition types.Condition) {
	b.condition = condition
}

// GetTargets returns the target addresses
func (b *BaseAttack) GetTargets() []common.Address {
	return b.targets
}

// IsTargeted checks if an address is targeted
func (b *BaseAttack) IsTargeted(addr common.Address) bool {
	if len(b.targets) == 0 {
		return true // No specific targets means all are targeted
	}

	for _, target := range b.targets {
		if target == addr {
			return true
		}
	}

	return false
}

// RecordExecution records an execution result
func (b *BaseAttack) RecordExecution(success bool, err error) {
	result := types.AttackResult{
		Success:   success,
		Error:     err,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"sequence": b.sequence,
			"round":    b.round,
		},
	}

	if success {
		b.logger.Debug("Attack executed successfully",
			"name", b.name,
			"details", result.Details)
	} else {
		b.logger.Error("Attack execution failed",
			"name", b.name,
			"error", err,
			"details", result.Details)
	}

	// Update status if this was a one-time attack
	if b.sequence != 0 && b.round != 0 {
		b.SetStatus(types.AttackStatusExecuted)
	}
}
