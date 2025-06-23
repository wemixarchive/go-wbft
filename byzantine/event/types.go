package event

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// EventType defines the type of QBFT event
type EventType int

const (
	EventTypeMessage EventType = iota
	EventTypeStateChange
	EventTypeBlock
)

// WBFTMessage represents a QBFT consensus message
type WBFTMessage struct {
	Code     uint64
	Sequence uint64
	Round    uint64
	Address  common.Address
	Data     []byte
}

// WBFTEvent represents an event in the QBFT consensus
type WBFTEvent struct {
	Type     EventType
	Sequence uint64
	Round    uint64

	// For message events
	Message *WBFTMessage

	// For state change events
	OldState types.ConsensusState
	NewState types.ConsensusState

	// For block events
	Block interface{} // Would be *types.Block in real implementation
}

// WBFTEventSource interface for subscribing to QBFT events
type WBFTEventSource interface {
	Subscribe(ch chan<- WBFTEvent) error
	Unsubscribe(ch chan<- WBFTEvent) error
}
