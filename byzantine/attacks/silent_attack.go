package attacks

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// SilentMessageAttack implements silent proposer attack
type SilentMessageAttack struct {
	*registry.BaseAttack
	code      types.MessageCode
	direction types.MessageDirection
	targets   []common.Address
	params    *types.SilentAttackParams
}

var _ types.Attack = (*SilentMessageAttack)(nil)

// NewSilentMessageAttack creates a new silent proposer attack
func NewSilentMessageAttack(config types.AttackConfig) (*SilentMessageAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.SilentAttackParams)

	attack := &SilentMessageAttack{
		BaseAttack: registry.NewBaseAttack(config),
		code:       params.Code,
		direction:  types.MessageDirection(params.Direction),
		targets:    params.Targets,
		params:     params,
	}

	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *SilentMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if sequence is in range
	if !config.IsInSequenceRange(event.Sequence) {
		//log.Debug("this silent attack is not matched ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check round (0 means any round)
	if config.Round != 0 && event.Round != config.Round {
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
	if params, ok := config.ParsedParameters.(*types.SilentAttackParams); ok {
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
