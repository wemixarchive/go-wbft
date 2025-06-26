package attacks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	coretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// OmitMessageAttack implements omit message attack
type OmitMessageAttack struct {
	*registry.BaseAttack
	omitCommand uint64
	omitCount   uint64
	targets     []common.Address
}

var _ (types.Attack) = (*OmitMessageAttack)(nil)

// NewOmitMessageAttack creates a new omit message attack
func NewOmitMessageAttack(config types.AttackConfig) (*OmitMessageAttack, error) {
	targets, err := registry.ParseTargets(config)
	if err != nil {
		return nil, err
	}

	return &OmitMessageAttack{
		BaseAttack:  registry.NewBaseAttack(config),
		omitCommand: registry.GetUint64Parameter(config, "cmd", 0),
		omitCount:   registry.GetUint64Parameter(config, "cnt", 0),
		targets:     targets,
	}, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *OmitMessageAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		log.Debug("this omit attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check if this is the right message type to attack
	if event.Type == types.EventTypeMessageSent || event.Type == types.EventTypeBlockCommitted {
		return true
	}

	return false
}

// Execute performs the omit message attack
func (a *OmitMessageAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()
	config := a.GetConfig()

	// Determine what to omit based on message code and command
	omittedMessage, err := a.createOmittedMessage(config.Code, event)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send message with omitted fields
	if err := a.sendMessage(omittedMessage, a.targets); err != nil {
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
			"omit_command": a.omitCommand,
			"omit_count":   a.omitCount,
			"targets":      len(a.targets),
			"action":       "omitted_message_sent",
		},
	}, nil
}

// createOmittedMessage creates a message with omitted fields
func (a *OmitMessageAttack) createOmittedMessage(messageCode types.MessageCode, event types.Event) ([]byte, error) {
	switch messageCode {
	case types.MessageCodePrePrepare:
		return a.omitPrePrepareFields(event)
	case types.MessageCodePropagation:
		return a.omitPropagationFields(event)
	case types.MessageCodeRoundChangePrePrepare:
		return a.omitRoundChangePrePrepareFields(event)
	default:
		return nil, fmt.Errorf("unsupported message type for omit attack: %v", messageCode)
	}
}

// omitPrePrepareFields omits fields from PrePrepare message
func (a *OmitMessageAttack) omitPrePrepareFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit prev Prepare Seal
	// 2: omit prev Commit Seal

	blockEvent, ok := event.Data.(*types.BlockEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data for PrePrepare omit")
	}

	// Create a proposal with omitted seals
	proposal := blockEvent.Block

	switch a.omitCommand {
	case 1:
		// Omit prepare seals
		// Remove prepare seals from extra data
		proposal = a.removePrepareSeal(proposal, a.omitCount)
	case 2:
		// Omit commit seals
		// Remove commit seals from extra data
		proposal = a.removeCommitSeal(proposal, a.omitCount)
	default:
		return nil, fmt.Errorf("invalid omit command: %d", a.omitCommand)
	}

	// Serialize the modified proposal
	return serializeProposal(proposal), nil
}

// omitPropagationFields omits fields from block propagation
func (a *OmitMessageAttack) omitPropagationFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit Prepare Seal
	// 2: omit Commit Seal

	blockEvent, ok := event.Data.(*types.BlockEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data for Propagation omit")
	}

	block := blockEvent.Block

	switch a.omitCommand {
	case 1:
		// Omit current prepare seals
		block = a.removePrepareSeal(block, a.omitCount)
	case 2:
		// Omit current commit seals
		block = a.removeCommitSeal(block, a.omitCount)
	default:
		return nil, fmt.Errorf("invalid omit command: %d", a.omitCommand)
	}

	return serializeBlock(block), nil
}

// omitRoundChangePrePrepareFields omits fields from RoundChange-PrePrepare
func (a *OmitMessageAttack) omitRoundChangePrePrepareFields(event types.Event) ([]byte, error) {
	// Based on omitCommand:
	// 1: omit RoundChangeMessages
	// 2: omit PrepareMessages

	// Create a PrePrepare message with omitted justification
	message := &types.QBFTMessage{
		Code:     types.MessageCodePrePrepare,
		Sequence: event.Sequence,
		Round:    event.Round,
		Address:  common.HexToAddress("0x0000000000000000000000000000000000000001"),
	}

	// Normally would include justification, but we're omitting it
	switch a.omitCommand {
	case 1:
		// Omit RoundChangeMessages (but might include PrepareMessages)
		// This creates an unjustified PrePrepare after round change
	case 2:
		// Omit PrepareMessages (but might include RoundChangeMessages)
		// This creates a PrePrepare without proper prepare justification
	default:
		return nil, fmt.Errorf("invalid omit command: %d", a.omitCommand)
	}

	return serializeQBFTMessage(message), nil
}

// Helper functions to remove seals
func (a *OmitMessageAttack) removePrepareSeal(block *coretypes.Block, count uint64) *coretypes.Block {
	// Implementation would modify the block's extra data to remove prepare seals
	// If count is 0, remove all; otherwise remove specified number
	return block
}

func (a *OmitMessageAttack) removeCommitSeal(block *coretypes.Block, count uint64) *coretypes.Block {
	// Implementation would modify the block's extra data to remove commit seals
	// If count is 0, remove all; otherwise remove specified number
	return block
}

// sendMessage sends a message to targets
func (a *OmitMessageAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	return nil
}

// Serialization helpers (placeholders)
func serializeProposal(proposal *coretypes.Block) []byte {
	return []byte("serialized_proposal_with_omissions")
}

func serializeBlock(block *coretypes.Block) []byte {
	return []byte("serialized_block_with_omissions")
}

// OmitAttackFactory creates omit attacks
func OmitAttackFactory(config types.AttackConfig) (types.Attack, error) {
	return NewOmitMessageAttack(config)
}

// Register the attack
func init() {
	err := registry.Register(types.AttackTypeOmitMessage, OmitAttackFactory)
	if err != nil {
		panic(err)
	}
}
