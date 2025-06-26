package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// FakeMessageAttack implements fake message attack
type FakeMessageAttack struct {
	*registry.BaseAttack
	fakeMessage []byte
	targets     []common.Address
}

var _ (types.Attack) = (*FakeMessageAttack)(nil)

// NewFakeMessageAttack creates a new fake message attack
func NewFakeMessageAttack(config types.AttackConfig) (*FakeMessageAttack, error) {
	targets, err := registry.ParseTargets(config)
	if err != nil {
		return nil, err
	}

	// Get fake message from parameters
	var fakeMessage []byte
	if msgParam, exists := config.Parameters["fakeMessage"]; exists {
		switch v := msgParam.(type) {
		case []byte:
			fakeMessage = v
		case string:
			fakeMessage = []byte(v)
		default:
			// If not provided, we'll generate it during execution
			fakeMessage = nil
		}
	}

	return &FakeMessageAttack{
		BaseAttack:  registry.NewBaseAttack(config),
		fakeMessage: fakeMessage,
		targets:     targets,
	}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *FakeMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		log.Debug("this fake attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// For fake message, we can trigger on various events
	switch event.Type {
	case types.EventTypeMessageSent, types.EventTypeRoundChange, types.EventTypeProposalCreated:
		return true
	default:
		return false
	}
}

// Execute performs the fake message attack
func (a *FakeMessageAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()
	config := a.GetConfig()

	// Generate or use provided fake message
	var messageToSend []byte
	if a.fakeMessage != nil {
		messageToSend = a.fakeMessage
	} else {
		// Generate fake message based on message type
		var err error
		messageToSend, err = a.generateFakeMessage(config.Code, event)
		if err != nil {
			return &types.AttackResult{
				UID:        a.GetUID(),
				Success:    false,
				Error:      err,
				ExecutedAt: time.Now(),
				Duration:   time.Since(startTime),
			}, err
		}
	}

	// Send fake message
	if err := a.sendMessage(messageToSend, a.targets); err != nil {
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
			"message_type":      config.Code,
			"fake_message_size": len(messageToSend),
			"targets":           len(a.targets),
			"action":            "fake_message_sent",
		},
	}, nil
}

// generateFakeMessage generates a fake message based on type
func (a *FakeMessageAttack) generateFakeMessage(messageCode types.MessageCode, event types.Event) ([]byte, error) {
	// This would generate appropriate fake messages based on the message type
	// For example:
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
	fakeMsg := &types.QBFTMessage{
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
	fakeMsg := &types.QBFTMessage{
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
	// Implementation depends on actual network layer
	// This would broadcast the fake message to specified targets
	return nil
}

// serializeQBFTMessage serializes a QBFT message (placeholder)
func serializeQBFTMessage(msg *types.QBFTMessage) []byte {
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
