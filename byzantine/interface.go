package byzantine

import (
	"github.com/ethereum/go-ethereum/byzantine/attack"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

// Service interface for Byzantine service
type Service interface {
	// Start starts the service
	Start() error

	// Stop stops the service
	Stop() error

	// APIs returns RPC APIs
	APIs() []rpc.API
}

type ByzantineAPI interface {
	ConfigureAttack(params attack.AttackParams) (attackID string, err error)
	StopAttack(attackID string) error
	ListAttacks() []AttackInfo
}

// AttackManager interface for managing Byzantine attacks
type AttackManager interface {
	// LoadAttack loads an attack from configuration
	LoadAttack(config types.AttackConfig) error

	// ConfigureAttack configures a new attack via API
	ConfigureAttack(params AttackParams) (string, error)

	// ListAttacks returns all configured attacks
	ListAttacks() []AttackInfo

	// StopAttack stops a specific attack
	StopAttack(attackID string) error

	// GetAttacksToExecute returns attacks that should execute
	GetAttacksToExecute(sequence, round uint64, msgCode uint64) []Attack

	// Start starts the attack manager
	Start() error

	// Stop stops the attack manager
	Stop() error
}

// EventCollector interface for collecting consensus events
type EventCollector interface {
	// Start starts event collection
	Start() error

	// Stop stops event collection
	Stop() error
}

// Attack interface for individual attacks
type Attack interface {
	ID() string
	Config() types.AttackConfig
	ShouldExecute(sequence, round uint64, msgCode uint64) bool
	Execute(ctx AttackContext) error
	//RequiresData() []DataRequirement // RequiresData indicates if the attack needs additional data
}

// AttackContext provides context for attack execution
type AttackContext struct {
	CurrentSequence uint64
	CurrentRound    uint64
	MessageCode     uint64
}

// AttackParams for API
type AttackParams struct {
	Type     string                 `json:"type"`
	Sequence uint64                 `json:"sequence"`
	Round    uint64                 `json:"round"`
	Targets  []common.Address       `json:"targets"`
	Options  map[string]interface{} `json:"options"`
}

// AttackInfo for API responses
type AttackInfo struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Type     types.AttackType   `json:"type"`
	Sequence uint64             `json:"sequence"`
	Round    uint64             `json:"round"`
	Status   types.AttackStatus `json:"status"`
	Targets  []common.Address   `json:"targets"`
}

type AttackBuilder interface {
	WithType(attackType types.AttackType) AttackBuilder
	WithConfig(config types.AttackConfig) AttackBuilder
	WithTiming(sequence, round uint64) AttackBuilder
	WithTargets(targets []common.Address) AttackBuilder
	WithOptions(options map[string]interface{}) AttackBuilder
	Build() (Attack, error)
	Reset() AttackBuilder
}
