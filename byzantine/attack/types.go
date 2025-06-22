package attack

import "github.com/ethereum/go-ethereum/common"

// AttackType defines the type of byzantine attack
type AttackType uint64

const (
	AttackTypeDoublePrepare AttackType = iota + 1
	AttackTypeDoubleCommit
	AttackTypeSilentProposer
	AttackTypeSilentValidator
	AttackTypeTamperedHeader
	AttackTypeFakeTransaction
	// Add more attack types as needed
)

// MessageCode defines QBFT message types
type MessageCode uint64

const (
	MessageCodePrePrepare MessageCode = 1 << iota
	MessageCodePrepare
	MessageCodeCommit
	MessageCodeRoundChange
	MessageCodeRoundChangePrePrepare
	MessageCodePropagation
)

// AttackParams contains parameters for configuring an attack
type AttackParams struct {
	Type     AttackType
	Sequence uint64
	Round    uint64
	Target   []common.Address
	Options  map[string]interface{}
}

// AttackContext provides context for attack execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
	// Add more context as needed
}

// DataRequirement specifies what data an attack needs
type DataRequirement struct {
	Type       string // "messages", "blocks", "state"
	Filter     DataFilter
	MaxRecords int
}

// DataFilter for querying historical data
type DataFilter struct {
	FromSequence uint64
	ToSequence   uint64
	MessageTypes []uint64
	Validators   []common.Address
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

// AttackInfo provides information about a configured attack
type AttackInfo struct {
	ID       string
	Type     AttackType
	Sequence uint64
	Round    uint64
	Status   string // "active", "executed", "cancelled"
}

// Attack interface defines the contract for all attack implementations
type Attack interface {
	ID() string
	ShouldExecute(sequence, round uint64, msgCode uint64) bool
	Execute(ctx AttackContext) error
	RequiresData() []DataRequirement
}

// AttackBuilder interface for building attacks
type AttackBuilder interface {
	WithType(attackType AttackType) AttackBuilder
	WithTiming(sequence, round uint64) AttackBuilder
	WithTargets(targets []common.Address) AttackBuilder
	WithOptions(options map[string]interface{}) AttackBuilder
	Build() (Attack, error)
}
