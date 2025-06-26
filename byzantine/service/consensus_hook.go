package service

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// ConsensusHookImpl implements types.ConsensusHook
type ConsensusHookImpl struct {
	attackManager  types.AttackManager
	eventPublisher types.EventPublisher
}

var _ types.ConsensusHook = (*ConsensusHookImpl)(nil)

// NewConsensusHook creates a new consensus hook
func NewConsensusHook(attackManager types.AttackManager, eventPublisher types.EventPublisher) types.ConsensusHook {
	return &ConsensusHookImpl{
		attackManager:  attackManager,
		eventPublisher: eventPublisher,
	}
}

// BeforeBroadcast is called before broadcasting a message
func (h *ConsensusHookImpl) BeforeBroadcast(msgCode uint64, sequence, round uint64, from common.Address) bool {
	ctx := context.Background()
	event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)

	// First, publish event for async processing (logging, monitoring, etc.)
	go func() {
		if err := h.eventPublisher.Publish(event); err != nil {
			log.Error("Failed to publish broadcast event", "err", err)
		}
		// Also process async attacks
		_ = h.attackManager.ProcessEventAsync(ctx, event)
	}()

	// Then, evaluate attacks that need immediate decision
	decision, err := h.attackManager.EvaluateAndExecuteAttacks(ctx, event)
	if err != nil {
		log.Error("Failed to evaluate attacks", "err", err)
		return true // On error, allow message to proceed
	}

	if decision.ShouldBlock {
		log.Info("Byzantine: Blocking outbound message",
			"attack_type", decision.AttackType,
			"attack_uid", decision.AttackUID,
			"reason", decision.Reason,
			"msgCode", msgCode,
			"sequence", sequence,
			"round", round)
		return false
	}

	return true
}

// BeforeProcessMessage is called before processing a received message
func (h *ConsensusHookImpl) BeforeProcessMessage(msgCode uint64, sequence, round uint64, from common.Address) bool {
	ctx := context.Background()
	event := h.createEvent(types.EventTypeMessageReceived, msgCode, sequence, round, from, "receive")

	// First, publish event for async processing
	go func() {
		if err := h.eventPublisher.Publish(event); err != nil {
			log.Error("Failed to publish receive event", "err", err)
		}
		// Also process async attacks
		_ = h.attackManager.ProcessEventAsync(ctx, event)
	}()

	// Then, evaluate attacks that need immediate decision
	decision, err := h.attackManager.EvaluateAndExecuteAttacks(ctx, event)
	if err != nil {
		log.Error("Failed to evaluate attacks", "err", err)
		return true // On error, allow message to proceed
	}

	if decision.ShouldBlock {
		log.Info("Byzantine: Blocking inbound message",
			"attack_type", decision.AttackType,
			"attack_uid", decision.AttackUID,
			"reason", decision.Reason,
			"msgCode", msgCode,
			"sequence", sequence,
			"round", round,
			"from", from)
		return false
	}

	return true
}

// createEvent creates an event from consensus data
func (h *ConsensusHookImpl) createEvent(eventType types.EventType, msgCode uint64,
	sequence, round uint64, from common.Address, direction string) types.Event {

	// Map consensus message code to Byzantine message code
	var byzantineCode types.MessageCode
	switch msgCode {
	case 0x12:
		byzantineCode = types.MessageCodePrePrepare
	case 0x13:
		byzantineCode = types.MessageCodePrepare
	case 0x14:
		byzantineCode = types.MessageCodeCommit
	case 0x15:
		byzantineCode = types.MessageCodeRoundChange
	default:
		byzantineCode = types.MessageCode(msgCode)
	}

	return types.Event{
		Type:      eventType,
		Sequence:  sequence,
		Round:     round,
		Timestamp: time.Now(),
		Data: &types.MessageEvent{
			MessageType: byzantineCode,
			From:        from,
		},
		Metadata: map[string]interface{}{
			"hook":           "ConsensusHook",
			"direction":      direction,
			"consensus_code": msgCode,
		},
	}
}
