package attacks

import (
	"context"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// StoreMessage implements store message
type StoreMessage struct {
	*registry.BaseAttack
}

var _ (types.Attack) = (*StoreMessage)(nil)

// NewStoreMessage creates a new store message
func NewStoreMessage(config types.AttackConfig) (*StoreMessage, error) {

	attack := &StoreMessage{
		BaseAttack: registry.NewBaseAttack(config),
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *StoreMessage) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
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
		//log.Debug("this store message is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.StoreAttackParams); ok {
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

// StoreMessageFactory creates store attacks
func StoreMessageFactory(config types.AttackConfig) (types.Attack, error) {
	return NewStoreMessage(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeStoreMessage, StoreMessageFactory)
	if err != nil {
		panic(err)
	}
}
