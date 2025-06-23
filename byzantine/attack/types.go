package attack

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// AttackContext provides context for attack execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
	// Add more context as needed
}

// TamperField specifies a field to tamper and its new value
type TamperField struct {
	Target string      // e.g., "Proposal.Header.Coinbase"
	Value  interface{} // New value for the field
}

// AttackStatus defines the status of an attack
type AttackStatus string

const (
	AttackStatusActive    AttackStatus = "active"
	AttackStatusExecuted  AttackStatus = "executed"
	AttackStatusCancelled AttackStatus = "cancelled"
)

// Attack interface defines the contract for all attack implementations
type Attack interface {
	ID() string
	ShouldExecute(sequence, round uint64, msgCode uint64) bool
	Execute(ctx AttackContext) error
	RequiresData() []types.DataRequirement
}

// AttackBuilder interface for building attacks
type AttackBuilder interface {
	WithType(attackType types.AttackType) AttackBuilder
	WithTiming(sequence, round uint64) AttackBuilder
	WithTargets(targets []common.Address) AttackBuilder
	WithOptions(options map[string]interface{}) AttackBuilder
	Build() (Attack, error)
}
