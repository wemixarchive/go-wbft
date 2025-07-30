package attacks

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// ReplayAttack implements replay attack
type ReplayAttack struct {
	*registry.BaseAttack
	useOriginalView bool
	params          *types.ReplayAttackParams
}

var _ types.Attack = (*ReplayAttack)(nil)

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
		params:          params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *ReplayAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
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
		//log.Debug("this replay attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
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

// GetParams returns the Replay attack parameters
func (a *ReplayAttack) GetParams() *types.ReplayAttackParams {
	return a.params
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
