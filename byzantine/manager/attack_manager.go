package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// AttackManager implements the AttackManager interface
type AttackManager struct {
	mu sync.RWMutex

	// Primary UID-based map for fast lookup
	attacksByUID map[string]types.Attack

	// Status tracking - stores UIDs instaed of Attack objects
	uidsByStatus map[types.AttackStatus]map[string]bool

	// UID generator
	uidGenerator types.UIDGenerator

	// Dependencies
	registry     *registry.AttackRegistry
	chainHandler *ChainHandler

	// Metrics
	totalAttacks  uint64
	activeAttacks uint64
}

var _ types.AttackManager = (*AttackManager)(nil)

// NewAttackManager creates a new attack manager
func NewAttackManager(registry *registry.AttackRegistry) *AttackManager {
	manager := &AttackManager{
		attacksByUID: make(map[string]types.Attack),
		uidsByStatus: make(map[types.AttackStatus]map[string]bool),
		uidGenerator: types.NewUIDGenerator(),
		registry:     registry,
	}

	// Initialize status maps
	for _, status := range []types.AttackStatus{
		types.AttackStatusPending,
		types.AttackStatusActive,
		types.AttackStatusExecuted,
		types.AttackStatusCompleted,
		types.AttackStatusFailed,
		types.AttackStatusCancelled,
	} {
		manager.uidsByStatus[status] = make(map[string]bool)
	}

	// Create chain handler
	manager.chainHandler = NewChainHandler(manager)

	return manager
}

// RegisterAttack registers a new attack
func (m *AttackManager) RegisterAttack(attack types.Attack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config := attack.GetConfig()

	// Generate standardized UID
	uid := m.uidGenerator.GenerateWithRange(
		config.Type,
		config.SequenceStart,
		config.SequenceEnd,
		config.Round,
	)

	// Check for duplicates
	if _, exists := m.attacksByUID[uid]; exists {
		return fmt.Errorf("attack with UID %s already exists", uid)
	}

	// Update config with generated UID
	config.UID = uid
	config.Status = types.AttackStatusPending
	attack.SetConfig(config)

	// Store in primary map
	m.attacksByUID[uid] = attack

	// Update status tracking
	m.updateStatusTracking(uid, types.AttackStatusPending, "")

	// Update metrics
	m.totalAttacks++
	m.activeAttacks++

	return nil
}

// GetAttacksBySequenceRange retrieves attacks that match the given sequence
func (m *AttackManager) GetAttacksBySequenceRange(attackType types.AttackType, sequence, round uint64) []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matchingAttacks []types.Attack

	for _, attack := range m.attacksByUID {
		config := attack.GetConfig()

		if config.Type != attackType || config.Round != round {
			continue
		}

		if config.IsInSequenceRange(sequence) && config.CanExecute() {
			matchingAttacks = append(matchingAttacks, attack)
		}
	}

	return matchingAttacks
}

// FindAttackForExecution finds an attack that can be executed at given sequence/round
func (m *AttackManager) FindAttackForExecution(attackType types.AttackType, sequence, round uint64) (types.Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// First try exact match for single sequence attacks
	exactUID := m.uidGenerator.Generate(attackType, sequence,
		round)
	if attack, exists := m.attacksByUID[exactUID]; exists {
		config := attack.GetConfig()
		if config.Type == attackType && config.Round == round &&
			config.CanExecute() {
			return attack, true
		}
	}

	// Then check range-based attacks
	for _, attack := range m.attacksByUID {
		config := attack.GetConfig()

		// Skip if wrong type or round
		if config.Type != attackType || config.Round != round {
			continue
		}

		// Check if sequence is in range and attack can execute
		if config.IsInSequenceRange(sequence) &&
			config.CanExecute() {
			return attack, true
		}
	}

	return nil, false
}

// MarkAttackExecuted updates attack execution state
func (m *AttackManager) MarkAttackExecuted(uid string, sequence uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	attack, exists := m.attacksByUID[uid]
	if !exists {
		return fmt.Errorf("attack not found: %s", uid)
	}

	config := attack.GetConfig()

	// TODO:
	// check if sequence is already marked as executed
	if config.LastExecutedSeq == sequence {
		log.Debug("[BYZ] Mark executed attack to completed",
			"uid", uid,
			"sequence", sequence,
			"sequence_range", fmt.Sprintf("%d-%d", config.SequenceStart, config.SequenceEnd),
			"execution_count", config.ExecutionCount,
			"max_execution_count", config.MaxExecutionCount,
			"status", config.Status,
			"last_executed_seq", config.LastExecutedSeq)
		return nil
	}

	// Update execution count and sequence
	config.ExecutionCount++
	config.LastExecutedSeq = sequence
	now := time.Now()
	config.ExecutedAt = &now

	// Check if max executions reached
	// For range attacks with MaxExecutionCount, check against that limit
	if config.MaxExecutionCount > 0 && config.SequenceStart != config.SequenceEnd {
		if config.ExecutionCount >= config.MaxExecutionCount {
			config.Status = types.AttackStatusCompleted
			attack.SetStatus(types.AttackStatusCompleted)
		} else {
			// For range attacks, keep status as executed
			config.Status = types.AttackStatusExecuted
			attack.SetStatus(types.AttackStatusExecuted)
		}
	} else {
		// For single sequence attacks or attacks without MaxExecutionCount, use default limit of 1
		if config.ExecutionCount >= uint64(1) {
			config.Status = types.AttackStatusCompleted
			attack.SetStatus(types.AttackStatusCompleted)
		}
	}

	// Update the attack's config
	attack.SetConfig(config)
	m.attacksByUID[uid] = attack

	// Update status tracking
	if config.Status == types.AttackStatusCompleted {
		m.updateStatusTracking(uid, types.AttackStatusCompleted, types.AttackStatusExecuted)
	} else {
		m.updateStatusTracking(uid, types.AttackStatusExecuted, config.Status)
	}

	log.Debug("[BYZ] Mark executed attack to completed",
		"uid", uid,
		"sequence", sequence,
		"sequence_range", fmt.Sprintf("%d-%d", config.SequenceStart, config.SequenceEnd),
		"execution_count", config.ExecutionCount,
		"max_execution_count", config.MaxExecutionCount,
		"status", config.Status,
		"last_executed_seq", config.LastExecutedSeq)

	return nil
}

// GetAttackByUID retrieves an attack by its string UID
func (m *AttackManager) GetAttackByUID(uid string) (types.Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attack, exists := m.attacksByUID[uid]
	return attack, exists
}

// GetAttacksByCondition retrieves attacks matching the given condition
func (m *AttackManager) GetAttacksByCondition(attackType types.AttackType, sequence, round uint64) []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Generate lookup UID
	lookupUID := m.uidGenerator.Generate(attackType, sequence, round)
	if attack, exists := m.GetAttackByUID(lookupUID); exists {
		// Check if the attack is eligible
		config := attack.GetConfig()
		if m.isAttackEligible(config) {
			return []types.Attack{attack}
		}
	}

	// If no direct match, use GetAttacksBySequenceRange
	return m.GetAttacksBySequenceRange(attackType, sequence, round)
}

// generateLookupPatterns generates possible UID patterns for wildcard matching
func (m *AttackManager) generateLookupPatterns(attackType types.AttackType, sequence, round uint64) []string {
	patterns := []string{
		// Exact match
		m.uidGenerator.Generate(attackType, sequence, round),
		// Wildcard round (attacks that apply to all rounds)
		fmt.Sprintf("%s-%d-*", types.AttackTypeToString(attackType), sequence),
		// Wildcard sequence and round (attacks that apply globally)
		fmt.Sprintf("%s-*-*", types.AttackTypeToString(attackType)),
	}
	return patterns
}

// UnregisterAttack removes an attack by UID
func (m *AttackManager) UnregisterAttack(uid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	attack, exists := m.attacksByUID[uid]
	if !exists {
		return fmt.Errorf("attack with UID %s not found", uid)
	}

	config := attack.GetConfig()

	// Update status to cancelled first
	m.updateStatusTracking(uid, types.AttackStatusCancelled, config.Status)

	// Update metrics
	if config.Status == types.AttackStatusPending || config.Status == types.AttackStatusActive {
		m.activeAttacks--
	}

	delete(m.attacksByUID, uid)

	// Update metrics
	m.totalAttacks--
	m.activeAttacks--

	log.Info("Attack unregistered", "uid", uid)

	return nil
}

// GetAttack retrieves an attack by UID
func (m *AttackManager) GetAttack(uid string) (types.Attack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attack, exists := m.GetAttackByUID(uid)
	if !exists {
		return nil, fmt.Errorf("attack with UID %s not found", uid)
	}

	return attack, nil
}

// ListAttacks returns all registered attacks
func (m *AttackManager) ListAttacks() []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attacks := make([]types.Attack, 0, len(m.attacksByUID))
	for _, attack := range m.attacksByUID {
		attacks = append(attacks, attack)
	}

	return attacks
}

// GetActiveAttacks returns attacks in active status
func (m *AttackManager) GetActiveAttacks() []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var activeAttacks []types.Attack

	// Collect UIDs from pending and active status
	activeUIDs := make(map[string]bool)
	for uid := range m.uidsByStatus[types.AttackStatusPending] {
		activeUIDs[uid] = true
	}
	for uid := range m.uidsByStatus[types.AttackStatusActive] {
		activeUIDs[uid] = true
	}

	// Retrieve attacks by UID
	for uid := range activeUIDs {
		if attack, exists := m.attacksByUID[uid]; exists {
			activeAttacks = append(activeAttacks, attack)
		}
	}

	return activeAttacks
}

// UpdateStatusMap updates the status map when attack status changes
func (m *AttackManager) UpdateStatusMap(attack types.Attack, newStatus types.AttackStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	config := attack.GetConfig()
	uid := config.UID
	oldStatus := config.Status

	// Update status tracking
	m.updateStatusTracking(uid, newStatus, oldStatus)

	// Update attack status
	attack.SetStatus(newStatus)
	m.attacksByUID[uid] = attack

	// Update metrics
	if (oldStatus == types.AttackStatusPending || oldStatus == types.AttackStatusActive) &&
		(newStatus != types.AttackStatusPending && newStatus != types.AttackStatusActive) {
		m.activeAttacks--
	} else if (oldStatus != types.AttackStatusPending && oldStatus != types.AttackStatusActive) &&
		(newStatus == types.AttackStatusPending || newStatus == types.AttackStatusActive) {
		m.activeAttacks++
	}
}

// ProcessEvent processes an event through the chain handler
func (m *AttackManager) ProcessEvent(ctx context.Context, event types.Event) error {
	return m.chainHandler.ProcessEvent(ctx, event)
}

// ProcessEventAsync processes event asynchronously
func (m *AttackManager) ProcessEventAsync(ctx context.Context, event types.Event) error {
	go func() {
		if err := m.chainHandler.ProcessEvent(ctx, event); err != nil {
			log.Error("Failed to process event asynchronously", "error", err)
		}
	}()
	return nil
}

// GetUIDGenerator returns the UID generator
func (m *AttackManager) GetUIDGenerator() types.UIDGenerator {
	return m.uidGenerator
}

// FindExecutableAttack finds an executable attack based on type, sequence, round, and message code
func (m *AttackManager) FindExecutableAttack(attackType types.AttackType,
	sequence, round uint64, msgCode types.MessageCode) (types.Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// TODO:
	// 1. m.attacksByUID -> GetActiveAttacks() : m.attacksByUID는 지난 attack도 포함되므로,
	for _, attack := range m.attacksByUID {
		config := attack.GetConfig()

		// check enabled, status and execution limits
		if !config.CanExecute() {
			continue
		}

		// Check attack type
		if config.Type != attackType {
			continue
		}

		// Check sequence range
		if !config.IsInSequenceRange(sequence) {
			continue
		}

		// Check round
		if config.Round != round {
			continue
		}

		// Check Code
		// Extract code from ParsedParameters
		var attackCode types.MessageCode
		switch params := config.ParsedParameters.(type) {
		case *types.SilentAttackParams:
			attackCode = params.Code
		case *types.TamperAttackParams:
			attackCode = params.Code
		case *types.FakeAttackParams:
			attackCode = params.Code
		case *types.OmitAttackParams:
			attackCode = params.Code
		case *types.RoleSpoofAttackParams:
			attackCode = params.Code
		case *types.ReplayAttackParams:
			attackCode = params.Code
		case *types.StoreAttackParams:
			attackCode = params.Code
		case *types.DosAttackParams:
			attackCode = params.Code
		default:
			// Fallback to Parameters map
			if codeVal, ok := config.Parameters["code"]; ok {
				switch v := codeVal.(type) {
				case float64:
					attackCode = types.MessageCode(v)
				case int:
					attackCode = types.MessageCode(v)
				}
			}
		}

		// Match message code
		if attackCode != msgCode {
			continue
		}

		log.Trace("[BYZ] Found executable attack",
			"uid", config.UID,
			"type", config.Type,
			"sequence", sequence,
			"sequence_range", fmt.Sprintf("%d-%d", config.SequenceStart, config.SequenceEnd),
			"execution_count", config.ExecutionCount,
			"code", config.Parameters["code"])

		return attack, true
	}

	return nil, false
}

// isAttackEligible checks if an attack is eligible based on its status
func (m *AttackManager) isAttackEligible(config types.AttackConfig) bool {
	// Check if enabled
	if !config.Enabled {
		return false
	}

	// Check status
	switch config.Status {
	case types.AttackStatusCompleted, types.AttackStatusCancelled, types.AttackStatusFailed:
		return false
	default:
		return true
	}
}

// canExecuteAttack checks if an attack can be executed based on execution limits
func (m *AttackManager) canExecuteAttack(attack types.Attack, sequence uint64) bool {
	config := attack.GetConfig()

	if !config.IsInSequenceRange(sequence) {
		log.Trace("[BYZ] Attack sequence out of range",
			"uid", config.UID,
			"sequence", sequence)
		return false
	}

	if config.ExecutionCount >= config.MaxExecutionCount && config.MaxExecutionCount > 0 {
		log.Trace("[BYZ] Attack execution limit reached",
			"uid", config.UID,
			"executed", config.ExecutionCount,
			"max_executions", config.MaxExecutionCount)
		return false
	}

	// For single sequence attacks, check if already executed at this sequence
	if config.LastExecutedSeq == sequence {
		log.Trace("[BYZ] Attack already executed at this sequence",
			"uid", config.UID,
			"sequence", sequence)
		return false
	}

	return true
}

// Helper methods

// updateStatusTracking updates the UID tracking by status
func (m *AttackManager) updateStatusTracking(uid string, newStatus, oldStatus types.AttackStatus) {
	// Remove from old status if exists
	if oldStatus != "" {
		delete(m.uidsByStatus[oldStatus], uid)
	}

	// Add to new status
	m.uidsByStatus[newStatus][uid] = true
}

// getApplicableAttackTypes returns attack types applicable to the message and direction
func (m *AttackManager) getApplicableAttackTypes(msgCode types.MessageCode, direction string) []types.AttackType {
	attackTypes := []types.AttackType{
		types.AttackTypeSilentMessage, // Can apply to any message
	}

	// Add message-specific attack types
	switch msgCode {
	case types.MessageCodePrePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeFakeMessage, types.AttackTypeStoreMessage)
	case types.MessageCodePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage, types.AttackTypeStoreMessage)
	case types.MessageCodeCommit:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage, types.AttackTypeStoreMessage)
	case types.MessageCodeRoundChange:
		attackTypes = append(attackTypes, types.AttackTypeFakeMessage, types.AttackTypeStoreMessage)
	}

	// Add direction-specific types
	if direction == types.DirectionSend {
		attackTypes = append(attackTypes, types.AttackTypeRoleSpoofed)
	} else {
		attackTypes = append(attackTypes, types.AttackTypeReplay)
	}

	return attackTypes
}
