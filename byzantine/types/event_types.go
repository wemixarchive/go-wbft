package types

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EventType represents different types of events
type EventType string

const (
	EventTypeMessageReceived EventType = "message_received"
	EventTypeMessageSent     EventType = "message_sent"
	EventTypeRoundChange     EventType = "round_change"
	EventTypeProposalCreated EventType = "proposal_created"
	EventTypeBlockCommitted  EventType = "block_committed"
	EventTypeAttackExecuted  EventType = "attack_executed"
	EventTypeAttackFailed    EventType = "attack_failed"
)

// Event represents a consensus event
type Event struct {
	Type      EventType              `json:"type"`
	Sequence  uint64                 `json:"sequence"`
	Round     uint64                 `json:"round"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    common.Address         `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageEvent represents a message-related event
type MessageEvent struct {
	MessageCode MessageCode      `json:"message_code"`
	From        common.Address   `json:"from"`
	To          []common.Address `json:"to,omitempty"`
	Content     []byte           `json:"content"`
	Hash        common.Hash      `json:"hash"`
}

// BlockEvent represents a block-related event
type BlockEvent struct {
	Block      *types.Block     `json:"block"`
	Proposer   common.Address   `json:"proposer"`
	Validators []common.Address `json:"validators"`
}

// ConsensusState defines QBFT consensus states
type ConsensusState int

const (
	StateIdle ConsensusState = iota
	StatePreprepared
	StatePrepared
	StateCommitted
	StateRoundChange
)

// CollectedData represents data collected by the EventCollector
type CollectedData struct {
	Type     string
	Sequence uint64
	Round    uint64
	Data     interface{}
}

// StateTransition represents a consensus state transition
type StateTransition struct {
	Sequence uint64
	Round    uint64
	OldState ConsensusState
	NewState ConsensusState
}
