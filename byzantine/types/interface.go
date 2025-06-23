package types

import (
	"github.com/ethereum/go-ethereum/common"
	wbftmessage "github.com/ethereum/go-ethereum/consensus/wbft/messages"
	"github.com/ethereum/go-ethereum/core/types"
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
	ByzantineTests() []AttackInfo
	StopByzantineTests(uids []uint64) error
	SilentMessage(params SilentMessageParams) error
	SendTamperedMessage(params TamperedMessageParams) error
	SendFakeMessage(params FakeMessageParams) error
	SendRoleSpoofedMessage(params RoleSpoofParams) error
	SendReplayMessage(params ReplayMessageParams) error
	UpgradeGovContract() error
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

	Register(attack Attack) error
	Unregister(attackID string) error
	GetAttack(attackID string) (Attack, bool)
	GetAttackStatus(attackID string) AttackStatus
	MarkExecuted(attackID string)
	ListAllAttacks() []Attack
	CleanupExecuted() int
	Builder() AttackBuilder
	ConfigureSilentMessage(sequence, round, code, direction uint64, targets []common.Address) error
	ConfigureTamperedMessage(params TamperMessageParams) error
	ConfigureFakeMessage(params FakeMessageParams) error
	ConfigureOmitMessage(params OmitMessageParams) error
	ConfigureRoleSpoofedMessage(params RoleSpoofParams) error
	ConfigureReplayMessage(params ReplayMessageParams) error
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
	Execute(ctx *AttackContext) error
	RequiresData() []DataRequirement // RequiresData indicates if the attack needs additional data
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

type Message interface {
	// Basic properties
	Code() MessageCode
	Sequence() uint64
	Round() uint64
	From() common.Address
	To() []common.Address

	// Message content
	Payload() []byte
	Signature() []byte

	// Validation
	Validate() error
	VerifySignature() error

	// Encoding/decoding
	Encode() ([]byte, error)
	Decode([]byte) error

	// Block reference
	Block() *types.Block

	// QBFT message conversion
	QBFTMessage() wbftmessage.WBFTMessage

	// Cloning
	Clone() Message
}

// Condition represents a condition for attack execution
type Condition interface {
	// Evaluate returns true if the condition is satisfied
	Evaluate(ctx *ConditionContext) bool

	// String returns a human-readable description
	String() string
}
