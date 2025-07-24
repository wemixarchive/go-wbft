package service

import (
	"context"
	"fmt"
	"sync"
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
	mu              sync.RWMutex
	attackTypeCache map[string][]types.AttackType

	callMetrics *CallMetrics // for debug
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
		callMetrics:     NewCallMetrics(),
	}
}

// GetExecutableAttacks iterates over every defined AttackType and returns
// a map keyed by AttackType containing attacks that are both eligible
// and ready to run for the given <msgCode, sequence, round>.
func (h *ConsensusHookImpl) GetExecutableAttacks(msgCode types.MessageCode, sequence, round uint64) map[types.AttackType]*types.ExecutableAttack {
	if h.callMetrics != nil {
		caller := GetCallerInfo()
		h.callMetrics.TrackCall(caller)
		log.Trace("[BYZ] GetExecutableAttacks called",
			"caller", caller,
			"msgCode", msgCode,
			"sequence", sequence,
			"round", round,
			"thread_id", GetGoroutineID())
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[types.AttackType]*types.ExecutableAttack)

	// ───────────────────────────────────────────────────────────────
	//    Scan all attack types and pick those registered for this
	//    <sequence, round>. Skip anything that isn’t enabled or that
	//    fails execution‐condition checks.
	// ───────────────────────────────────────────────────────────────
	for _, at := range types.AllAttackTypes {
		attack, ok := h.attackManager.FindExecutableAttack(at, sequence, round, msgCode)
		if !ok {
			continue
		}
		cfg := attack.GetConfig()
		if !h.isAttackEligible(cfg) {
			continue
		}

		result[at] = &types.ExecutableAttack{
			Enabled: cfg.Enabled,
			UID:     cfg.UID,
			NAME:    cfg.Name,
			Status:  cfg.Status,
		}
		err := h.extractParams(cfg, result[at])
		if err != nil {
			delete(result, at)
			continue
		}

		if !h.IsMessageCodeMatched(cfg, msgCode, result[at]) {
			delete(result, at)
			continue
		}
	}
	return result
}

// extractParams attempts to parse and assign the concrete param struct for the given cfg
// into the appropriate field of the provided ExecutableAttack object.
// Returns an error if param extraction fails or returns nil.
func (h *ConsensusHookImpl) extractParams(cfg types.AttackConfig, attacks *types.ExecutableAttack) error {
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
		types.AttackTypeStoreMessage: {
			extract: func(cfg types.AttackConfig) (interface{}, error) {
				return cfg.GetStoreParams()
			},
			assign: func(v interface{}) {
				attacks.StoreMessageParams = v.(*types.StoreAttackParams)
			},
			errMsg: "StoreMessageParams is nil",
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
func (h *ConsensusHookImpl) IsMessageCodeMatched(cfg types.AttackConfig, msgCode types.MessageCode, attacks *types.ExecutableAttack) bool {
	switch cfg.Type {
	case types.AttackTypeSilentMessage:
		return attacks.SilentParams != nil && attacks.SilentParams.HasMessageCode(msgCode)

	case types.AttackTypeTamperedMessage:
		return attacks.TamperParams != nil && attacks.TamperParams.HasMessageCode(msgCode)

	case types.AttackTypeFakeMessage:
		return attacks.FakeParams != nil && attacks.FakeParams.HasMessageCode(msgCode)

	case types.AttackTypeOmitMessage:
		return attacks.OmitParams != nil && attacks.OmitParams.HasMessageCode(msgCode)

	case types.AttackTypeRoleSpoofed:
		return attacks.RoleSpoofParams != nil && attacks.RoleSpoofParams.HasMessageCode(msgCode)

	case types.AttackTypeReplay:
		return attacks.ReplayParams != nil && attacks.ReplayParams.HasMessageCode(msgCode)

	case types.AttackTypeStoreMessage:
		return attacks.StoreMessageParams != nil && attacks.StoreMessageParams.HasMessageCode(msgCode)
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
func (h *ConsensusHookImpl) MarkAttackExecuted(uid string, sequence uint64) error {
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
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeFakeMessage, types.AttackTypeStoreMessage)
	case types.MessageCodePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage, types.AttackTypeStoreMessage)
	case types.MessageCodeCommit:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage, types.AttackTypeStoreMessage)
	case types.MessageCodeRoundChange:
		attackTypes = append(attackTypes, types.AttackTypeFakeMessage, types.AttackTypeStoreMessage)
	}

	// Add direction-specific types
	if direction == types.DirectionSend {
		attackTypes = append(attackTypes, types.AttackTypeRoleSpoofed)
	} else {
		attackTypes = append(attackTypes, types.AttackTypeReplay)
	}

	// Cache the result
	h.mu.Lock()
	h.attackTypeCache[cacheKey] = attackTypes
	h.mu.Unlock()

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
func (h *ConsensusHookImpl) createEvent(eventType types.EventType, msgCode,
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
			"hook":      "ConsensusHook",
			"direction": direction,
			"msg_code":  msgCode,
		},
	}
}

// publishEvent publishes an event asynchronously
func (h *ConsensusHookImpl) publishEvent(event types.Event) {
	if h.eventPublisher != nil {
		if err := h.eventPublisher.Publish(event); err != nil {
			log.Error("[BYZ] Failed to publish event",
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
			log.Error("[BYZ] Failed to process event asynchronously", "error", err)
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

func (h *ConsensusHookImpl) GetCallMetrics() map[string]CallStat {
	return h.callMetrics.GetStats()
}
