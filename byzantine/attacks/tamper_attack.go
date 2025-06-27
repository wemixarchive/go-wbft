package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// TamperedMessageAttack implements tampered message attack
type TamperedMessageAttack struct {
	*registry.BaseAttack
	tamperFields     []types.TamperField
	withValidMessage bool
	delay            time.Duration
	targets          []common.Address
}

var _ (types.Attack) = (*TamperedMessageAttack)(nil)

// NewTamperedMessageAttack creates a new tampered message attack
func NewTamperedMessageAttack(config types.AttackConfig) (*TamperedMessageAttack, error) {
	targets, err := registry.ParseTargets(config)
	if err != nil {
		return nil, err
	}

	// Parse tamper fields
	tamperFieldsRaw, ok := config.Parameters["tamperFields"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("tamperFields parameter required")
	}

	tamperFields := make([]types.TamperField, len(tamperFieldsRaw))
	for i, field := range tamperFieldsRaw {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid tamper field format")
		}

		tamperFields[i] = types.TamperField{
			Target: fieldMap["target"].(string),
			Value:  fieldMap["value"],
		}
	}

	delay := time.Duration(registry.GetUint64Parameter(config, "delay", 0)) * time.Millisecond

	return &TamperedMessageAttack{
		BaseAttack:       registry.NewBaseAttack(config),
		tamperFields:     tamperFields,
		withValidMessage: registry.GetBoolParameter(config, "withValidMessage", false),
		delay:            delay,
		targets:          targets,
	}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *TamperedMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		//log.Debug("this tamper attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}
	
	// Check if this is the right message type to attack
	if event.Type == types.EventTypeMessageSent {
		return true
	}

	return messageEvent.MessageType&config.Code != 0
}

// Execute performs the tampered message attack
func (a *TamperedMessageAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()

	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data type")
	}

	// Apply tampering
	tamperedContent, err := a.applyTampering(messageEvent.Content)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send tampered message
	if err := a.sendMessage(tamperedContent, a.targets); err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send valid message if configured
	if a.withValidMessage {
		time.Sleep(a.delay)

		if err := a.sendMessage(messageEvent.Content, a.targets); err != nil {
			return &types.AttackResult{
				UID:        a.GetUID(),
				Success:    false,
				Error:      err,
				ExecutedAt: time.Now(),
				Duration:   time.Since(startTime),
			}, err
		}
	}

	return &types.AttackResult{
		UID:        a.GetUID(),
		Success:    true,
		ExecutedAt: time.Now(),
		Duration:   time.Since(startTime),
		Details: map[string]interface{}{
			"fields_tampered":    len(a.tamperFields),
			"valid_message_sent": a.withValidMessage,
			"targets":            len(a.targets),
		},
	}, nil
}

// applyTampering applies tampering to message content
func (a *TamperedMessageAttack) applyTampering(content []byte) ([]byte, error) {
	// This is a simplified implementation
	// Actual implementation would parse the message and modify specific fields

	tamperedContent := make([]byte, len(content))
	copy(tamperedContent, content)

	// Apply each tamper field
	for _, field := range a.tamperFields {
		// Implementation depends on actual message structure
		// This is just a placeholder
		switch field.Target {
		case "Proposal.Header.Coinbase":
			// Modify coinbase field
			if len(tamperedContent) > 20 {
				copy(tamperedContent[10:30], field.Value.([]byte))
			}
		case "Proposal.Header.Number":
			// Modify block number
			if len(tamperedContent) > 40 {
				// Convert value to bytes and copy
			}
			// Add more cases as needed
		}
	}

	return tamperedContent, nil
}

// sendMessage sends a message to targets
func (a *TamperedMessageAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	return nil
}

// TamperAttackFactory creates tamper attacks
func TamperAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewTamperedMessageAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeTamperedMessage, TamperAttackFactory)
	if err != nil {
		panic(err)
	}
}
