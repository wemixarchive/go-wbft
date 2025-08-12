package attacks

import (
	"context"
	"fmt"
	"math"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// SetMessagePolicyAttack controls message sending per sequence/round for target nodes
type SetMessagePolicyAttack struct {
	*registry.BaseAttack
	code   types.MessageCode
	fields []types.Field
	params *types.MessagePolicyParams
}

var _ types.Attack = (*SetMessagePolicyAttack)(nil)

// NewSetMessagePolicyAttack creates a message control attack per sequence/round
func NewSetMessagePolicyAttack(config types.AttackConfig) (*SetMessagePolicyAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.MessagePolicyParams)

	attack := &SetMessagePolicyAttack{
		BaseAttack: registry.NewBaseAttack(config),
		code:       params.Code,
		fields:     params.Fields,
		params:     params,
	}

	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *SetMessagePolicyAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if sequence is in range
	if !config.IsInSequenceRange(event.Sequence) {
		//log.Debug("this policy attack is not matched ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check round - math.MaxUint64 means match all rounds
	if config.Round != math.MaxUint64 && event.Round != config.Round {
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

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.MessagePolicyParams); ok {
		return data.MessageCode == params.Code
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

	return data.MessageCode == attackCode
}

// MessagePolicyAttackFactory creates sequence/round-based message control attacks
func MessagPolicyAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewSetMessagePolicyAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeMessagePolicy, MessagPolicyAttackFactory)
	if err != nil {
		panic(err)
	}
}
