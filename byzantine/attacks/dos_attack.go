package attacks

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
)

// DosAttack represents a DoS attack that floods messages
type DosAttack struct {
	*registry.BaseAttack
	fields         []types.Field
	targets        []common.Address
	params         *types.DosAttackParams
	messageBuffer  []interface{} // Buffer for storing generated messages
	executionState *DosExecutionState
}

var _ types.Attack = (*DosAttack)(nil)

// DosExecutionState tracks the state of ongoing DoS attack
type DosExecutionState struct {
	TotalMessagesSent int
	LastExecutionTime time.Time
	CurrentBatch      int
}

// NewDosAttack creates a new DoS attack instance
func NewDosAttack(config types.AttackConfig) (*DosAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DoS attack parameters: %w", err)
	}

	params := parsedParams.(*types.DosAttackParams)

	// Initialize attack
	attack := &DosAttack{
		BaseAttack:    registry.NewBaseAttack(config),
		fields:        params.Fields,
		targets:       params.Targets,
		params:        params,
		messageBuffer: make([]interface{}, 0),
		executionState: &DosExecutionState{
			TotalMessagesSent: 0,
			LastExecutionTime: time.Time{},
			CurrentBatch:      0,
		},
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
	if !a.params.HasMessageCode(msgEvent.MessageCode) {
		return false
	}

	return true
}

// GetParams returns the DoS attack parameters
func (a *DosAttack) GetParams() *types.DosAttackParams {
	return a.params
}

// GetExecutionState returns the current execution state
func (a *DosAttack) GetExecutionState() *DosExecutionState {
	return a.executionState
}

// UpdateExecutionState updates the execution state after sending messages
func (a *DosAttack) UpdateExecutionState(messagesSent int) {
	a.executionState.TotalMessagesSent += messagesSent
	a.executionState.LastExecutionTime = time.Now()
	a.executionState.CurrentBatch++
}

// GetMessageBuffer returns the message buffer
func (a *DosAttack) GetMessageBuffer() []interface{} {
	return a.messageBuffer
}

// SetMessageBuffer sets the message buffer
func (a *DosAttack) SetMessageBuffer(buffer []interface{}) {
	a.messageBuffer = buffer
}

// Factory function for creating DoS attacks
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
