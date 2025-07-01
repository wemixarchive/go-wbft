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

	// Pattern index for wildcard matching
	patternIndex map[string][]string

	// UID generator
	uidGenerator types.UIDGenerator

	// Dependencies
	registry       *registry.AttackRegistry
	historyStorage types.HistoryStorage
	chainHandler   *ChainHandler

	// Metrics
	totalAttacks  uint64
	activeAttacks uint64

	//attacks         map[uint64]types.Attack
	//attacks map[string]types.Attack
	//attacksByStatus map[types.AttackStatus]map[uint64]types.Attack
	//attacksByStatus map[types.AttackStatus]map[string]types.Attack
	//attackIndex    map[string][]string // key: "type-code-seq-round", value: []UIDs
}

var _ types.AttackManager = (*AttackManager)(nil)

// NewAttackManager creates a new attack manager
func NewAttackManager(registry *registry.AttackRegistry, historyStorage types.HistoryStorage) *AttackManager {
	manager := &AttackManager{
		attacksByUID:   make(map[string]types.Attack),
		uidsByStatus:   make(map[types.AttackStatus]map[string]bool),
		patternIndex:   make(map[string][]string),
		uidGenerator:   types.NewUIDGenerator(),
		registry:       registry,
		historyStorage: historyStorage,
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
	//log.Info("[byzantine] attack manager ", "attack", attack)
	//log.Info("[byzantine] attack manager ", "config", config)

	// Generate standardized UID
	uid := m.uidGenerator.Generate(
		config.Type,
		config.Sequence,
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

	// Update pattern index for wildcard matching
	m.updatePatternIndex(uid, config)

	// Update metrics
	m.totalAttacks++
	m.activeAttacks++

	// Save to history if available
	if m.historyStorage != nil {
		if err := m.historyStorage.SaveAttackConfig(config); err != nil {
			log.Error("Failed to save attack config to history", "uid", uid, "error", err)
		}
	}

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
		if m.isAttackEligible(attack) {
			return []types.Attack{attack}
		}
	}

	// If no lookup found, check pattern index for wildcard
	var matches []types.Attack
	patterns := m.generateLookupPatterns(attackType, sequence, round)

	for _, pattern := range patterns {
		if uids, exists := m.patternIndex[pattern]; exists {
			for _, uid := range uids {
				if attack, ok := m.attacksByUID[uid]; ok && m.isAttackEligible(attack) {
					matches = append(matches, attack)
				}
			}
		}
	}

	return matches
}

// generateLookupPatterns generates possible UID patterns for wildcard matching
func (m *AttackManager) generateLookupPatterns(attackType types.AttackType, sequence, round uint64) []string {
	patterns := []string{
		// Exact match
		m.uidGenerator.Generate(attackType, sequence, round),
		// Wildcard round (attacks that apply to all rounds)
		fmt.Sprintf("%s-%d-*", types.AttachTypeToString(attackType), sequence),
		// Wildcard sequence and round (attacks that apply globally)
		fmt.Sprintf("%s-*-*", types.AttachTypeToString(attackType)),
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

	// Remove from pattern index
	m.removeFromPatternIndex(uid, config)

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

// EvaluateAndExecuteAttacks evaluates all active attacks and executes them if conditions are met
// Returns true if any attack indicates the message should be blocked
func (m *AttackManager) EvaluateAndExecuteAttacks(ctx context.Context, event types.Event) (types.AttackDecision, error) {
	// Extract message details from event
	msgEvent, ok := event.Data.(*types.MessageEvent)
	if !ok {
		return types.AttackDecision{
			ShouldAttack: false,
			Reason:       "Not a message event",
		}, nil
	}

	// Get direction from metadata
	direction := event.Metadata["direction"].(string)

	// Try each applicable attack type
	applicableTypes := m.getApplicableAttackTypes(msgEvent.MessageCode, direction)

	for _, attackType := range applicableTypes {
		// Generate UID for direct lookup
		uid := m.uidGenerator.Generate(attackType, event.Sequence, event.Round)

		if attack, exists := m.GetAttackByUID(uid); exists {
			if !m.isAttackEligible(attack) {
				continue
			}

			// Check execution condition
			if !attack.CheckExecuteCondition(ctx, event) {
				continue
			}

			// Execute attack
			result, err := m.executeAttack(ctx, attack, event)
			if err != nil {
				log.Error("Failed to execute attack", "uid", uid, "error", err)
				continue
			}

			// Check if message should be blocked
			if result != nil && result.BlockMessage {
				return types.AttackDecision{
					ShouldAttack: true,
					AttackUID:    uid,
					AttackType:   attackType,
					Reason:       result.BlockReason,
					Result:       result,
				}, nil
			}
		}
	}

	// No attack blocked the message
	return types.AttackDecision{
		ShouldAttack: false,
		Reason:       "No attack conditions met",
	}, nil
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

// updatePatternIndex updates the pattern index for wildcard matching
func (m *AttackManager) updatePatternIndex(uid string, config types.AttackConfig) {
	// Index exact pattern
	exactPattern := m.uidGenerator.Generate(config.Type, config.Sequence, config.Round)
	m.patternIndex[exactPattern] = append(m.patternIndex[exactPattern], uid)

	// Index wildcard patterns for flexible matching
	// Pattern for any round: "type-code-sequence-*"
	anyRoundPattern := fmt.Sprintf("%s-%d-%d-*",
		types.AttachTypeToString(config.Type), config.Parameters["code"], config.Sequence)
	m.patternIndex[anyRoundPattern] = append(m.patternIndex[anyRoundPattern], uid)

	// Pattern for any sequence and round: "type-code-*-*"
	globalPattern := fmt.Sprintf("%s-%d-*-*",
		types.AttachTypeToString(config.Type), config.Parameters["code"])
	m.patternIndex[globalPattern] = append(m.patternIndex[globalPattern], uid)
}

// removeFromPatternIndex removes UID from pattern index
func (m *AttackManager) removeFromPatternIndex(uid string, config types.AttackConfig) {
	patterns := []string{
		m.uidGenerator.Generate(config.Type, config.Sequence, config.Round),
		fmt.Sprintf("%s-%d-*", types.AttachTypeToString(config.Type), config.Sequence),
		fmt.Sprintf("%s-*-*", types.AttachTypeToString(config.Type)),
	}

	for _, pattern := range patterns {
		if uids, exists := m.patternIndex[pattern]; exists {
			// Remove uid from slice
			for i, u := range uids {
				if u == uid {
					m.patternIndex[pattern] = append(uids[:i], uids[i+1:]...)
					break
				}
			}
			// Clean up empty entries
			if len(m.patternIndex[pattern]) == 0 {
				delete(m.patternIndex, pattern)
			}
		}
	}
}

// generatePatterns generates patterns for wildcard matching
func (m *AttackManager) generatePatterns(attackType types.AttackType, code types.MessageCode, sequence, round uint64) []string {
	return []string{
		// Exact match already tried in GetAttacksByCondition
		// Try wildcard round
		fmt.Sprintf("%s-%d-%d-*", types.AttachTypeToString(attackType), code, sequence),
		// Try global wildcard
		fmt.Sprintf("%s-%d-*-*", types.AttachTypeToString(attackType), code),
	}
}

// isAttackEligible checks if an attack is eligible for execution
func (m *AttackManager) isAttackEligible(attack types.Attack) bool {
	config := attack.GetConfig()
	return config.Enabled &&
		config.Status != types.AttackStatusCompleted &&
		config.Status != types.AttackStatusCancelled &&
		config.Status != types.AttackStatusFailed
}

// getApplicableAttackTypes returns attack types applicable to the message and direction
func (m *AttackManager) getApplicableAttackTypes(msgCode types.MessageCode, direction string) []types.AttackType {
	attackTypes := []types.AttackType{
		types.AttackTypeSilentMessage, // Can apply to any message
	}

	// Add message-specific attack types
	switch msgCode {
	case types.MessageCodePrePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeFakeMessage)
	case types.MessageCodePrepare:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage)
	case types.MessageCodeCommit:
		attackTypes = append(attackTypes, types.AttackTypeTamperedMessage, types.AttackTypeOmitMessage)
	case types.MessageCodeRoundChange:
		attackTypes = append(attackTypes, types.AttackTypeFakeMessage)
	}

	// Add direction-specific types
	if direction == types.DirectionSend {
		attackTypes = append(attackTypes, types.AttackTypeRoleSpoofed)
	} else {
		attackTypes = append(attackTypes, types.AttackTypeReplay)
	}

	return attackTypes
}

// executeAttack executes an attack and handles status updates
func (m *AttackManager) executeAttack(ctx context.Context, attack types.Attack, event types.Event) (*types.AttackResult, error) {
	startTime := time.Now()

	// Execute the attack
	result, err := attack.Execute(ctx, event)

	// Update status based on result
	if err != nil {
		m.UpdateStatusMap(attack, types.AttackStatusFailed)
		return nil, err
	}

	m.UpdateStatusMap(attack, types.AttackStatusExecuted)

	// Save to history if available
	if m.historyStorage != nil && result != nil {
		if err := m.historyStorage.SaveAttackResult(*result); err != nil {
			log.Error("Failed to save attack result", "uid", attack.GetUID(), "error", err)
		}
	}

	// Add execution time to result
	if result != nil {
		result.Duration = time.Since(startTime)
	}

	return result, nil
}
