package adapter

import (
	"context"
	"sync"
	"time"

	btypes "github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/core/types"
)

// HookAdapter implements the HookAdapter interface
type HookAdapter struct {
	mu               sync.RWMutex
	byzantineService btypes.ByzantineService
	eventPublisher   btypes.EventPublisher
	messageStorage   btypes.MessageStorage
}

// NewHookAdapter creates a new hook adapter
func NewHookAdapter(
	service btypes.ByzantineService,
	publisher btypes.EventPublisher,
	storage btypes.MessageStorage,
) *HookAdapter {
	return &HookAdapter{
		byzantineService: service,
		eventPublisher:   publisher,
		messageStorage:   storage,
	}
}

// BeforeProposal is called before creating a proposal
func (h *HookAdapter) BeforeProposal(ctx context.Context, sequence, round uint64) error {
	event := btypes.Event{
		Type:      btypes.EventTypeProposalCreated,
		Sequence:  sequence,
		Round:     round,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"hook": "BeforeProposal",
		},
	}

	return h.eventPublisher.Publish(event)
}

// AfterProposal is called after creating a proposal
func (h *HookAdapter) AfterProposal(ctx context.Context, proposal *types.Block) error {
	event := btypes.Event{
		Type:      btypes.EventTypeProposalCreated,
		Sequence:  proposal.NumberU64(),
		Timestamp: time.Now(),
		Data: &btypes.BlockEvent{
			Block: proposal,
		},
		Metadata: map[string]interface{}{
			"hook": "AfterProposal",
		},
	}

	return h.eventPublisher.Publish(event)
}

// BeforePrepare is called before sending prepare message
func (h *HookAdapter) BeforePrepare(ctx context.Context, message *btypes.QBFTMessage) error {
	event := btypes.Event{
		Type:      btypes.EventTypeMessageSent,
		Sequence:  message.Sequence,
		Round:     message.Round,
		Timestamp: time.Now(),
		Data: &btypes.MessageEvent{
			MessageType: btypes.MessageCodePrepare,
			From:        message.Address,
			Content:     message.Signature,
			Hash:        message.Hash,
		},
		Metadata: map[string]interface{}{
			"hook": "BeforePrepare",
		},
	}

	return h.eventPublisher.Publish(event)
}

// AfterPrepare is called after receiving prepare message
func (h *HookAdapter) AfterPrepare(ctx context.Context, message *btypes.QBFTMessage) error {
	// Store the message
	if h.messageStorage != nil {
		storedMsg := &btypes.StoredMessage{
			Message:    message,
			ReceivedAt: time.Now(),
			FromPeer:   message.Address,
		}
		_ = h.messageStorage.Store(storedMsg)
	}

	event := btypes.Event{
		Type:      btypes.EventTypeMessageReceived,
		Sequence:  message.Sequence,
		Round:     message.Round,
		Timestamp: time.Now(),
		Data: &btypes.MessageEvent{
			MessageType: btypes.MessageCodePrepare,
			From:        message.Address,
			Content:     message.Signature,
			Hash:        message.Hash,
		},
		Metadata: map[string]interface{}{
			"hook": "AfterPrepare",
		},
	}

	return h.eventPublisher.Publish(event)
}

// BeforeCommit is called before sending commit message
func (h *HookAdapter) BeforeCommit(ctx context.Context, message *btypes.QBFTMessage) error {
	event := btypes.Event{
		Type:      btypes.EventTypeMessageSent,
		Sequence:  message.Sequence,
		Round:     message.Round,
		Timestamp: time.Now(),
		Data: &btypes.MessageEvent{
			MessageType: btypes.MessageCodeCommit,
			From:        message.Address,
			Content:     message.CommittedSeal,
			Hash:        message.Hash,
		},
		Metadata: map[string]interface{}{
			"hook": "BeforeCommit",
		},
	}

	return h.eventPublisher.Publish(event)
}

// AfterCommit is called after receiving commit message
func (h *HookAdapter) AfterCommit(ctx context.Context, message *btypes.QBFTMessage) error {
	// Store the message
	if h.messageStorage != nil {
		storedMsg := &btypes.StoredMessage{
			Message:    message,
			ReceivedAt: time.Now(),
			FromPeer:   message.Address,
		}
		_ = h.messageStorage.Store(storedMsg)
	}

	event := btypes.Event{
		Type:      btypes.EventTypeMessageReceived,
		Sequence:  message.Sequence,
		Round:     message.Round,
		Timestamp: time.Now(),
		Data: &btypes.MessageEvent{
			MessageType: btypes.MessageCodeCommit,
			From:        message.Address,
			Content:     message.CommittedSeal,
			Hash:        message.Hash,
		},
		Metadata: map[string]interface{}{
			"hook": "AfterCommit",
		},
	}

	return h.eventPublisher.Publish(event)
}

// OnRoundChange is called on round change
func (h *HookAdapter) OnRoundChange(ctx context.Context, sequence, round uint64) error {
	event := btypes.Event{
		Type:      btypes.EventTypeRoundChange,
		Sequence:  sequence,
		Round:     round,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"hook": "OnRoundChange",
		},
	}

	return h.eventPublisher.Publish(event)
}

// OnMessageReceive is called when receiving any message
func (h *HookAdapter) OnMessageReceive(ctx context.Context, message *btypes.QBFTMessage) error {
	// Store the message
	if h.messageStorage != nil {
		storedMsg := &btypes.StoredMessage{
			Message:    message,
			ReceivedAt: time.Now(),
			FromPeer:   message.Address,
		}
		_ = h.messageStorage.Store(storedMsg)
	}

	event := btypes.Event{
		Type:      btypes.EventTypeMessageReceived,
		Sequence:  message.Sequence,
		Round:     message.Round,
		Timestamp: time.Now(),
		Data: &btypes.MessageEvent{
			MessageType: message.Code,
			From:        message.Address,
			Hash:        message.Hash,
		},
		Metadata: map[string]interface{}{
			"hook": "OnMessageReceive",
		},
	}

	return h.eventPublisher.Publish(event)
}
