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
	uidGenerator   types.UIDGenerator

	// Cache for attack type mapping to improve performance
	attackTypeCache map[string][]types.AttackType
}

var _ types.ConsensusHook = (*ConsensusHookImpl)(nil)

// NewConsensusHook creates a new consensus hook
func NewConsensusHook(attackManager types.AttackManager, eventPublisher types.EventPublisher) types.ConsensusHook {
	return &ConsensusHookImpl{
		attackManager:   attackManager,
		eventPublisher:  eventPublisher,
		uidGenerator:    types.NewUIDGenerator(),
		attackTypeCache: make(map[string][]types.AttackType),
	}
}

// CheckAttackCondition checks if the attack condition is met
func (h *ConsensusHookImpl) CheckAttackCondition(ctx types.ConsensusContext) bool {
	// generate uid
	uid := h.uidGenerator.Generate(
		ctx.MessageType,
		ctx.Sequence,
		ctx.Round,
	)

	log.Debug("[byzantine] Checking attack condition with UID", "uid", uid)

	if attack, exists := h.attackManager.GetAttackByUID(uid); exists {
		config := attack.GetConfig()

		// Check if attack is eligible
		if h.isAttackEligible(config) {
			log.Info("[byzantine] Attack condition met",
				"uid", uid,
				"name", config.Name,
				"type", config.Type,
				"status", config.Status)
			return true
		}
	}

	// If no exact match, check for wildcard patterns
	// This is still optimized as we only check relevant patterns
	patterns := h.generateWildcardPatterns(ctx)
	for _, pattern := range patterns {
		if attack, exists := h.attackManager.GetAttackByUID(pattern); exists {
			config := attack.GetConfig()
			if h.isAttackEligible(config) {
				log.Info("[Byzantine] Wildcard attack condition met",
					"pattern", pattern,
					"name", config.Name,
					"type", config.Type)
				return true
			}
		}
	}

	return false
}

// BeforeBroadcast is called before broadcasting a message
func (h *ConsensusHookImpl) BeforeBroadcast(msgCode, sequence, round uint64, from common.Address) bool {
	ctx := context.Background()

	// Map consensus message code to Byzantine message code
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)
	log.Info("[byzantine] BeforeBroadcast", "code", byzantineCode)

	// NEW: Generate UIDs for all possible attack types that could apply
	attackTypes := h.getApplicableAttackTypes(byzantineCode, types.DirectionSend)
	log.Info("[byzantine] applicable attack type", "attack type", attackTypes)

	// Check each attack type
	for _, attackType := range attackTypes {
		uid := h.uidGenerator.Generate(attackType, sequence, round)
		log.Info("[byzantine] BeforeBroadcast", "uid", uid)

		// lookup
		attack, exists := h.attackManager.GetAttackByUID(uid)
		if !exists {
			continue
		}
		log.Info("[byzantine] exist attack : ", "uid", uid)

		// Check if attack is eligible
		config := attack.GetConfig()
		if !h.isAttackEligible(config) {
			continue
		}

		// Create event for condition checking
		event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)

		// Check execution condition
		if !attack.CheckExecuteCondition(ctx, event) {
			continue
		}

		// For silent attacks, we can immediately decide to block
		if attackType == types.AttackTypeSilentMessage {
			log.Info("[byzantine] Blocking outbound message lookup",
				"attack_uid", uid,
				"attack_type", attackType,
				"msgCode", msgCode,
				"sequence", sequence,
				"round", round)

			// Update attack status
			h.attackManager.UpdateStatusMap(attack, types.AttackStatusExecuted)

			// Publish event asynchronously for monitoring
			//go h.publishEvent(event)

			return true // Block the message
		}

		// For other attack types that might need execution
		if h.shouldExecuteAttack(attackType, types.DirectionSend) {
			decision := h.evaluateAttack(ctx, attack, event)
			if decision.ShouldAttack {
				log.Info("[byzantine] Attack execution decision",
					"attack_uid", uid,
					"attack_type", attackType,
					"block", decision.ShouldAttack)

				// Publish event asynchronously
				//go h.publishEvent(event)

				return decision.ShouldAttack
			}
		}
	}

	// Create general event for monitoring/logging
	//event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)

	// Publish event asynchronously
	//go h.publishEvent(event)

	return false
}

// BeforeProcessMessage is called before processing a received message
func (h *ConsensusHookImpl) BeforeProcessMessage(msgCode, sequence, round uint64, from common.Address) bool {
	ctx := context.Background()
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	// Get applicable attack types for receiving messages
	attackTypes := h.getApplicableAttackTypes(byzantineCode, types.DirectionReceive)

	// Check each attack type with lookup
	for _, attackType := range attackTypes {
		uid := h.uidGenerator.Generate(attackType, sequence, round)

		// lookup
		attack, exists := h.attackManager.GetAttackByUID(uid)
		if !exists {
			continue
		}

		// Check if attack is eligible
		config := attack.GetConfig()
		if !h.isAttackEligible(config) {
			continue
		}

		// Create event
		event := h.createEvent(types.EventTypeMessageReceived, msgCode, sequence, round, from, types.DirectionReceive)

		// Check execution condition
		if !attack.CheckExecuteCondition(ctx, event) {
			continue
		}

		// For silent attacks on receive
		if attackType == types.AttackTypeSilentMessage {
			log.Info("[byzantine] Blocking inbound message",
				"attack_uid", uid,
				"attack_type", attackType,
				"msgCode", msgCode,
				"sequence", sequence,
				"round", round,
				"from", from)

			// Update attack status
			h.attackManager.UpdateStatusMap(attack, types.AttackStatusExecuted)

			// Publish event asynchronously
			//go h.publishEvent(event)

			return false // Block the message (return false for receive)
		}

		// Evaluate other attack types
		if h.shouldExecuteAttack(attackType, types.DirectionReceive) {
			decision := h.evaluateAttack(ctx, attack, event)
			if decision.ShouldAttack {
				log.Info("[byzantine] Blocking inbound message",
					"attack_uid", uid,
					"attack_type", attackType,
					"reason", decision.Reason)

				// Publish event asynchronously
				//go h.publishEvent(event)

				return false // Block inbound message
			}
		}
	}

	// Create event for monitoring
	//event := h.createEvent(types.EventTypeMessageReceived, msgCode, sequence, round, from, types.DirectionReceive)

	// Publish event asynchronously
	//go h.publishEvent(event)

	return true // Allow processing
}

// DoubleVote is called before broadcasting a message
func (h *ConsensusHookImpl) DoubleVote(msgCode, sequence, round uint64, from common.Address) bool {
	ctx := context.Background()
	//byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	// Direct lookup for tamper attack
	uid := h.uidGenerator.Generate(types.AttackTypeTamperedMessage, sequence, round)

	attack, exists := h.attackManager.GetAttackByUID(uid)
	if exists {
		config := attack.GetConfig()
		if h.isAttackEligible(config) {
			// Create event
			event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)

			// Check execution condition
			if attack.CheckExecuteCondition(ctx, event) {
				log.Info("[byzantine] Double vote attack triggered",
					"attack_uid", uid,
					"msgCode", msgCode,
					"sequence", sequence,
					"round", round)

				// Update attack status
				h.attackManager.UpdateStatusMap(attack, types.AttackStatusExecuted)

				// Publish event asynchronously
				//go h.publishEvent(event)

				return true
			}
		}
	}

	// Fallback to general evaluation if needed
	event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)
	decision, err := h.attackManager.EvaluateAndExecuteAttacks(ctx, event)
	if err != nil {
		log.Error("Failed to evaluate attacks for double vote", "error", err)
		return false
	}

	if decision.ShouldAttack && decision.AttackType == types.AttackTypeTamperedMessage {
		return true
	}

	return false
}

// Helper methods

// mapConsensusCodeToByzantine maps consensus message code to Byzantine message code
func (h *ConsensusHookImpl) mapConsensusCodeToByzantine(msgCode uint64) types.MessageCode {
	switch msgCode {
	case 0x12:
		return types.MessageCodePrePrepare
	case 0x13:
		return types.MessageCodePrepare
	case 0x14:
		return types.MessageCodeCommit
	case 0x15:
		return types.MessageCodeRoundChange
	default:
		return types.MessageCode(msgCode)
	}
}

// getApplicableAttackTypes returns cached attack types for a message and direction
func (h *ConsensusHookImpl) getApplicableAttackTypes(msgCode types.MessageCode, direction string) []types.AttackType {
	// Create cache key
	cacheKey := string(msgCode) + "-" + direction

	// Check cache first
	if cached, exists := h.attackTypeCache[cacheKey]; exists {
		return cached
	}

	// Build attack type list
	attackTypes := []types.AttackType{
		types.AttackTypeSilentMessage, // Always applicable
	}

	// Add message-specific attack types
	switch msgCode {
	case types.MessageCodePrePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeFakeMessage)
	case types.MessageCodePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage)
	case types.MessageCodeCommit:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage)
	case types.MessageCodeRoundChange:
		attackTypes = append(attackTypes, types.AttackTypeFakeMessage)
	}

	// Add direction-specific types
	if direction == types.DirectionSend {
		attackTypes = append(attackTypes, types.AttackTypeRoleSpoofed)
	} else {
		attackTypes = append(attackTypes, types.AttackTypeReplay)
	}

	// Cache the result
	h.attackTypeCache[cacheKey] = attackTypes

	return attackTypes
}

// isAttackEligible checks if an attack configuration is eligible for execution
func (h *ConsensusHookImpl) isAttackEligible(config types.AttackConfig) bool {
	return config.Enabled &&
		config.Status != types.AttackStatusCompleted &&
		config.Status != types.AttackStatusCancelled &&
		config.Status != types.AttackStatusFailed
}

// shouldExecuteAttack determines if an attack type should be executed for a direction
func (h *ConsensusHookImpl) shouldExecuteAttack(attackType types.AttackType, direction string) bool {
	// Silent attacks are handled separately for immediate blocking
	if attackType == types.AttackTypeSilentMessage {
		return false
	}

	// Tamper and fake attacks on send
	if direction == types.DirectionSend &&
		(attackType == types.AttackTypeTamperedMessage ||
			attackType == types.AttackTypeFakeMessage ||
			attackType == types.AttackTypeRoleSpoofed) {
		return true
	}

	// Replay attacks on receive
	if direction == types.DirectionReceive && attackType == types.AttackTypeReplay {
		return true
	}

	// Omit attacks can apply to both directions
	if attackType == types.AttackTypeOmitMessage {
		return true
	}

	return false
}

// evaluateAttack evaluates a single attack
func (h *ConsensusHookImpl) evaluateAttack(ctx context.Context, attack types.Attack, event types.Event) types.AttackDecision {
	result, err := attack.Execute(ctx, event)
	if err != nil {
		log.Error("Failed to execute attack",
			"uid", attack.GetUID(),
			"error", err)
		return types.AttackDecision{
			ShouldAttack: false,
			Reason:       "Attack execution failed",
		}
	}

	if result != nil && result.BlockMessage {
		return types.AttackDecision{
			ShouldAttack: true,
			AttackUID:    attack.GetUID(),
			AttackType:   attack.GetType(),
			Reason:       result.BlockReason,
			Result:       result,
		}
	}

	return types.AttackDecision{
		ShouldAttack: false,
		Reason:       "Attack executed but no blocking required",
	}
}

// generateWildcardPatterns generates wildcard patterns for flexible matching
func (h *ConsensusHookImpl) generateWildcardPatterns(ctx types.ConsensusContext) []string {
	// For now, we'll generate basic wildcard patterns
	// This can be extended based on requirements
	typeStr := types.AttachTypeToString(ctx.MessageType)
	return []string{
		// Any round for same sequence
		typeStr + "-" + string(ctx.MessageCode) + "-" + string(ctx.Sequence) + "-*",
		// Any sequence and round
		typeStr + "-" + string(ctx.MessageCode) + "-*-*",
	}
}

// createEvent creates an event from consensus data
func (h *ConsensusHookImpl) createEvent(eventType types.EventType, msgCode uint64,
	sequence, round uint64, from common.Address, direction string) types.Event {

	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

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

// publishEvent publishes an event asynchronously
func (h *ConsensusHookImpl) publishEvent(event types.Event) {
	if h.eventPublisher != nil {
		if err := h.eventPublisher.Publish(event); err != nil {
			log.Error("[byzantine] Failed to publish event",
				"type", event.Type,
				"sequence", event.Sequence,
				"round", event.Round,
				"error", err)
		}
	}

	// Also trigger async processing if needed
	if h.attackManager != nil {
		ctx := context.Background()
		if err := h.attackManager.ProcessEventAsync(ctx, event); err != nil {
			log.Error("[byzantine] Failed to process event asynchronously", "error", err)
		}
	}
}
