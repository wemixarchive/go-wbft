package attack

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"sync"
)

// EventCollector interface for data collection
//type EventCollector interface {
//	StartCollection(dataReqs []types.DataRequirement) error
//	StopCollection(attackID string) error
//}

// AttackManager interface for managing attacks
type AttackManager interface {
	Register(attack Attack) error
	Unregister(attackID string) error
	GetAttack(attackID string) (Attack, bool)
	GetAttacksToExecute(sequence, round uint64, msgCode uint64) []Attack
	GetAttackStatus(attackID string) AttackStatus
	MarkExecuted(attackID string)
	ListAllAttacks() []Attack
	CleanupExecuted() int
}

// attackInfo holds attack metadata
type attackInfo struct {
	attack Attack
	status AttackStatus
}

// attackManager implements AttackManager interface
type attackManager struct {
	attacks   map[string]*attackInfo
	collector types.EventCollector
	mu        sync.RWMutex
}

// NewAttackManager creates a new attack manager
func NewAttackManager(collector types.EventCollector) AttackManager {
	return &attackManager{
		attacks:   make(map[string]*attackInfo),
		collector: collector,
	}
}

// Register registers a new attack
func (m *attackManager) Register(attack Attack) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if attack already exists
	if _, exists := m.attacks[attack.ID()]; exists {
		return errors.New("attack already registered")
	}

	// Start data collection if required
	if m.collector != nil {
		dataReqs := attack.RequiresData()
		if len(dataReqs) > 0 {
			if err := m.collector.Start(dataReqs); err != nil {
				return fmt.Errorf("failed to start data collection: %w", err)
			}
		}
	}

	// Register the attack
	m.attacks[attack.ID()] = &attackInfo{
		attack: attack,
		status: AttackStatusActive,
	}

	return nil
}

// Unregister removes an attack
func (m *attackManager) Unregister(attackID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if attack exists
	if _, exists := m.attacks[attackID]; !exists {
		return errors.New("attack not found")
	}

	// Stop data collection
	if m.collector != nil {
		if err := m.collector.Stop(attackID); err != nil {
			// Log error but continue with unregistration
			// In real implementation, proper logging would be added
		}
	}

	// Remove the attack
	delete(m.attacks, attackID)

	return nil
}

// GetAttack retrieves a specific attack
func (m *attackManager) GetAttack(attackID string) (Attack, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if info, exists := m.attacks[attackID]; exists {
		return info.attack, true
	}

	return nil, false
}

// GetAttacksToExecute finds attacks that should execute for given conditions
func (m *attackManager) GetAttacksToExecute(sequence, round uint64, msgCode uint64) []Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var attacksToExecute []Attack

	for _, info := range m.attacks {
		// Only check active attacks
		if info.status != AttackStatusActive {
			continue
		}

		// Check if attack should execute
		if info.attack.ShouldExecute(sequence, round, msgCode) {
			attacksToExecute = append(attacksToExecute, info.attack)
		}
	}

	return attacksToExecute
}

// GetAttackStatus returns the status of an attack
func (m *attackManager) GetAttackStatus(attackID string) AttackStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if info, exists := m.attacks[attackID]; exists {
		return info.status
	}

	return AttackStatusCancelled // Default for non-existent attacks
}

// MarkExecuted marks an attack as executed
func (m *attackManager) MarkExecuted(attackID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, exists := m.attacks[attackID]; exists {
		info.status = AttackStatusExecuted
	}
}

// ListAllAttacks returns all registered attacks
func (m *attackManager) ListAllAttacks() []Attack {
	m.mu.RLock()
	defer m.mu.RUnlock()

	attacks := make([]Attack, 0, len(m.attacks))
	for _, info := range m.attacks {
		attacks = append(attacks, info.attack)
	}

	return attacks
}

// CleanupExecuted removes all executed attacks
func (m *attackManager) CleanupExecuted() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := []string{}

	// Find executed attacks
	for id, info := range m.attacks {
		if info.status == AttackStatusExecuted {
			toRemove = append(toRemove, id)
		}
	}

	// Remove executed attacks
	for _, id := range toRemove {
		// Stop data collection
		if m.collector != nil {
			m.collector.Stop(id)
		}
		delete(m.attacks, id)
	}

	return len(toRemove)
}
