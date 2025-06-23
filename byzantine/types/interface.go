package types

import (
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
	ConfigureAttack(params AttackParams) (attackID string, err error)
	StopAttack(attackID string) error
	ListAttacks() []AttackInfo
}

// AttackManager interface for managing Byzantine attacks
type AttackManager interface {
	// LoadAttack loads an attack from configuration
	LoadAttack(config AttackConfig) error

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
	Start(dataReqs []DataRequirement) error
	// Stop stops event collection
	Stop(attackId string) error
	GetHistoricalData(filter DataFilter) []CollectedData
	GetStateHistory(sequence uint64) []StateTransition
}

// Attack interface for individual attacks
type Attack interface {
	ID() string
	Config() AttackConfig
	ShouldExecute(sequence, round uint64, msgCode uint64) bool
	Execute(ctx AttackContext) error
	//RequiresData() []DataRequirement // RequiresData indicates if the attack needs additional data
}

type AttackBuilder interface {
	WithType(attackType AttackType) AttackBuilder
	WithConfig(config AttackConfig) AttackBuilder
	WithTiming(sequence, round uint64) AttackBuilder
	WithTargets(targets []common.Address) AttackBuilder
	WithOptions(options map[string]interface{}) AttackBuilder
	Build() (Attack, error)
	Reset() AttackBuilder
}
