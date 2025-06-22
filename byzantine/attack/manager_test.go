package attack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock EventCollector
type MockEventCollector struct {
	mock.Mock
}

func (m *MockEventCollector) StartCollection(dataReqs []DataRequirement) error {
	args := m.Called(dataReqs)
	return args.Error(0)
}

func (m *MockEventCollector) StopCollection(attackID string) error {
	args := m.Called(attackID)
	return args.Error(0)
}

// Test: Manager should register attacks
func TestAttackManager_RegisterAttack(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	attack := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-123",
			sequence: 100,
			round:    0,
		},
	}

	// Mock expectations
	mockCollector.On("StartCollection", attack.RequiresData()).Return(nil)

	// When
	err := manager.Register(attack)

	// Then
	assert.NoError(t, err)

	// Verify attack is registered
	retrievedAttack, exists := manager.GetAttack("attack-123")
	assert.True(t, exists)
	assert.Equal(t, attack, retrievedAttack)

	mockCollector.AssertExpectations(t)
}

// Test: Manager should not register duplicate attacks
func TestAttackManager_RegisterDuplicateAttack(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	attack := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-123",
			sequence: 100,
			round:    0,
		},
	}

	// First registration
	mockCollector.On("StartCollection", attack.RequiresData()).Return(nil).Once()
	err1 := manager.Register(attack)
	assert.NoError(t, err1)

	// When - Try to register same attack again
	err2 := manager.Register(attack)

	// Then
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already registered")
}

// Test: Manager should unregister attacks
func TestAttackManager_UnregisterAttack(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	attack := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-123",
			sequence: 100,
			round:    0,
		},
	}

	// Register attack first
	mockCollector.On("StartCollection", attack.RequiresData()).Return(nil)
	mockCollector.On("StopCollection", "attack-123").Return(nil)

	manager.Register(attack)

	// When
	err := manager.Unregister("attack-123")

	// Then
	assert.NoError(t, err)

	// Verify attack is removed
	_, exists := manager.GetAttack("attack-123")
	assert.False(t, exists)

	mockCollector.AssertExpectations(t)
}

// Test: Manager should find attacks to execute
func TestAttackManager_GetAttacksToExecute(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	// Register multiple attacks
	attack1 := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-1",
			sequence: 100,
			round:    0,
		},
	}

	attack2 := &doubleCommitAttack{
		baseAttack: baseAttack{
			id:       "attack-2",
			sequence: 100,
			round:    0,
		},
	}

	attack3 := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-3",
			sequence: 200,
			round:    0,
		},
	}

	mockCollector.On("StartCollection", mock.Anything).Return(nil)

	manager.Register(attack1)
	manager.Register(attack2)
	manager.Register(attack3)

	// When - Check for sequence 100, round 0, prepare message
	attacks := manager.GetAttacksToExecute(100, 0, uint64(MessageCodePrepare))

	// Then
	assert.Len(t, attacks, 1)
	assert.Equal(t, "attack-1", attacks[0].ID())

	// When - Check for sequence 100, round 0, commit message
	attacks = manager.GetAttacksToExecute(100, 0, uint64(MessageCodeCommit))

	// Then
	assert.Len(t, attacks, 1)
	assert.Equal(t, "attack-2", attacks[0].ID())

	// When - Check for sequence 200
	attacks = manager.GetAttacksToExecute(200, 0, uint64(MessageCodePrepare))

	// Then
	assert.Len(t, attacks, 1)
	assert.Equal(t, "attack-3", attacks[0].ID())
}

// Test: Manager should track attack status
func TestAttackManager_AttackStatus(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	attack := &doublePrepareAttack{
		baseAttack: baseAttack{
			id:       "attack-123",
			sequence: 100,
			round:    0,
		},
	}

	mockCollector.On("StartCollection", attack.RequiresData()).Return(nil)

	// When - Register attack
	manager.Register(attack)

	// Then - Check initial status
	status := manager.GetAttackStatus("attack-123")
	assert.Equal(t, AttackStatusActive, status)

	// When - Mark as executed
	manager.MarkExecuted("attack-123")

	// Then - Check updated status
	status = manager.GetAttackStatus("attack-123")
	assert.Equal(t, AttackStatusExecuted, status)
}

// Test: Manager should handle concurrent access
func TestAttackManager_ConcurrentAccess(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	mockCollector.On("StartCollection", mock.Anything).Return(nil).Maybe()
	mockCollector.On("StopCollection", mock.Anything).Return(nil).Maybe()

	// When - Concurrent operations
	done := make(chan bool)

	// Goroutine 1: Register attacks
	go func() {
		for i := 0; i < 10; i++ {
			attack := &doublePrepareAttack{
				baseAttack: baseAttack{
					id:       fmt.Sprintf("attack-%d", i),
					sequence: uint64(100 + i),
					round:    0,
				},
			}
			manager.Register(attack)
		}
		done <- true
	}()

	// Goroutine 2: Query attacks
	go func() {
		for i := 0; i < 20; i++ {
			manager.GetAttacksToExecute(uint64(100+i%10), 0, uint64(MessageCodePrepare))
		}
		done <- true
	}()

	// Goroutine 3: Unregister attacks
	go func() {
		for i := 0; i < 5; i++ {
			manager.Unregister(fmt.Sprintf("attack-%d", i))
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Then - No panic, manager still functional
	attacks := manager.ListAllAttacks()
	assert.NotNil(t, attacks)
}

// Test: Manager should clean up executed attacks
func TestAttackManager_CleanupExecutedAttacks(t *testing.T) {
	// Given
	mockCollector := new(MockEventCollector)
	manager := NewAttackManager(mockCollector)

	// Register multiple attacks
	for i := 0; i < 5; i++ {
		attack := &doublePrepareAttack{
			baseAttack: baseAttack{
				id:       fmt.Sprintf("attack-%d", i),
				sequence: uint64(100 + i),
				round:    0,
			},
		}
		mockCollector.On("StartCollection", attack.RequiresData()).Return(nil)
		manager.Register(attack)
	}

	// Mark some as executed
	manager.MarkExecuted("attack-0")
	manager.MarkExecuted("attack-2")
	manager.MarkExecuted("attack-4")

	mockCollector.On("StopCollection", mock.Anything).Return(nil).Times(3)

	// When
	cleaned := manager.CleanupExecuted()

	// Then
	assert.Equal(t, 3, cleaned)

	// Verify only active attacks remain
	remaining := manager.ListAllAttacks()
	assert.Len(t, remaining, 2)

	for _, attack := range remaining {
		status := manager.GetAttackStatus(attack.ID())
		assert.Equal(t, AttackStatusActive, status)
	}
}
