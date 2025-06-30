package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// RoleSpoofedAttack implements role spoofed attack
type RoleSpoofedAttack struct {
	*registry.BaseAttack
	fakeMessage []byte
	targets     []common.Address
	nodeAddress common.Address
}

var _ (types.Attack) = (*RoleSpoofedAttack)(nil)

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
		BaseAttack:  registry.NewBaseAttack(config),
		fakeMessage: params.FakeMessage,
		targets:     params.Targets,
		//nodeAddress: params.nodeAddress,
	}
	return attack, nil
	//
	//targets, err := registry.ParseTargets(config)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Get fake message from parameters
	//var fakeMessage []byte
	//if msgParam, exists := config.Parameters["fakeMessage"]; exists {
	//	switch v := msgParam.(type) {
	//	case []byte:
	//		fakeMessage = v
	//	case string:
	//		fakeMessage = []byte(v)
	//	default:
	//		fakeMessage = nil
	//	}
	//}
	//
	//// Get node address from config
	//nodeAddress := common.HexToAddress("0x0000000000000000000000000000000000000001")
	//if addrParam, exists := config.Parameters["nodeAddress"]; exists {
	//	if addr, ok := addrParam.(string); ok {
	//		nodeAddress = common.HexToAddress(addr)
	//	}
	//}
	//
	//return &RoleSpoofedAttack{
	//	BaseAttack:  registry.NewBaseAttack(config),
	//	fakeMessage: fakeMessage,
	//	targets:     targets,
	//	nodeAddress: nodeAddress,
	//}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *RoleSpoofedAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		//log.Debug("this role spoofed attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Role spoofing can be triggered on various events
	switch event.Type {
	case types.EventTypeMessageSent, types.EventTypeRoundChange:
		return true
	default:
		return false
	}
}

// Execute performs the role spoofed attack
func (a *RoleSpoofedAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()
	config := a.GetConfig()

	// Generate spoofed message based on message type
	spoofedMessage, spoofedRole, err := a.createSpoofedMessage(config.Code, event)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send spoofed message
	if err := a.sendMessage(spoofedMessage, a.targets); err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	return &types.AttackResult{
		UID:        a.GetUID(),
		Success:    true,
		ExecutedAt: time.Now(),
		Duration:   time.Since(startTime),
		Details: map[string]interface{}{
			"message_type": config.Code,
			"spoofed_role": spoofedRole,
			"targets":      len(a.targets),
			"action":       "role_spoofed_message_sent",
		},
	}, nil
}

// createSpoofedMessage creates a message spoofing a different role
func (a *RoleSpoofedAttack) createSpoofedMessage(messageCode types.MessageCode, event types.Event) ([]byte, string, error) {
	switch messageCode {
	case types.MessageCodePrePrepare, types.MessageCodeRoundChangePrePrepare:
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
	// Create a PrePrepare message as if we were the proposer
	// In reality, we're not the designated proposer for this round

	if a.fakeMessage != nil {
		return a.fakeMessage, "proposer", nil
	}

	// Generate fake proposal
	fakeProposal := &types.QBFTMessage{
		Code:     types.MessageCodePrePrepare,
		Sequence: event.Sequence,
		Round:    event.Round,
		Address:  a.nodeAddress, // Our address, but we're not the proposer
		// Proposal would be included here
	}

	// Add fake block proposal
	// In real implementation, this would include a proper block

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
		Address:   a.nodeAddress, // Our address, but we might not be a validator
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
