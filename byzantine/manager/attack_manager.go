package manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
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
