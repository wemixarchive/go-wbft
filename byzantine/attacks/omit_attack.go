package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// OmitMessageAttack implements omit message attack
type OmitMessageAttack struct {
	*registry.BaseAttack
	cmd     uint64
	targets []common.Address
}

var _ (types.Attack) = (*OmitMessageAttack)(nil)

// NewOmitMessageAttack creates a new omit message attack
func NewOmitMessageAttack(config types.AttackConfig) (*OmitMessageAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.OmitAttackParams)

	attack := &OmitMessageAttack{
		BaseAttack: registry.NewBaseAttack(config),
		cmd:        params.Cmd,
		targets:    params.Targets,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *OmitMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if sequence is in range
	if !config.IsInSequenceRange(event.Sequence) {
		return false
	}

	// Check round (0 means any round)
	if config.Round != 0 && event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		//log.Debug("this omit attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.OmitAttackParams); ok {
		return messageEvent.MessageCode == params.Code
	}

	// Fallback to Parameters map
	var attackCode types.MessageCode
	switch v := config.Parameters["code"].(type) {
	case float64:
		attackCode = types.MessageCode(v)
	case int:
		attackCode = types.MessageCode(v)
	default:
		return false
	}

	return messageEvent.MessageCode == attackCode
}

// Execute performs the omit message attack
func (a *OmitMessageAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()
	config := a.GetConfig()

	// Determine what to omit based on message code and command
	// Safe type conversion for code parameter
	var messageCode types.MessageCode
	switch v := config.Parameters["code"].(type) {
	case float64:
		messageCode = types.MessageCode(v)
	case int:
		messageCode = types.MessageCode(v)
	case types.MessageCode:
		messageCode = v
	default:
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      fmt.Errorf("invalid message code type: %T", config.Parameters["code"]),
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, nil
	}
	omittedMessage, err := a.createOmittedMessage(messageCode, event)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send message with omitted fields
	if err := a.sendMessage(omittedMessage, a.targets); err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	return &types.AttackResult{
		UID:        a.GetUID(),
		Success:    true,
		ExecutedAt: time.Now(),
		Duration:   time.Since(startTime),
		Details: map[string]interface{}{
			"message_code": config.Parameters["code"],
			"omit_command": a.cmd,
			"targets":      len(a.targets),
			"action":       "omitted_message_sent",
		},
	}, nil
}

// createOmittedMessage creates a message with omitted fields
func (a *OmitMessageAttack) createOmittedMessage(messageCode types.MessageCode, event types.Event) ([]byte, error) {
	switch messageCode {
	case types.MessageCodePrePrepare:
		return a.omitPrePrepareFields(event)
	case types.MessageCodePropagation:
		return a.omitPropagationFields(event)
	//case types.MessageCodeRoundChangePrePrepare:
	//	return a.omitRoundChangePrePrepareFields(event)
	default:
		return nil, fmt.Errorf("unsupported message type for omit attack: %v", messageCode)
	}
}

// omitPrePrepareFields omits fields from PrePrepare message
func (a *OmitMessageAttack) omitPrePrepareFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit prev Prepare Seal
	// 2: omit prev Commit Seal

	// For PrePrepare, the event would contain a block proposal
	// Since we're creating an attack, we need to create a modified message

	log.Info("[BYZ] Omitting fields from PrePrepare",
		"cmd", a.cmd,
		"sequence", event.Sequence,
		"round", event.Round)

	// TODO: In real implementation, this would:
	// 1. Get the current block proposal from consensus layer
	// 2. Extract WBFT extra data
	// 3. Modify seals based on cmd
	// 4. Re-encode and send modified PrePrepare

	// For now, return the original content as we need consensus layer integration
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data for PrePrepare omit")
	}

	return msgEvent.Content, nil
}

// omitPropagationFields omits fields from block propagation
func (a *OmitMessageAttack) omitPropagationFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit Prepare Seal
	// 2: omit Commit Seal

	log.Info("[BYZ] Omitting fields from Propagation",
		"cmd", a.cmd,
		"sequence", event.Sequence)

	// TODO: In real implementation, this would:
	// 1. Get the finalized block from consensus layer
	// 2. Extract and modify WBFT extra data
	// 3. Remove prepare/commit seals based on cmd
	// 4. Re-encode and propagate modified block

	// For now, return the original content
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data for Propagation omit")
	}

	return msgEvent.Content, nil
}

// omitRoundChangePrePrepareFields omits fields from RoundChange-PrePrepare
func (a *OmitMessageAttack) omitRoundChangePrePrepareFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit RoundChangeMessages
	// 2: omit PrepareMessages

	log.Info("[BYZ] Omitting justification from RoundChange-PrePrepare",
		"cmd", a.cmd,
		"sequence", event.Sequence,
		"round", event.Round)

	// TODO: In real implementation, this would:
	// 1. Get the PrePrepare message after round change
	// 2. Remove justification messages based on cmd:
	//    - cmd=1: Remove RoundChange messages
	//    - cmd=2: Remove Prepare messages
	// 3. Re-encode and send modified message

	// For now, return the original content
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data for RoundChange-PrePrepare omit")
	}

	return msgEvent.Content, nil
}

// sendMessage sends a message to targets
func (a *OmitMessageAttack) sendMessage(content []byte, targets []common.Address) error {
	// TODO: Implement actual message sending through consensus layer
	// This would involve:
	// 1. Getting the backend/broadcaster interface
	// 2. Sending the message to specified targets or all validators
	log.Info("[BYZ] Sending omitted message",
		"targets", len(targets),
		"content_size", len(content))
	return nil
}

// OmitAttackFactory creates omit attacks
func OmitAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewOmitMessageAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeOmitMessage, OmitAttackFactory)
	if err != nil {
		panic(err)
	}
}
