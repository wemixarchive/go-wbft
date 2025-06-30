package attacks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
)

// StorageProvider provides access to message storage
type StorageProvider interface {
	GetMessageStorage() types.MessageStorage
}

// Global storage provider
var (
	storageProviderMu sync.RWMutex
	storageProvider   StorageProvider
)

// SetStorageProvider sets the global storage provider
func SetStorageProvider(provider StorageProvider) {
	storageProviderMu.Lock()
	defer storageProviderMu.Unlock()
	storageProvider = provider
}

// GetStorageProvider gets the global storage provider
func GetStorageProvider() StorageProvider {
	storageProviderMu.RLock()
	defer storageProviderMu.RUnlock()
	return storageProvider
}

// ReplayAttack implements replay attack
type ReplayAttack struct {
	*registry.BaseAttack
	oriSequence     uint64
	oriRound        uint64
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
		oriSequence:     params.OriSequence,
		oriRound:        params.OriRound,
		useOriginalView: params.UseOriginalView,
		targets:         params.Targets,
	}
	return attack, nil

	//targets, err := registry.ParseTargets(config)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return &ReplayAttack{
	//	BaseAttack:       registry.NewBaseAttack(config),
	//	originalSequence: registry.GetUint64Parameter(config, "originalSequence", 0),
	//	originalRound:    registry.GetUint64Parameter(config, "originalRound", 0),
	//	useOriginalView:  registry.GetBoolParameter(config, "useOriginalView", false),
	//	targets:          targets,
	//}, nil
}

// getMessageStorage gets message storage from provider
func (a *ReplayAttack) getMessageStorage() (types.MessageStorage, error) {
	provider := GetStorageProvider()
	if provider == nil {
		return nil, fmt.Errorf("storage provider not initialized")
	}

	storage := provider.GetMessageStorage()
	if storage == nil {
		return nil, fmt.Errorf("message storage not available")
	}

	return storage, nil
}

// CheckExecuteCondition checks if the attack should be executed
func (a *ReplayAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	config := a.GetConfig()

	// Check sequence and round
	if event.Sequence != config.Sequence || event.Round != config.Round {
		return false
	}

	// Check attack status
	if config.Status == types.AttackStatusCancelled || config.Status == types.AttackStatusCompleted {
		//log.Debug("this replay attack is already cancelled or completed ", "sequence", event.Sequence, "round", event.Round)
		return false
	}

	// Check if storage is available
	if _, err := a.getMessageStorage(); err != nil {
		return false
	}

	// Check if this is the right event type
	return event.Type == types.EventTypeMessageSent
}

// Execute performs the replay attack
func (a *ReplayAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()

	// Get message storage
	messageStorage, err := a.getMessageStorage()
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Retrieve original message from storage
	messages, err := messageStorage.GetBySequenceRound(a.oriSequence, a.oriRound)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	if len(messages) == 0 {
		err := fmt.Errorf("no messages found for sequence %d round %d", a.oriSequence, a.oriRound)
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Select the best message to replay based on message type
	messageToReplay := a.selectBestMessage(messages, a.GetConfig().Code)
	if messageToReplay == nil {
		err := fmt.Errorf("no suitable message found for replay")
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Prepare replay message
	replayContent, err := a.prepareReplayMessage(messageToReplay.Message, event.Sequence, event.Round)
	if err != nil {
		return &types.AttackResult{
			UID:        a.GetUID(),
			Success:    false,
			Error:      err,
			ExecutedAt: time.Now(),
			Duration:   time.Since(startTime),
		}, err
	}

	// Send replayed message
	if err := a.sendMessage(replayContent, a.targets); err != nil {
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
			"original_sequence": a.oriSequence,
			"original_round":    a.oriRound,
			"replayed_to":       fmt.Sprintf("seq:%d,round:%d", event.Sequence, event.Round),
			"use_original_view": a.useOriginalView,
			"message_type":      messageToReplay.Message.Code,
			"targets":           len(a.targets),
		},
	}, nil
}

// selectBestMessage selects the most appropriate message for replay
func (a *ReplayAttack) selectBestMessage(messages []*types.StoredMessage, targetCode types.MessageCode) *types.StoredMessage {
	// First, try to find exact match
	for _, msg := range messages {
		if msg.Message.Code == targetCode {
			return msg
		}
	}

	// If no exact match, return the first message
	if len(messages) > 0 {
		return messages[0]
	}

	return nil
}

// prepareReplayMessage prepares a message for replay
func (a *ReplayAttack) prepareReplayMessage(original *types.QBFTMessage, newSequence, newRound uint64) ([]byte, error) {
	// Clone the original message
	replayed := &types.QBFTMessage{
		Code:          original.Code,
		Sequence:      original.Sequence,
		Round:         original.Round,
		Address:       original.Address,
		Signature:     make([]byte, len(original.Signature)),
		CommittedSeal: make([]byte, len(original.CommittedSeal)),
		Proposal:      original.Proposal,
		Hash:          original.Hash,
	}

	// Copy slices to avoid modifying original
	copy(replayed.Signature, original.Signature)
	copy(replayed.CommittedSeal, original.CommittedSeal)

	// Update sequence and round if not using original view
	if !a.useOriginalView {
		replayed.Sequence = newSequence
		replayed.Round = newRound

		// Recalculate hash for the modified message
		replayed.Hash = a.calculateMessageHash(replayed)
	}

	// Serialize the message
	return a.serializeMessage(replayed)
}

// calculateMessageHash calculates hash for a message
func (a *ReplayAttack) calculateMessageHash(msg *types.QBFTMessage) common.Hash {
	// Implementation would calculate proper hash based on message fields
	// This is a simplified version
	data := fmt.Sprintf("%d-%d-%s-%d", msg.Sequence, msg.Round, msg.Address.Hex(), msg.Code)
	return common.BytesToHash([]byte(data))
}

// serializeMessage serializes a QBFT message
func (a *ReplayAttack) serializeMessage(msg *types.QBFTMessage) ([]byte, error) {
	// In production, this would use proper encoding (RLP, protobuf, etc.)
	// This is a placeholder implementation

	// Create a structured byte array
	// Format: [Code(1)] [Sequence(8)] [Round(8)] [Address(20)] [Signature(65)] [Data...]

	size := 1 + 8 + 8 + 20 + 65 // minimum size
	if len(msg.CommittedSeal) > 0 {
		size += len(msg.CommittedSeal)
	}

	data := make([]byte, 0, size)

	// Add code
	data = append(data, byte(msg.Code))

	// Add sequence (big endian)
	sequenceBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		sequenceBytes[7-i] = byte(msg.Sequence >> (8 * i))
	}
	data = append(data, sequenceBytes...)

	// Add round (big endian)
	roundBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		roundBytes[7-i] = byte(msg.Round >> (8 * i))
	}
	data = append(data, roundBytes...)

	// Add address
	data = append(data, msg.Address.Bytes()...)

	// Add signature
	if len(msg.Signature) > 0 {
		data = append(data, msg.Signature...)
	} else {
		data = append(data, make([]byte, 65)...) // empty signature
	}

	// Add committed seal if present
	if len(msg.CommittedSeal) > 0 {
		data = append(data, msg.CommittedSeal...)
	}

	return data, nil
}

// sendMessage sends a message to targets
func (a *ReplayAttack) sendMessage(content []byte, targets []common.Address) error {
	// Implementation depends on actual network layer
	// This would interface with the P2P network to send messages

	// For now, we'll simulate sending
	// In production, this would:
	// 1. Get network interface from context or service
	// 2. Encode message for network transmission
	// 3. Send to each target through P2P layer

	if len(content) == 0 {
		return fmt.Errorf("empty message content")
	}

	if len(targets) == 0 {
		return fmt.Errorf("no targets specified")
	}

	// Simulate network delay
	time.Sleep(10 * time.Millisecond)

	return nil
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
