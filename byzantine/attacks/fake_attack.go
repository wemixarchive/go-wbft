package attacks

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// FakeMessageAttack implements fake message attack
type FakeMessageAttack struct {
	*registry.BaseAttack
	fields []types.Field
	params *types.FakeAttackParams
}

var _ types.Attack = (*FakeMessageAttack)(nil)

// NewFakeMessageAttack creates a new fake message attack
func NewFakeMessageAttack(config types.AttackConfig) (*FakeMessageAttack, error) {
	paramRegistry := registry.NewParameterParserRegistry()

	// Parse parameters for the specific attack type
	parsedParams, err := paramRegistry.ParseParameters(config.Type, config.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	params := parsedParams.(*types.FakeAttackParams)

	attack := &FakeMessageAttack{
		BaseAttack: registry.NewBaseAttack(config),
		fields:     params.Fields,
		params:     params,
	}
	return attack, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *FakeMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check if attack is enabled
	if !config.Enabled {
		return false
	}

	// Check if a sequence is in range
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
		//log.Debug("this fake attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check a message type
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return false
	}

	// Check if the message code matches the attack parameters
	if params := a.GetParams(); params == nil {
		return false
	} else {
		if !params.HasMessageCode(msgEvent.MessageCode) {
			return false
		}
	}

	return true
}

// generateFakeMessage generates a fake message based on type
func (a *FakeMessageAttack) generateFakeMessage(messageCode types.MessageCode, event types.Event) ([]byte, error) {
	// This would generate appropriate fake messages based on the message type
	// For example,
	// - For RoundChange: generate a valid-looking round change message
	// - For PrePrepare: generate a fake proposal with non-existent transactions
	// - For Prepare/Commit: generate messages with valid signatures but wrong content

	switch messageCode {
	case types.MessageCodeRoundChange:
		return a.generateFakeRoundChange(event)
	case types.MessageCodePrePrepare:
		return a.generateFakeProposal(event)
	case types.MessageCodePrepare, types.MessageCodeCommit:
		return a.generateFakeVote(messageCode, event)
	default:
		return nil, fmt.Errorf("unsupported message type for fake attack: %v", messageCode)
	}
}

// generateFakeRoundChange generates a fake round change message
func (a *FakeMessageAttack) generateFakeRoundChange(event types.Event) ([]byte, error) {
	// Create a round change message that looks valid but isn't justified
	// This would include proper formatting and signatures
	fakeMsg := &types.WBFTMessage{
		Code:      types.MessageCodeRoundChange,
		Sequence:  event.Sequence,
		Round:     event.Round + 1, // Request next round
		Address:   common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Signature: []byte("fake_roundchange_signature"),
	}

	// Serialize the message (placeholder)
	return serializeQBFTMessage(fakeMsg), nil
}

// generateFakeProposal generates a fake proposal
func (a *FakeMessageAttack) generateFakeProposal(event types.Event) ([]byte, error) {
	// Create a proposal with fake transactions that don't exist in the pool
	// This would include a valid block structure but with invalid content
	return []byte("fake_proposal_with_invalid_transactions"), nil
}

// generateFakeVote generates a fake prepare or commit message
func (a *FakeMessageAttack) generateFakeVote(messageCode types.MessageCode, event types.Event) ([]byte, error) {
	fakeMsg := &types.WBFTMessage{
		Code:      messageCode,
		Sequence:  event.Sequence,
		Round:     event.Round,
		Address:   common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Signature: []byte("fake_vote_signature"),
	}

	if messageCode == types.MessageCodeCommit {
		fakeMsg.CommittedSeal = []byte("fake_committed_seal")
	}

	return serializeQBFTMessage(fakeMsg), nil
}

// sendMessage sends a message to targets
func (a *FakeMessageAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on the actual network layer
	// This would broadcast the fake message to specified targets
	return nil
}

// GetParams returns the Fake attack parameters
func (a *FakeMessageAttack) GetParams() *types.FakeAttackParams {
	return a.params
}

// serializeQBFTMessage serializes a WBFT message (placeholder)
func serializeQBFTMessage(msg *types.WBFTMessage) []byte {
	// Actual implementation would properly serialize the message
	return []byte(fmt.Sprintf("%+v", msg))
}

// FakeAttackFactory creates fake attacks
func FakeAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewFakeMessageAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeFakeMessage, FakeAttackFactory)
	if err != nil {
		panic(err)
	}
}
