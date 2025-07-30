package attacks

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// DosAttack represents a DoS attack that floods messages
type DosAttack struct {
	*registry.BaseAttack
	params *types.DosAttackParams
}

var _ types.Attack = (*DosAttack)(nil)

// NewDosAttack creates a new DoS attack instance
func NewDosAttack(config types.AttackConfig) (*DosAttack, error) {
	// Parse parameters for the specific attack type
	paramRegistry := registry.NewParameterParserRegistry()
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DoS attack parameters: %w", err)
	}

	params := parsedParams.(*types.DosAttackParams)

	// Initialize attack
	attack := &DosAttack{
		BaseAttack: registry.NewBaseAttack(config),
		params:     params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the DoS attack should be executed
func (a *DosAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if attack is enabled
	if !config.Enabled {
		return false
	}

	// Check sequence range
	if !config.IsInSequenceRange(event.Sequence) {
		return false
	}

	// Check round
	if event.Round != config.Round {
		return false
	}

	// Check if attack is already completed or cancelled
	status := a.GetStatus()
	if status == types.AttackStatusCompleted || status == types.AttackStatusCancelled {
		return false
	}

	// Check message code
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

// GetParams returns the DoS attack parameters
func (a *DosAttack) GetParams() *types.DosAttackParams {
	return a.params
}

// DosAttackFactory Factory function for creating DoS attacks
func DosAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewDosAttack(config)
}

// Register the DoS attack factory
func init() {
	err := registry.Register(types.AttackTypeDos, DosAttackFactory)
	if err != nil {
		panic(err)
	}
}
