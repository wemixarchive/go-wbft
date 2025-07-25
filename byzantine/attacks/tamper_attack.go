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
	tamperFields     []types.Field
	withValidMessage bool
	delay            time.Duration
	targets          []common.Address
	params           *types.TamperAttackParams
}

var _ types.Attack = (*TamperedMessageAttack)(nil)

// NewTamperedMessageAttack creates a new tampered message attack
func NewTamperedMessageAttack(config types.AttackConfig) (*TamperedMessageAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.TamperAttackParams)

	attack := &TamperedMessageAttack{
		BaseAttack:       registry.NewBaseAttack(config),
		tamperFields:     params.Fields,
		withValidMessage: params.WithValidMessage,
		delay:            time.Duration(int64(params.Delay)),
		targets:          params.Targets,
		params:           params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *TamperedMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if sequence is in range
	if !config.IsInSequenceRange(event.Sequence) {
		return false
	}

	// Check round
	if event.Round != config.Round {
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

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.TamperAttackParams); ok {
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
		case "Header.Coinbase":
			// Modify coinbase field
			if len(tamperedContent) > 20 {
				copy(tamperedContent[10:30], field.Value.([]byte))
			}
		case "Header.Number":
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
