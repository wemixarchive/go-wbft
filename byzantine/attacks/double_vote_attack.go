package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// DoubleVoteAttack implements double vote attack
type DoubleVoteAttack struct {
	*registry.BaseAttack
	targets          []common.Address
	withValidMessage bool
}

var _ (types.Attack) = (*DoubleVoteAttack)(nil)

// NewDoubleVoteAttack creates a new double vote attack
func NewDoubleVoteAttack(config types.AttackConfig) (*DoubleVoteAttack, error) {
	targets, err := registry.ParseTargets(config)
	if err != nil {
		return nil, err
	}

	return &DoubleVoteAttack{
		BaseAttack:       registry.NewBaseAttack(config),
		targets:          targets,
		withValidMessage: registry.GetBoolParameter(config, "withValidMessage", true),
	}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *DoubleVoteAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Check if this is the right message type to attack
	return messageEvent.MessageType&config.Code != 0
}

// Execute performs the double vote attack
func (a *DoubleVoteAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()

	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data type")
	}

	// Create tampered message
	tamperedMessage := a.createTamperedMessage(messageEvent)

	// Send tampered message first
	if err := a.sendMessage(tamperedMessage, a.targets); err != nil {
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
		time.Sleep(100 * time.Millisecond) // Small delay

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
			"tampered_message_sent": true,
			"valid_message_sent":    a.withValidMessage,
			"targets":               len(a.targets),
		},
	}, nil
}

// createTamperedMessage creates a tampered message
func (a *DoubleVoteAttack) createTamperedMessage(event *types.MessageEvent) []byte {
	// Implementation depends on actual message structure
	// This is a simplified version
	tamperedContent := make([]byte, len(event.Content))
	copy(tamperedContent, event.Content)

	// Modify some bytes to create invalid signature
	if len(tamperedContent) > 10 {
		tamperedContent[5] ^= 0xFF
	}

	return tamperedContent
}

// sendMessage sends a message to targets
func (a *DoubleVoteAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	// This is a placeholder
	return nil
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeDoubleVote, func(config types.AttackConfig) (types.Attack, error) {
		return NewDoubleVoteAttack(config)
	})
	if err != nil {
		panic(err)
	}
}
