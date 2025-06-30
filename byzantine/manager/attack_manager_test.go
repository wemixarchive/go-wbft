package manager

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/byzantine/registry"
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAttack implements types.Attack for testing
type MockAttack struct {
	config      types.AttackConfig
	executed    bool
	shouldBlock bool
}

func NewMockAttack(config types.AttackConfig) *MockAttack {
	return &MockAttack{
		config: config,
	}
}

func (m *MockAttack) CheckExecuteCondition(ctx context.Context, event types.Event) bool {
	return m.config.Sequence == event.Sequence && m.config.Round == event.Round
}

func (m *MockAttack) Execute(ctx context.Context, event types.Event) (*types.AttackResult, error) {
	m.executed = true
	return &types.AttackResult{
		UID:          m.config.UID,
		Success:      true,
		ExecutedAt:   time.Now(),
		BlockMessage: m.shouldBlock,
		BlockReason:  "Test block",
	}, nil
}

func (m *MockAttack) SetStatus(status types.AttackStatus) {
	m.config.Status = status
}

func (m *MockAttack) SetConfig(config types.AttackConfig) {
	m.config = config
}

func (m *MockAttack) GetUID() string {
	return m.config.UID
}

func (m *MockAttack) GetType() types.AttackType {
	return m.config.Type
}

func (m *MockAttack) GetConfig() types.AttackConfig {
	return m.config
}

func (m *MockAttack) GetStatus() types.AttackStatus {
	return m.config.Status
}

// TestRefactoredAttackManagerLookup tests UID-based lookup
func TestRefactoredAttackManagerLookup(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register multiple attacks
	attacks := []types.AttackConfig{
		{
			Name:     "attack1",
			Type:     types.AttackTypeTamperedMessage,
			Code:     types.MessageCodePrePrepare,
			Sequence: 100,
			Round:    0,
			Enabled:  true,
		},
		{
			Name:     "attack2",
			Type:     types.AttackTypeSilentMessage,
			Code:     types.MessageCodePrepare,
			Sequence: 200,
			Round:    1,
			Enabled:  true,
		},
		{
			Name:     "attack3",
			Type:     types.AttackTypeFakeMessage,
			Code:     types.MessageCodeCommit,
			Sequence: 300,
			Round:    2,
			Enabled:  true,
		},
	}

	registeredUIDs := make([]string, 0)

	// Register attacks
	for _, config := range attacks {
		attack := NewMockAttack(config)
		err := manager.RegisterAttack(attack)
		require.NoError(t, err)

		// Generate expected UID
		expectedUID := manager.GetUIDGenerator().Generate(
			config.Type, config.Code, config.Sequence, config.Round)
		registeredUIDs = append(registeredUIDs, expectedUID)

		// Verify UID was set correctly
		assert.Equal(t, expectedUID, attack.GetConfig().UID)
	}

	// Test lookup for each attack
	for i, uid := range registeredUIDs {
		start := time.Now()
		found, exists := manager.GetAttackByUID(uid)
		duration := time.Since(start)

		assert.True(t, exists, "Attack should exist for UID: %s", uid)
		assert.NotNil(t, found)
		assert.Equal(t, attacks[i].Name, found.GetConfig().Name)

		// Verify performance (should be very fast)
		assert.Less(t, duration.Nanoseconds(), int64(1000), "Lookup should be O(1)")
	}

	// Test non-existent UID
	_, exists := manager.GetAttackByUID("non-existent-uid")
	assert.False(t, exists)
}

// TestGetAttacksByCondition tests condition-based lookup
func TestGetAttacksByCondition(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register attacks with same type/code but different sequence/round
	baseConfig := types.AttackConfig{
		Type:    types.AttackTypeTamperedMessage,
		Code:    types.MessageCodePrePrepare,
		Enabled: true,
	}

	// Attack for specific sequence and round
	attack1 := NewMockAttack(baseConfig)
	attack1.config.Sequence = 100
	attack1.config.Round = 0
	attack1.config.Name = "specific"
	manager.RegisterAttack(attack1)

	// Attack for any round (wildcard)
	attack2 := NewMockAttack(baseConfig)
	attack2.config.Sequence = 100
	attack2.config.Round = 999 // Will be matched by wildcard
	attack2.config.Name = "wildcard-round"
	manager.RegisterAttack(attack2)

	// Test exact match
	attacks := manager.GetAttacksByCondition(
		types.AttackTypeTamperedMessage,
		types.MessageCodePrePrepare,
		100,
		0,
	)
	assert.Len(t, attacks, 1)
	assert.Equal(t, "specific", attacks[0].GetConfig().Name)

	// Test wildcard matching (currently not implemented in basic version)
	// This would require additional implementation in the pattern matching logic
}

// TestAttackStatusTracking tests status-based tracking
func TestAttackStatusTracking(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register multiple attacks
	configs := []types.AttackConfig{
		{Name: "pending1", Type: types.AttackTypeTamperedMessage, Code: 1, Sequence: 1, Round: 0, Enabled: true},
		{Name: "pending2", Type: types.AttackTypeSilentMessage, Code: 2, Sequence: 2, Round: 0, Enabled: true},
		{Name: "active1", Type: types.AttackTypeFakeMessage, Code: 3, Sequence: 3, Round: 0, Enabled: true},
	}

	attacks := make([]*MockAttack, 0)
	for _, config := range configs {
		attack := NewMockAttack(config)
		err := manager.RegisterAttack(attack)
		require.NoError(t, err)
		attacks = append(attacks, attack)
	}

	// All should be pending initially
	activeAttacks := manager.GetActiveAttacks()
	assert.Len(t, activeAttacks, 3)

	// Update one to active status
	manager.UpdateStatusMap(attacks[2], types.AttackStatusActive)

	// Update one to completed
	manager.UpdateStatusMap(attacks[1], types.AttackStatusCompleted)

	// Check active attacks (pending + active)
	activeAttacks = manager.GetActiveAttacks()
	assert.Len(t, activeAttacks, 2)

	// Verify status tracking
	for _, attack := range activeAttacks {
		status := attack.GetStatus()
		assert.True(t, status == types.AttackStatusPending || status == types.AttackStatusActive)
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Concurrent attack registration
	done := make(chan bool)
	attackCount := 100

	// Register attacks concurrently
	for i := 0; i < attackCount; i++ {
		go func(idx int) {
			config := types.AttackConfig{
				Name:     fmt.Sprintf("attack%d", idx),
				Type:     types.AttackTypeTamperedMessage,
				Code:     types.MessageCode(idx % 5),
				Sequence: uint64(idx),
				Round:    0,
				Enabled:  true,
			}
			attack := NewMockAttack(config)
			err := manager.RegisterAttack(attack)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < attackCount; i++ {
		<-done
	}

	// Verify all attacks registered
	allAttacks := manager.ListAttacks()
	assert.Len(t, allAttacks, attackCount)

	// Concurrent lookups
	lookupDone := make(chan bool)
	for i := 0; i < attackCount; i++ {
		go func(idx int) {
			uid := manager.GetUIDGenerator().Generate(
				types.AttackTypeTamperedMessage,
				types.MessageCode(idx%5),
				uint64(idx),
				0,
			)
			_, exists := manager.GetAttackByUID(uid)
			assert.True(t, exists)
			lookupDone <- true
		}(i)
	}

	// Wait for all lookups
	for i := 0; i < attackCount; i++ {
		<-lookupDone
	}
}

// BenchmarkUIDLookup benchmarks UID-based lookup performance
func BenchmarkUIDLookup(b *testing.B) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register 10000 attacks
	for i := 0; i < 10000; i++ {
		config := types.AttackConfig{
			Type:     types.AttackTypeTamperedMessage,
			Code:     types.MessageCode(i % 10),
			Sequence: uint64(i),
			Round:    uint64(i % 5),
			Enabled:  true,
		}
		attack := NewMockAttack(config)
		manager.RegisterAttack(attack)
	}

	// Generate UIDs for lookup
	uids := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		uids[i] = manager.GetUIDGenerator().Generate(
			types.AttackTypeTamperedMessage,
			types.MessageCode(i%10),
			uint64(i),
			uint64(i%5),
		)
	}

	// Benchmark lookup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := uids[i%1000]
		_, _ = manager.GetAttackByUID(uid)
	}
}

// BenchmarkConditionLookup benchmarks condition-based lookup
func BenchmarkConditionLookup(b *testing.B) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register attacks
	for i := 0; i < 1000; i++ {
		config := types.AttackConfig{
			Type:     types.AttackType(fmt.Sprintf("type%d", i%5)),
			Code:     types.MessageCode(i % 10),
			Sequence: uint64(i % 100),
			Round:    uint64(i % 5),
			Enabled:  true,
		}
		attack := NewMockAttack(config)
		manager.RegisterAttack(attack)
	}

	// Benchmark condition-based lookup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetAttacksByCondition(
			types.AttackType(fmt.Sprintf("type%d", i%5)),
			types.MessageCode(i%10),
			uint64(i%100),
			uint64(i%5),
		)
	}
}

// TestPatternIndexing tests wildcard pattern indexing
func TestPatternIndexing(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register attack
	config := types.AttackConfig{
		Name:     "test-attack",
		Type:     types.AttackTypeTamperedMessage,
		Code:     types.MessageCodePrePrepare,
		Sequence: 100,
		Round:    0,
		Enabled:  true,
	}
	attack := NewMockAttack(config)
	err := manager.RegisterAttack(attack)
	require.NoError(t, err)

	// Verify pattern index was created
	// This requires exposing pattern index or adding a test method
	// For now, we test indirectly through lookups

	// Should find exact match
	found := manager.GetAttacksByCondition(
		types.AttackTypeTamperedMessage,
		types.MessageCodePrePrepare,
		100,
		0,
	)
	assert.Len(t, found, 1)

	// Test unregistration cleans up patterns
	uid := attack.GetConfig().UID
	err = manager.UnregisterAttack(uid)
	assert.NoError(t, err)

	// Should not find after unregistration
	found = manager.GetAttacksByCondition(
		types.AttackTypeTamperedMessage,
		types.MessageCodePrePrepare,
		100,
		0,
	)
	assert.Len(t, found, 0)
}

// TestEvaluateAndExecuteAttacks tests attack evaluation with O(1) lookup
func TestEvaluateAndExecuteAttacks(t *testing.T) {
	attackRegistry := registry.NewAttackRegistry()
	manager := NewAttackManager(attackRegistry, nil)

	// Register a blocking attack
	config := types.AttackConfig{
		Name:     "blocker",
		Type:     types.AttackTypeSilentMessage,
		Code:     types.MessageCodePrePrepare,
		Sequence: 100,
		Round:    0,
		Enabled:  true,
	}
	attack := NewMockAttack(config)
	attack.shouldBlock = true
	err := manager.RegisterAttack(attack)
	require.NoError(t, err)

	// Create matching event
	event := types.Event{
		Type:     types.EventTypeMessageSent,
		Sequence: 100,
		Round:    0,
		Data: &types.MessageEvent{
			MessageCode: types.MessageCodePrePrepare,
			From:        common.HexToAddress("0x1234"),
		},
		Metadata: map[string]interface{}{
			"direction": types.DirectionSend,
		},
	}

	// Evaluate attacks
	ctx := context.Background()
	decision, err := manager.EvaluateAndExecuteAttacks(ctx, event)
	require.NoError(t, err)

	// Should decide to block
	assert.True(t, decision.ShouldAttack)
	assert.Equal(t, types.AttackTypeSilentMessage, decision.AttackType)
	assert.NotEmpty(t, decision.Reason)
	assert.NotNil(t, decision.Result)
}
