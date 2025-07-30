package attacks

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// StoreMessage implements store message
type StoreMessage struct {
	*registry.BaseAttack
	params *types.StoreAttackParams
}

var _ types.Attack = (*StoreMessage)(nil)

// NewStoreMessage creates a new store message
func NewStoreMessage(config types.AttackConfig) (*StoreMessage, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.StoreAttackParams)

	attack := &StoreMessage{
		BaseAttack: registry.NewBaseAttack(config),
		params:     params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *StoreMessage) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if attack is enabled
	if !config.Enabled {
		return false
	}

	// Check if sequence is in range
	if !config.IsInSequenceRange(event.Sequence) {
		return false
	}

	// Check round
	if event.Round != config.Round {
		return false
	}

	// Check attack status
	status := a.GetStatus()
	if status == types.AttackStatusCancelled || status == types.AttackStatusCompleted {
		//log.Debug("this store message is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Check if message code matches
	if params := a.GetParams(); params == nil {
		return false
	} else {
		if !params.HasMessageCode(msgEvent.MessageCode) {
			return false
		}
	}

	return true
}

// GetParams returns the store message parameters
func (a *StoreMessage) GetParams() *types.StoreAttackParams {
	return a.params
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
