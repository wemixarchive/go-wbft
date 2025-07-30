package types

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

// ByzantineService is the main service interface
type ByzantineService interface {
	// Start starts the service
	Start() error

	// Stop stops the service
	Stop() error

	// Configure configures the service
	Configure(config *ByzantineConfig) error

	// RegisterAttack registers a new attack
	RegisterAttack(config AttackConfig) (string, error)

	// CancelAttack cancels an attack
	CancelAttack(uid string) error

	// ListAttacks lists all attacks
	ListAttacks() []AttackConfig

	// GetStatus returns the service status
	GetStatus() ServiceStatus

	// GetAttackManager gets attack manager
	GetAttackManager() AttackManager

	// GetConsensusHook gets consensus hook
	GetConsensusHook() ConsensusHook
}

// Attack represents the interface for all attack implementations
type Attack interface {
	// CheckExecuteCondition checks if the attack should be executed
	CheckExecuteCondition(ctx context.Context, event Event) bool

	// SetStatus updates the attack status
	SetStatus(status AttackStatus)

	// SetConfig updates the attack configuration
	SetConfig(config AttackConfig)

	// GetUID returns the unique identifier of the attack
	GetUID() string

	// GetType returns the type of the attack
	GetType() AttackType

	// GetConfig returns the attack configuration
	GetConfig() AttackConfig

	// GetStatus returns the current status of the attack
	GetStatus() AttackStatus

	// UpdateParameters updates attack parameters
	UpdateParameters(params map[string]interface{}) error

	// CanExecute checks if an attack can be executed at a given sequence
	CanExecute(sequence uint64) bool
}

// AttackManager manages all registered attacks
type AttackManager interface {
	// RegisterAttack registers a new attack
	RegisterAttack(attack Attack) error

	// UnregisterAttack removes an attack by UID
	UnregisterAttack(uid string) error

	// ProcessEvent processes an event through all attacks
	ProcessEvent(ctx context.Context, event Event) error

	// ProcessEventAsync processes an event asynchronously through all attacks
	ProcessEventAsync(ctx context.Context, event Event) error

	// UpdateStatusMap updates the status of an attack in the internal map
	UpdateStatusMap(attack Attack, newStatus AttackStatus)

	// GetAttack retrieves an attack by UID
	GetAttack(uid string) (Attack, error)

	// GetAttackByUID retrieves an attack by its unique identifier
	GetAttackByUID(uid string) (Attack, bool)

	// GetAttacksByCondition retrieves attacks matching the given condition
	GetAttacksByCondition(attackType AttackType, sequence, round uint64) []Attack

	// ListAttacks returns all registered attacks
	ListAttacks() []Attack

	// GetActiveAttacks returns attacks in active status
	GetActiveAttacks() []Attack

	// GetUIDGenerator returns the UID generator
	GetUIDGenerator() UIDGenerator

	// GetAttacksBySequenceRange retrieves attacks that match the given sequence
	GetAttacksBySequenceRange(attackType AttackType, sequence, round uint64) []Attack

	// FindAttackForExecution finds an attack that can be executed at given sequence/round
	FindAttackForExecution(attackType AttackType, sequence, round uint64) (Attack, bool)

	// FindExecutableAttack finds an executable attack based on type, sequence, round, and message code
	FindExecutableAttack(attackType AttackType, sequence, round uint64, msgCode MessageCode) (Attack, bool)

	// MarkAttackExecuted updates attack execution state
	MarkAttackExecuted(uid string, sequence uint64) error
}

// EventPublisher publishes events
type EventPublisher interface {
	// Publish publishes an event
	Publish(event Event) error

	// Subscribe subscribes to events
	Subscribe(eventType EventType, handler EventHandler) error

	// Unsubscribe removes a subscription
	Unsubscribe(eventType EventType, handler EventHandler) error
}

// EventHandler handles events
type EventHandler func(event Event) error

// HookAdapter provides hooks for consensus integration
type HookAdapter interface {
	// BeforeProposal is called before creating a proposal
	BeforeProposal(ctx context.Context, sequence, round uint64) error

	// AfterProposal is called after creating a proposal
	AfterProposal(ctx context.Context, proposal *types.Block) error

	// BeforePrepare is called before sending a prepare message
	BeforePrepare(ctx context.Context, message *WBFTMessage) error

	// AfterPrepare is called after receiving a prepare message
	AfterPrepare(ctx context.Context, message *WBFTMessage) error

	// BeforeCommit is called before sending a commit message
	BeforeCommit(ctx context.Context, message *WBFTMessage) error

	// AfterCommit is called after receiving a commit message
	AfterCommit(ctx context.Context, message *WBFTMessage) error

	// OnRoundChange is called on round change
	OnRoundChange(ctx context.Context, sequence, round uint64) error

	// OnMessageReceive is called when receiving any message
	OnMessageReceive(ctx context.Context, message *WBFTMessage) error
}

// ServiceStatus represents the service status
type ServiceStatus struct {
	Running         bool       `json:"running"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	ActiveAttacks   int        `json:"active_attacks"`
	ExecutedAttacks int        `json:"executed_attacks"`
	FailedAttacks   int        `json:"failed_attacks"`
	StoredMessages  int        `json:"stored_messages"`
}

// Metrics represents service metrics
type Metrics struct {
	AttacksRegistered int64 `json:"attacks_registered"`
	AttacksExecuted   int64 `json:"attacks_executed"`
	AttacksFailed     int64 `json:"attacks_failed"`
	MessagesStored    int64 `json:"messages_stored"`
	EventsProcessed   int64 `json:"events_processed"`
	StorageSize       int64 `json:"storage_size_bytes"`
	Uptime            int64 `json:"uptime_seconds"`
}

// UIDGenerator Format: "attackType-Code-Sequence-Round"
type UIDGenerator interface {
	// Generate with 3 parameters for backward compatibility
	Generate(attackType AttackType, sequence, round uint64) string
	// GenerateWithRange with 4 parameters for range support
	GenerateWithRange(attackType AttackType, sequenceStart, sequenceEnd, round uint64) string
	GenerateForLookup(attackType AttackType, sequence, round uint64) []string
	Parse(uid string) (AttackType, uint64, uint64, error)
	ParseRange(uid string) (AttackType, uint64, uint64, uint64, error)
	IsWildcard(uid string) bool
}

// ConsensusHook represents consensus hook
type ConsensusHook interface {
	// GetExecutableAttacks returns a map keyed by AttackType,
	// where each value contains the full attack configuration and its
	// parsed parameters (e.g. TamperAttackParams, FakeAttackParams).
	// Only attacks that are enabled, match the given message code, and
	// satisfy runtime execution conditions will be included.
	GetExecutableAttacks(msgCode MessageCode, sequence, round uint64) map[AttackType]*ExecutableAttack
	ShouldExecuteAttack(attackType AttackType, msgCode, sequence, round uint64) (*AttackConfig, bool)
	GetAttackConfig(attackType AttackType, sequence, round uint64) (*AttackConfig, error)
	MarkAttackExecuted(uid string, sequence uint64) error
}

// AttackParamsParser defines an interface for parsing and validating attack parameters
type AttackParamsParser interface {
	Parse(raw map[string]interface{}) (interface{}, error)
	ParseJSON(data []byte) (interface{}, error)
	Validate() error
	ValidateWith(params interface{}) error
	HasMessageCode(code MessageCode) bool
}

type AttackExecutionContext struct {
	Config           AttackConfig
	CurrentSequence  uint64
	CurrentRound     uint64
	MessageCode      MessageCode
	ExecutionCount   uint64
	IsFirstExecution bool
}
