package attacks

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// RoleSpoofedAttack implements role spoofed attack
type RoleSpoofedAttack struct {
	*registry.BaseAttack
	fields  []types.Field
	targets []common.Address
	params  *types.RoleSpoofAttackParams

	mu sync.RWMutex
}

var _ types.Attack = (*RoleSpoofedAttack)(nil)

// NewRoleSpoofedAttack creates a new role spoofed attack
func NewRoleSpoofedAttack(config types.AttackConfig) (*RoleSpoofedAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.RoleSpoofAttackParams)

	attack := &RoleSpoofedAttack{
		BaseAttack: registry.NewBaseAttack(config),
		fields:     params.Fields,
		targets:    params.Targets,
		params:     params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *RoleSpoofedAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
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
		//log.Debug("this role spoofed attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check message type
	messageEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Use ParsedParameters first
	if params, ok := config.ParsedParameters.(*types.RoleSpoofAttackParams); ok {
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

// createSpoofedMessage creates a message spoofing a different role
func (a *RoleSpoofedAttack) createSpoofedMessage(messageCode types.MessageCode, event types.Event) ([]byte, string, error) {
	switch messageCode {
	case types.MessageCodePrePrepare:
		// Spoof proposer role
		return a.spoofProposerMessage(event)
	case types.MessageCodePrepare, types.MessageCodeCommit:
		// Spoof validator role (if not a validator)
		return a.spoofValidatorMessage(messageCode, event)
	default:
		return nil, "", fmt.Errorf("unsupported message type for role spoofing: %v", messageCode)
	}
}

// spoofProposerMessage creates a PrePrepare message while not being the proposer
func (a *RoleSpoofedAttack) spoofProposerMessage(event types.Event) ([]byte, string, error) {
	// This method handles both regular PrePrepare and PrePrepare after round change
	// The context (presence of RCMessages) is checked in the consensus layer

	log.Info("[BYZ] Spoofing PrePrepare message",
		"sequence", event.Sequence,
		"round", event.Round)

	if a.fields != nil {
		return []byte(fmt.Sprintf("%+v", a.fields)), "proposer", nil
	}

	// Generate fake proposal
	fakeProposal := &types.QBFTMessage{
		Code:     types.MessageCodePrePrepare,
		Sequence: event.Sequence,
		Round:    event.Round,
		Address:  common.HexToAddress("0x0000000000000000000000000000000000000001"),
		// In real implementation, this would include:
		// - Block proposal
		// - RoundChange justification if round > 0
		// - Prepare messages if applicable
	}

	return serializeQBFTMessage(fakeProposal), "proposer", nil
}

// spoofValidatorMessage creates a Prepare/Commit message while not being a validator
func (a *RoleSpoofedAttack) spoofValidatorMessage(messageCode types.MessageCode, event types.Event) ([]byte, string, error) {
	// Create a vote message as if we were a validator
	// In reality, we might not be in the validator set

	fakeVote := &types.QBFTMessage{
		Code:      messageCode,
		Sequence:  event.Sequence,
		Round:     event.Round,
		Address:   common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Signature: []byte("spoofed_validator_signature"),
	}

	if messageCode == types.MessageCodeCommit {
		fakeVote.CommittedSeal = []byte("spoofed_commit_seal")
	}

	return serializeQBFTMessage(fakeVote), "validator", nil
}

// sendMessage sends a message to targets
func (a *RoleSpoofedAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	// This would send the message as if from the spoofed role
	return nil
}

// RoleSpoofAttackFactory creates role spoof attacks
func RoleSpoofAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewRoleSpoofedAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeRoleSpoofed, RoleSpoofAttackFactory)
	if err != nil {
		panic(err)
	}
}
