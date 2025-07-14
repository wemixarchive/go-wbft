package attacks

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// ReplayAttack implements replay attack
type ReplayAttack struct {
	*registry.BaseAttack
	useOriginalView bool
	targets         []common.Address
}

var _ (types.Attack) = (*ReplayAttack)(nil)

// NewReplayAttack creates a new replay attack
func NewReplayAttack(config types.AttackConfig) (*ReplayAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.ReplayAttackParams)

	attack := &ReplayAttack{
		BaseAttack:      registry.NewBaseAttack(config),
		useOriginalView: params.UseOriginalView,
		targets:         params.Targets,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *ReplayAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
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
		//log.Debug("this replay attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.ReplayAttackParams); ok {
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

// ReplayAttackFactory creates replay attacks
func ReplayAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewReplayAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeReplay, ReplayAttackFactory)
	if err != nil {
		panic(err)
	}
}
