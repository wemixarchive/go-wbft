package manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// AttackManager implements the AttackManager interface
type AttackManager struct {
	mu              sync.RWMutex
	attacks         map[uint64]types.Attack
	attacksByStatus map[types.AttackStatus]map[uint64]types.Attack
	uidCounter      uint64
	registry        *registry.AttackRegistry
	historyStorage  types.HistoryStorage
	chainHandler    *ChainHandler
}

var _ types.AttackManager = (*AttackManager)(nil)

// NewAttackManager creates a new attack manager
func NewAttackManager(registry *registry.AttackRegistry, historyStorage types.HistoryStorage) *AttackManager {
	manager := &AttackManager{
		attacks:         make(map[uint64]types.Attack),
		attacksByStatus: make(map[types.AttackStatus]map[uint64]types.Attack),
		registry:        registry,
		historyStorage:  historyStorage,
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
		manager.attacksByStatus[status] = make(map[uint64]types.Attack)
	}

	// Create chain handler
	manager.chainHandler = NewChainHandler(manager)

	return manager
}

// RegisterAttack registers a new attack
func (m *AttackManager) RegisterAttack(attack types.Attack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid := attack.GetUID()
	if _, exists := m.attacks[uid]; exists {
		return fmt.Errorf("attack with UID %d already exists", uid)
	}

	m.attacks[uid] = attack
	m.updateStatusMap(attack, types.AttackStatusPending)

	// Save to history
	if m.historyStorage != nil {
		config := attack.GetConfig()
		if err := m.historyStorage.SaveAttackConfig(config); err != nil {
			return fmt.Errorf("failed to save attack config: %w", err)
		}
	}

	return nil
}

// UnregisterAttack removes an attack by UID
func (m *AttackManager) UnregisterAttack(uid uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	attack, exists := m.attacks[uid]
	if !exists {
		return fmt.Errorf("attack with UID %d not found", uid)
	}

	// Remove from status map
	status := attack.GetConfig().Status
	delete(m.attacksByStatus[status], uid)

	// Remove from main map
	delete(m.attacks, uid)

	// Update status to cancelled
	attack.SetStatus(types.AttackStatusCancelled)

	return nil
}

// GetAttack retrieves an attack by UID
func (m *AttackManager) GetAttack(uid uint64) (types.Attack, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attack, exists := m.attacks[uid]
	if !exists {
		return nil, fmt.Errorf("attack with UID %d not found", uid)
	}

	return attack, nil
}

// ListAttacks returns all registered attacks
func (m *AttackManager) ListAttacks() []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attacks := make([]types.Attack, 0, len(m.attacks))
	for _, attack := range m.attacks {
		attacks = append(attacks, attack)
	}

	return attacks
}

// ProcessEvent processes an event through all attacks
func (m *AttackManager) ProcessEvent(ctx context.Context, event types.Event) error {
	return m.chainHandler.ProcessEvent(ctx, event)
}

// GetActiveAttacks returns attacks in active status
func (m *AttackManager) GetActiveAttacks() []types.Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activeAttacks := make([]types.Attack, 0)

	// Get attacks in pending and active status
	for _, status := range []types.AttackStatus{types.AttackStatusPending, types.AttackStatusActive} {
		for _, attack := range m.attacksByStatus[status] {
			activeAttacks = append(activeAttacks, attack)
		}
	}

	return activeAttacks
}

// GenerateUID generates a new unique ID
func (m *AttackManager) GenerateUID() uint64 {
	return atomic.AddUint64(&m.uidCounter, 1)
}

// updateStatusMap updates the status map when attack status changes
func (m *AttackManager) updateStatusMap(attack types.Attack, newStatus types.AttackStatus) {
	config := attack.GetConfig()
	oldStatus := config.Status

	// Remove from old status map
	if oldMap, exists := m.attacksByStatus[oldStatus]; exists {
		delete(oldMap, config.UID)
	}

	// Add to new status map
	m.attacksByStatus[newStatus][config.UID] = attack

	// Update attack status
	attack.SetStatus(newStatus)
}

// EvaluateAndExecuteAttacks evaluates all active attacks and executes them if conditions are met
// Returns true if any attack indicates the message should be blocked
func (m *AttackManager) EvaluateAndExecuteAttacks(ctx context.Context, event types.Event) (types.AttackDecision, error) {
	activeAttacks := m.GetActiveAttacks()

	// Process attacks sequentially to get immediate decision
	for _, attack := range activeAttacks {
		// Check if attack conditions are met
		if !attack.CheckExecuteCondition(ctx, event) {
			continue
		}

		// Attack
		if attack.GetType() == types.AttackTamper || attack.GetType() == types.AttackOmit {
			log.Debug("this attack is Tamper or Omit")
			return types.AttackDecision{
				ShouldBlock: false,
				Reason:      "No attack blocked the message",
			}, nil
		}

		// Execute the attack
		result, err := attack.Execute(ctx, event)
		if err != nil {
			log.Error("Failed to execute attack",
				"attack", attack.GetConfig().Name,
				"error", err)
			// Update attack status
			m.mu.Lock()
			m.updateStatusMap(attack, types.AttackStatusFailed)
			m.mu.Unlock()
			continue
		}

		// Update attack status
		m.mu.Lock()
		m.updateStatusMap(attack, types.AttackStatusExecuted)
		m.mu.Unlock()

		// Save result to history
		if m.historyStorage != nil && result != nil {
			_ = m.historyStorage.SaveAttackResult(*result)
		}

		// Check if this attack wants to block the message
		if result != nil && result.BlockMessage {
			return types.AttackDecision{
				ShouldBlock: true,
				AttackUID:   attack.GetUID(),
				AttackType:  attack.GetType(),
				Reason:      result.BlockReason,
				Result:      result,
			}, nil
		}
	}

	// No attack blocked the message
	return types.AttackDecision{
		ShouldBlock: false,
		Reason:      "No attack blocked the message",
	}, nil
}

// ProcessEventAsync processes event asynchronously for non-blocking attacks
func (m *AttackManager) ProcessEventAsync(ctx context.Context, event types.Event) error {
	// For attacks that don't need immediate decision (logging, analysis, etc.)
	return m.chainHandler.ProcessEvent(ctx, event)
}
