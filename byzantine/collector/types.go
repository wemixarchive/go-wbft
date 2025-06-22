package collector

import (
	"github.com/ethereum/go-ethereum/common"
)

// EventType defines the type of QBFT event
type EventType int

const (
	EventTypeMessage EventType = iota
	EventTypeStateChange
	EventTypeBlock
)

// ConsensusState defines QBFT consensus states
type ConsensusState int

const (
	StateIdle ConsensusState = iota
	StatePreprepared
	StatePrepared
	StateCommitted
	StateRoundChange
)

// QBFTMessage represents a QBFT consensus message
type QBFTMessage struct {
	Code     uint64
	Sequence uint64
	Round    uint64
	Address  common.Address
	Data     []byte
}

// QBFTEvent represents an event in the QBFT consensus
type QBFTEvent struct {
	Type     EventType
	Sequence uint64
	Round    uint64

	// For message events
	Message *QBFTMessage

	// For state change events
	OldState ConsensusState
	NewState ConsensusState

	// For block events
	Block interface{} // Would be *types.Block in real implementation
}

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

// QBFTEventSource interface for subscribing to QBFT events
type QBFTEventSource interface {
	Subscribe(ch chan<- QBFTEvent) error
	Unsubscribe(ch chan<- QBFTEvent) error
}
