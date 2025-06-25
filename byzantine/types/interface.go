package types

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Attack represents the interface for all attack implementations
type Attack interface {
	// GetUID returns the unique identifier of the attack
	GetUID() uint64

	// GetType returns the type of the attack
	GetType() AttackType

	// CheckExecuteCondition checks if the attack should be executed
	CheckExecuteCondition(ctx context.Context, event Event) bool

	// Execute performs the attack
	Execute(ctx context.Context, event Event) (*AttackResult, error)

	// GetConfig returns the attack configuration
	GetConfig() AttackConfig

	// SetStatus updates the attack status
	SetStatus(status AttackStatus)
}

// AttackManager manages all registered attacks
type AttackManager interface {
	// RegisterAttack registers a new attack
	RegisterAttack(attack Attack) error

	// UnregisterAttack removes an attack by UID
	UnregisterAttack(uid uint64) error

	// GetAttack retrieves an attack by UID
	GetAttack(uid uint64) (Attack, error)

	// ListAttacks returns all registered attacks
	ListAttacks() []Attack

	// ProcessEvent processes an event through all attacks
	ProcessEvent(ctx context.Context, event Event) error

	// GetActiveAttacks returns attacks in active status
	GetActiveAttacks() []Attack

	EvaluateAndExecuteAttacks(ctx context.Context, event Event) (AttackDecision, error)

	ProcessEventAsync(ctx context.Context, event Event) error
}

// MessageStorage handles message storage and retrieval
type MessageStorage interface {
	// Store stores a message
	Store(message *StoredMessage) error

	// GetByHash retrieves a message by hash
	GetByHash(hash common.Hash) (*StoredMessage, error)

	// GetBySequenceRound retrieves messages by sequence and round
	GetBySequenceRound(sequence, round uint64) ([]*StoredMessage, error)

	// GetRecentMessages retrieves recent messages
	GetRecentMessages(limit int) ([]*StoredMessage, error)

	// Prune removes old messages
	Prune(before time.Time) error
}

// HistoryStorage handles attack history storage
type HistoryStorage interface {
	// SaveAttackConfig saves an attack configuration
	SaveAttackConfig(config AttackConfig) error

	// SaveAttackResult saves an attack result
	SaveAttackResult(result AttackResult) error

	// GetAttackHistory retrieves attack history by UID
	GetAttackHistory(uid uint64) ([]AttackResult, error)

	// GetAllHistory retrieves all attack history
	GetAllHistory(limit int) ([]AttackResult, error)

	// Clear clears all history
	Clear() error
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

	// BeforePrepare is called before sending prepare message
	BeforePrepare(ctx context.Context, message *QBFTMessage) error

	// AfterPrepare is called after receiving prepare message
	AfterPrepare(ctx context.Context, message *QBFTMessage) error

	// BeforeCommit is called before sending commit message
	BeforeCommit(ctx context.Context, message *QBFTMessage) error

	// AfterCommit is called after receiving commit message
	AfterCommit(ctx context.Context, message *QBFTMessage) error

	// OnRoundChange is called on round change
	OnRoundChange(ctx context.Context, sequence, round uint64) error

	// OnMessageReceive is called when receiving any message
	OnMessageReceive(ctx context.Context, message *QBFTMessage) error
}

// ByzantineService is the main service interface
type ByzantineService interface {
	// Start starts the service
	Start() error

	// Stop stops the service
	Stop() error

	// GetStatus returns the service status
	GetStatus() ServiceStatus

	// Configure configures the service
	Configure(config ByzantineConfig) error

	// GetMetrics returns service metrics
	GetMetrics() Metrics

	// RegisterAttack registers a new attack
	RegisterAttack(config AttackConfig) (uint64, error)

	// CancelAttack cancels an attack
	CancelAttack(uid uint64) error

	// ListAttacks lists all attacks
	ListAttacks() []AttackConfig

	// GetAttackHistory gets attack history
	GetAttackHistory(uid uint64) ([]AttackResult, error)

	// GetAttackManager gets attack manager
	GetAttackManager() AttackManager

	// GetMessageStorage gets message storage
	GetMessageStorage() MessageStorage

	// GetHistoryStorage gets history storage
	GetHistoryStorage() HistoryStorage
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

// ConsensusHook represents consensus hook
type ConsensusHook interface {
	// BeforeBroadcast is called before broadcasting a message
	// Returns true if the message should be sent, false to drop it
	BeforeBroadcast(msgCode, sequence, round uint64, from common.Address) bool

	// BeforeProcessMessage is called before processing a received message
	// Returns true if the message should be processed, false to drop it
	BeforeProcessMessage(msgCode, sequence, round uint64, from common.Address) bool
}
