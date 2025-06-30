package attacks

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// SilentMessageAttack implements silent proposer attack
type SilentMessageAttack struct {
	*registry.BaseAttack
	code      types.MessageCode
	direction types.MessageDirection
	targets   []common.Address
}

var _ (types.Attack) = (*SilentMessageAttack)(nil)

// NewSilentProposerAttack creates a new silent proposer attack
func NewSilentMessageAttack(config types.AttackConfig) (*SilentMessageAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.SilentAttackParams)
	//log.Info("[byzantine] silent attack config : ", "params", params)

	attack := &SilentMessageAttack{
		BaseAttack: registry.NewBaseAttack(config),
		code:       params.Code,
		direction:  types.MessageDirection(params.Direction),
		targets:    params.Targets,
	}

	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *SilentMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		//log.Debug("this silent attack is not matched ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		log.Debug("this silent attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check if this is a message event
	data, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	if data.MessageCode == config.Parameters["code"] {
		return true
	}

	// Check if we should be silent for this message type
	return event.Type == types.EventTypeMessageSent || event.Type == types.EventTypeMessageReceived
}

// Execute performs the silent proposer attack
func (a *SilentMessageAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()

	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data type for silent attack")
	}

	shouldBlock := false
	eventDirection := "send"
	if dir, ok := event.Metadata["direction"].(string); ok {
		eventDirection = dir
	}

	switch a.direction {
	case 1: // Send only
		shouldBlock = eventDirection == types.DirectionSend
	case 2: // Receive only
		shouldBlock = eventDirection == types.DirectionReceive
	case 3: // Both
		shouldBlock = true
	}

	// Check if target addresses match (if specified)
	if shouldBlock && len(a.targets) > 0 {
		found := false
		for _, target := range a.targets {
			if target == msgEvent.From {
				found = true
				break
			}
		}
		shouldBlock = found
	}

	blockReason := ""
	if shouldBlock {
		blockReason = fmt.Sprintf("Silent attack: dropping %s message at sequence %d, round %d",
			eventDirection, event.Sequence, event.Round)
	}

	// For silent attack, we don't actually send anything
	// Instead, we drop/ignore the message

	result := &types.AttackResult{
		UID:          a.GetUID(),
		Success:      true,
		ExecutedAt:   time.Now(),
		Duration:     time.Since(startTime),
		BlockMessage: shouldBlock,
		BlockReason:  blockReason,
		Details: map[string]interface{}{
			"action":       "silent_attack_evaluated",
			"blocked":      shouldBlock,
			"event_type":   event.Type,
			"sequence":     event.Sequence,
			"round":        event.Round,
			"message_type": msgEvent.MessageCode,
			"direction":    a.direction,
			"targets":      len(a.targets),
		},
	}

	return result, nil
}

// SilentAttackFactory creates silent attacks
func SilentAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewSilentMessageAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeSilentMessage, SilentAttackFactory)
	if err != nil {
		panic(err)
	}
}
