package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// ConsensusHookImpl implements types.ConsensusHook
type ConsensusHookImpl struct {
	attackManager  types.AttackManager
	eventPublisher types.EventPublisher
	uidGenerator   types.UIDGenerator
	paramRegistry  *registry.ParameterParserRegistry

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
		paramRegistry:   registry.NewParameterParserRegistry(),
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
func (h *ConsensusHookImpl) BeforeBroadcast(msgCode, sequence, round uint64) types.AttackConfig {
	ctx := context.Background()

	// Map consensus message code to Byzantine message code
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)
	// TODO:
	// 1. check MessageCodeRoundChangePrePrepare
	// 2. check MessageCodePropagation

	// Get applicable attack types for this message
	attackTypes := h.getApplicableAttackTypes(byzantineCode, types.DirectionSend)

	// Check each attack type
	for _, attackType := range attackTypes {
		// Use new FindExecutableAttack method with message code
		attack, found := h.attackManager.FindExecutableAttack(attackType, sequence, round, byzantineCode)
		if !found {
			continue
		}

		config := attack.GetConfig()

		// Create event for condition checking
		event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, types.DirectionSend)

		// Check attack-specific execution condition
		if !attack.CheckExecuteCondition(ctx, event) {
			//log.Debug("[byzantine] Attack condition not met",
			//	"uid", config.UID,
			//	"type", attackType,
			//	"sequence", sequence,
			//	"msgCode", msgCode,
			//	"byzantineCode", byzantineCode)
			continue
		}

		// Special handling for silent attacks
		if attackType == types.AttackTypeSilentMessage {
			log.Debug("[byzantine]", "direction", config.Parameters["direction"])
			if !(config.Parameters["direction"] == 1 || config.Parameters["direction"] == 3) {
				return types.AttackConfig{}
			}

			log.Info("[byzantine] Blocking outbound message",
				"attack_uid", config.UID,
				"attack_type", attackType,
				"msgCode", msgCode,
				"byzantineCode", byzantineCode,
				"sequence", sequence,
				"round", round,
				"sequence_range", fmt.Sprintf("%d-%d", config.SequenceStart, config.SequenceEnd),
				"execution_count", config.ExecutionCount+1,
				"max_executions", config.MaxExecutions)

			// Update execution state
			if err := h.attackManager.MarkAttackExecuted(config.UID, sequence); err != nil {
				log.Error("[byzantine] Failed to mark attack executed", "error", err)
			}

			// Publish event asynchronously for monitoring
			//go h.publishEvent(event)

			return config // Block the message
		}

		// Double vote attacks (tamper)
		if attackType == types.AttackTypeTamperedMessage {
			log.Info("[byzantine] Double vote attack triggered",
				"attack_uid", config.UID,
				"msgCode", msgCode,
				"byzantineCode", byzantineCode,
				"sequence", sequence,
				"round", round,
				"sequence_range", fmt.Sprintf("%d-%d", config.SequenceStart, config.SequenceEnd),
				"execution_count", config.ExecutionCount+1,
				"max_executions", config.MaxExecutions)

			// Update execution state
			if err := h.attackManager.MarkAttackExecuted(config.UID, sequence); err != nil {
				log.Error("[byzantine] Failed to mark attack executed", "error", err)
			}

			return config
		}

		// For other attack types that might need execution
		if h.shouldExecuteAttack(attackType, types.DirectionSend) {
			decision := h.evaluateAttack(ctx, attack, event)
			if decision.ShouldAttack {
				log.Info("[byzantine] Attack execution decision",
					"attack_uid", config.UID,
					"attack_type", attackType,
					"msgCode", msgCode,
					"byzantineCode", byzantineCode,
					"sequence", sequence,
					"round", round,
					"execution_count", config.ExecutionCount+1,
					"max_executions", config.MaxExecutions)

				// Update execution state
				if err := h.attackManager.MarkAttackExecuted(config.UID, sequence); err != nil {
					log.Error("[byzantine] Failed to mark attack executed", "error", err)
				}

				return config
			}
		}
	}

	// Create general event for monitoring/logging
	//event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, from, types.DirectionSend)

	// Publish event asynchronously
	//go h.publishEvent(event)

	return types.AttackConfig{}
}

// BeforeProcessMessage is called before processing a received message
func (h *ConsensusHookImpl) BeforeProcessMessage(msgCode, sequence, round uint64) types.AttackConfig {
	ctx := context.Background()
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	// Get applicable attack types for receiving messages
	attackTypes := h.getApplicableAttackTypes(byzantineCode, types.DirectionReceive)

	// Check each attack type with lookup
	for _, attackType := range attackTypes {
		// Use new FindExecutableAttack method with message code
		attack, found := h.attackManager.FindExecutableAttack(attackType, sequence, round, byzantineCode)
		if !found {
			continue
		}

		// Check if attack is eligible
		config := attack.GetConfig()
		if !h.isAttackEligible(config) {
			continue
		}

		// Create event
		event := h.createEvent(types.EventTypeMessageReceived, msgCode, sequence, round, types.DirectionReceive)

		// Check execution condition
		if !attack.CheckExecuteCondition(ctx, event) {
			continue
		}

		switch attackType {
		case types.AttackTypeSilentMessage:
			if !(config.Parameters["direction"] == 2 || config.Parameters["direction"] == 3) {
				return types.AttackConfig{}
			}
			log.Info("[byzantine] Blocking inbound message",
				"attack_uid", config.UID,
				"attack_type", attackType,
				"msgCode", msgCode,
				"sequence", sequence,
				"round", round)

			// Mark attack as executed
			h.attackManager.MarkAttackExecuted(config.UID,
				sequence)

			// Publish event asynchronously
			//go h.publishEvent(event)

			return config // Block inbound message
		case types.AttackTypeOmitMessage:
		case types.AttackTypeTamperedMessage:
		case types.AttackTypeFakeMessage:
		case types.AttackTypeRoleSpoofed:
		case types.AttackTypeReplay:
		default:
			log.Error("[byzantine] unknown attack type", "attackType", attackType)
			return types.AttackConfig{}
		}

		// Evaluate other attack types
		//if h.shouldExecuteAttack(attackType, types.DirectionReceive) {
		//	decision := h.evaluateAttack(ctx, attack, event)
		//	if decision.ShouldAttack {
		//		log.Info("[byzantine] Blocking inbound message",
		//			"attack_uid", config.UID,
		//			"attack_type", attackType,
		//			"reason", decision.Reason)
		//
		//		h.attackManager.MarkAttackExecuted(config.UID, sequence)
		//
		//		// Publish event asynchronously
		//		//go h.publishEvent(event)
		//
		//		return config
		//	}
		//}
	}

	// Create event for monitoring
	//event := h.createEvent(types.EventTypeMessageReceived, msgCode, sequence, round, from, types.DirectionReceive)

	// Publish event asynchronously
	//go h.publishEvent(event)

	return types.AttackConfig{}
}

// DoubleVote is called before broadcasting a message
func (h *ConsensusHookImpl) DoubleVote(attackType types.AttackType, msgCode, sequence, round uint64) types.AttackConfig {
	ctx := context.Background()
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	// Use new FindExecutableAttack method with message code
	attack, found := h.attackManager.FindExecutableAttack(attackType, sequence, round, byzantineCode)
	if found {
		config := attack.GetConfig()
		event := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, types.DirectionSend)

		if attack.CheckExecuteCondition(ctx, event) {
			log.Info("[byzantine] Double vote attack triggered",
				"attack_uid", config.UID,
				"msgCode", msgCode,
				"byzantineCode", byzantineCode,
				"sequence", sequence,
				"round", round,
				"execution_count", config.ExecutionCount+1,
				"max_executions", config.MaxExecutions)

			// Update execution state
			if err := h.attackManager.MarkAttackExecuted(config.UID, sequence); err != nil {
				log.Error("[byzantine] Failed to mark attack executed", "error", err)
			}

			// Publish event asynchronously
			//go h.publishEvent(event)

			return config
		}
	}
	return types.AttackConfig{}
}

// GetExecutableAttacks iterates over every defined AttackType and returns
// a map keyed by AttackType containing attacks that are both eligible
// and ready to run for the given <msgCode, sequence, round>.
func (h *ConsensusHookImpl) GetExecutableAttacks(msgCode, sequence, round uint64) map[types.AttackType]*types.ExecutableAttack {
	ctx := context.Background()
	result := make(map[types.AttackType]*types.ExecutableAttack)

	// ───────────────────────────────────────────────────────────────
	//    Scan all attack types and pick those registered for this
	//    <sequence, round>. Skip anything that isn’t enabled or that
	//    fails execution‐condition checks.
	// ───────────────────────────────────────────────────────────────
	for _, at := range types.AllAttackTypes {

		uid := h.uidGenerator.Generate(at, sequence, round)
		attack, ok := h.attackManager.GetAttackByUID(uid)
		if !ok {
			//log.Warn("[byzantine] no attack scheduled for this UID", "uid", uid)
			continue
		}

		cfg := attack.GetConfig()
		if !h.isAttackEligible(cfg) {
			log.Warn("[byzantine]globally disabled, or not in active window", "uid", uid)
			continue
		}

		result[at] = &types.ExecutableAttack{
			Enabled: cfg.Enabled,
			Status:  cfg.Status,
		}
		err := h.extractParams(cfg, result[at])
		if err != nil {
			// log.Warn("[byzantine] code mismatch or param-parse error", "uid", uid)
			delete(result, at)
			continue
		}

		if !h.IsMessageCodeMatched(cfg, msgCode, result[at]) {
			delete(result, at)
			continue
		}

		// Create event
		evt := h.createEvent(types.EventTypeMessageSent, msgCode, sequence, round, types.DirectionSend)
		// Check execution condition
		if !attack.CheckExecuteCondition(ctx, evt) {
			delete(result, at)
			log.Warn("[byzantine] CheckExecuteCondition error", "uid", uid)
			continue
		}

		log.Info("[byzantine] executable attack found", "uid", uid, "msgCode", msgCode)

	}
	return result
}

// extractParams attempts to parse and assign the concrete param struct for the given cfg
// into the appropriate field of the provided ExecutableAttack object.
// Returns an error if param extraction fails or returns nil.
func (h *ConsensusHookImpl) extractParams(cfg types.AttackConfig, attacks *types.ExecutableAttack) error {
	// log.Info("[Byzantine] extractParams", "cfg", cfg)

	type extractorFunc func(types.AttackConfig) (interface{}, error)
	type assignFunc func(interface{})

	handlers := map[types.AttackType]struct {
		extract extractorFunc
		assign  assignFunc
		errMsg  string
	}{
		types.AttackTypeSilentMessage: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetSilentParams()
			},
			assign: func(v interface{}) {
				attacks.SilentParams = v.(*types.SilentAttackParams)
			},
			errMsg: "SilentParams is nil",
		},
		types.AttackTypeTamperedMessage: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetTamperParams()
			},
			assign: func(v interface{}) {
				attacks.TamperParams = v.(*types.TamperAttackParams)
			},
			errMsg: "TamperParams is nil",
		},
		types.AttackTypeFakeMessage: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetFakeParams()
			},
			assign: func(v interface{}) {
				attacks.FakeParams = v.(*types.FakeAttackParams)
			},
			errMsg: "FakeParams is nil",
		},
		types.AttackTypeOmitMessage: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetOmitParams()
			},
			assign: func(v interface{}) {
				attacks.OmitParams = v.(*types.OmitAttackParams)
			},
			errMsg: "OmitParams is nil",
		},
		types.AttackTypeRoleSpoofed: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetRoleSpoofParams()
			},
			assign: func(v interface{}) {
				attacks.RoleSpoofParams = v.(*types.RoleSpoofAttackParams)
			},
			errMsg: "RoleSpoofParams is nil",
		},
		types.AttackTypeReplay: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetReplayParams()
			},
			assign: func(v interface{}) {
				attacks.ReplayParams = v.(*types.ReplayAttackParams)
			},
			errMsg: "ReplayParams is nil",
		},
	}

	handler, ok := handlers[cfg.Type]
	if !ok {
		return fmt.Errorf("unknown attack type: %s", cfg.Type)
	}

	param, err := handler.extract(cfg)
	if err != nil {
		return err
	}
	if param == nil {
		return fmt.Errorf(handler.errMsg)
	}

	handler.assign(param)
	return nil
}

// IsMessageCodeMatched checks if the internal message code of the given ExecutableAttack
// matches the expected msgCode for the specified attack type.
func (h *ConsensusHookImpl) IsMessageCodeMatched(cfg types.AttackConfig, msgCode uint64, attacks *types.ExecutableAttack) bool {
	switch cfg.Type {
	case types.AttackTypeSilentMessage:
		return attacks.SilentParams != nil &&
			types.MessageCodeToQBFT[attacks.SilentParams.Code] == msgCode

	case types.AttackTypeTamperedMessage:
		return attacks.TamperParams != nil &&
			types.MessageCodeToQBFT[attacks.TamperParams.Code] == msgCode

	case types.AttackTypeFakeMessage:
		return attacks.FakeParams != nil &&
			types.MessageCodeToQBFT[attacks.FakeParams.Code] == msgCode

	case types.AttackTypeOmitMessage:
		return attacks.OmitParams != nil &&
			types.MessageCodeToQBFT[attacks.OmitParams.Code] == msgCode

	case types.AttackTypeRoleSpoofed:
		return attacks.RoleSpoofParams != nil &&
			types.MessageCodeToQBFT[attacks.RoleSpoofParams.Code] == msgCode

	case types.AttackTypeReplay:
		return attacks.ReplayParams != nil &&
			types.MessageCodeToQBFT[attacks.ReplayParams.Code] == msgCode
	}
	return false
}

// ShouldExecuteAttack API for consensus to check if attack should be executed
func (h *ConsensusHookImpl) ShouldExecuteAttack(attackType types.AttackType, msgCode, sequence, round uint64) (*types.AttackConfig, bool) {
	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	attack, found := h.attackManager.FindExecutableAttack(attackType, sequence, round, byzantineCode)
	if !found {
		return nil, false
	}

	config := attack.GetConfig()
	return &config, true
}

// GetAttackConfig implements new interface method
func (h *ConsensusHookImpl) GetAttackConfig(attackType types.AttackType, sequence, round uint64) (*types.AttackConfig, error) {
	attack, found := h.attackManager.FindAttackForExecution(attackType, sequence, round)
	if !found {
		return nil, fmt.Errorf("no executable attack found for type %s at seq %d round %d", attackType, sequence, round)
	}

	config := attack.GetConfig()

	if config.ParsedParameters == nil && h.paramRegistry != nil {
		parser, exists := h.paramRegistry.GetParser(config.Type)
		if exists {
			parsed, _ := parser.Parse(config.Parameters)
			config.ParsedParameters = parsed
		}
	}

	configCopy := config
	return &configCopy, nil
}

// MarkAttackExecuted implements new interface method
func (h *ConsensusHookImpl) MarkAttackExecuted(uid string,
	sequence uint64) error {
	return h.attackManager.MarkAttackExecuted(uid, sequence)
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
	typeStr := types.AttackTypeToString(ctx.MessageType)
	return []string{
		// Any round for same sequence
		typeStr + "-" + string(ctx.MessageCode) + "-" + string(ctx.Sequence) + "-*",
		// Any sequence and round
		typeStr + "-" + string(ctx.MessageCode) + "-*-*",
	}
}

// createEvent creates an event from consensus data
func (h *ConsensusHookImpl) createEvent(eventType types.EventType, msgCode uint64,
	sequence, round uint64, direction string) types.Event {

	byzantineCode := h.mapConsensusCodeToByzantine(msgCode)

	return types.Event{
		Type:      eventType,
		Sequence:  sequence,
		Round:     round,
		Timestamp: time.Now(),
		Data: &types.MessageEvent{
			MessageCode: byzantineCode,
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

func (h *ConsensusHookImpl) isAttackApplicable(config types.AttackConfig, msgCode types.MessageCode) bool {
	// Check if attack has code parameter
	if codeParam, ok := config.Parameters["code"]; ok {
		// Safe type conversion for code parameter
		var attackCode types.MessageCode
		switch v := codeParam.(type) {
		case float64:
			attackCode = types.MessageCode(v)
		case int:
			attackCode = types.MessageCode(v)
		case types.MessageCode:
			attackCode = v
		default:
			return false
		}

		// Check if code matches
		if attackCode != msgCode {
			return false
		}
	}

	return true
}

// BeforeBlockCommit is called before committing a block
// Allows modification of seals for omit attack
func (h *ConsensusHookImpl) BeforeBlockCommit(block interface{}, preparedSeals, committedSeals []interface{}) ([]interface{}, []interface{}, error) {
	// Extract block number from interface
	var blockNumber uint64
	var round uint64

	// Type assertion for block
	if proposal, ok := block.(interface{ Number() *big.Int }); ok {
		blockNumber = proposal.Number().Uint64()
	}

	// Try to find omit attack for propagation
	// Note: Using MessageCodePropagation (32) for block propagation/commit phase
	attack, found := h.attackManager.FindExecutableAttack(
		types.AttackTypeOmitMessage,
		blockNumber,
		round,
		types.MessageCodePropagation, // 32
	)

	if !found {
		// No omit attack, return original seals
		return preparedSeals, committedSeals, nil
	}

	config := attack.GetConfig()

	// Log attack detection
	log.Info("[byzantine] Detected omit attack for block commit",
		"attack_uid", config.UID,
		"block_number", blockNumber,
		"prepared_seals", len(preparedSeals),
		"committed_seals", len(committedSeals))

	// Check if this is an omit attack for propagation
	if config.ParsedParameters != nil {
		params, ok := config.ParsedParameters.(*types.OmitAttackParams)
		if !ok {
			return preparedSeals, committedSeals, nil
		}

		originalPreparedCount := len(preparedSeals)
		originalCommittedCount := len(committedSeals)

		// Apply omit based on cmd
		switch params.Cmd {
		case 1: // Omit prepare seals
			preparedSeals = h.omitSeals(preparedSeals, params.Cnt)
			log.Info("[byzantine] Omitted prepare seals before block commit",
				"attack_uid", config.UID,
				"original_count", originalPreparedCount,
				"remaining_count", len(preparedSeals),
				"cnt", params.Cnt)
		case 2: // Omit commit seals
			committedSeals = h.omitSeals(committedSeals, params.Cnt)
			log.Info("[byzantine] Omitted commit seals before block commit",
				"attack_uid", config.UID,
				"original_count", originalCommittedCount,
				"remaining_count", len(committedSeals),
				"cnt", params.Cnt)
		}

		// Mark attack as executed
		h.attackManager.MarkAttackExecuted(config.UID, blockNumber)
	}

	return preparedSeals, committedSeals, nil
}

// Helper function to omit seals
func (h *ConsensusHookImpl) omitSeals(seals []interface{}, cnt uint64) []interface{} {
	if len(seals) == 0 {
		return seals
	}

	if cnt == 0 {
		// Omit all
		return []interface{}{}
	}

	if int(cnt) >= len(seals) {
		// Omit all if count exceeds available seals
		return []interface{}{}
	}

	// Return seals with first 'cnt' items omitted
	return seals[cnt:]
}
