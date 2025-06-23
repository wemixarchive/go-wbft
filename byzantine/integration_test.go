package byzantine

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ethereum/go-ethereum/byzantine/api"
	"github.com/ethereum/go-ethereum/byzantine/attack"
	"github.com/ethereum/go-ethereum/byzantine/event"
)

// Mock WBFTEventSource for integration testing
type MockQBFTEventSource struct {
	mock.Mock
	subscribers []chan<- event.WBFTEvent
	mu          sync.Mutex
}

func (m *MockQBFTEventSource) Subscribe(ch chan<- event.WBFTEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ch)
	m.subscribers = append(m.subscribers, ch)
	return args.Error(0)
}

func (m *MockQBFTEventSource) Unsubscribe(ch chan<- event.WBFTEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	args := m.Called(ch)

	// Remove subscriber
	for i, sub := range m.subscribers {
		if sub == ch {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			break
		}
	}

	return args.Error(0)
}

func (m *MockQBFTEventSource) BroadcastEvent(event event.WBFTEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, ch := range m.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

// Test: Full attack configuration flow
func TestIntegration_AttackConfigurationFlow(t *testing.T) {
	// Given - Setup all components
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(nil)
	mockEventSource.On("Unsubscribe", mock.Anything).Return(nil)

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// When - Configure a double prepare attack
	params := api.AttackParams{
		Type:     api.AttackTypeDoublePrepare,
		Sequence: 100,
		Round:    0,
		Target: []common.Address{
			common.HexToAddress("0x1234567890123456789012345678901234567890"),
		},
		Options: map[string]interface{}{
			"delay":            uint64(1000),
			"withValidMessage": true,
		},
	}

	attackID, err := byzantineAPI.ConfigureAttack(params)

	// Then - Attack should be configured successfully
	assert.NoError(t, err)
	assert.NotEmpty(t, attackID)

	// Verify attack is registered in manager
	registeredAttack, exists := attackManager.GetAttack(attackID)
	assert.True(t, exists)
	assert.NotNil(t, registeredAttack)

	// Verify attack properties
	assert.True(t, registeredAttack.ShouldExecute(100, 0, uint64(attack.MessageCodePrepare)))
	assert.False(t, registeredAttack.ShouldExecute(99, 0, uint64(attack.MessageCodePrepare)))
}

// Test: Attack execution based on QBFT events
func TestIntegration_AttackExecutionFlow(t *testing.T) {
	// Given - Setup components
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(nil)

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// Configure multiple attacks
	attacks := []api.AttackParams{
		{
			Type:     api.AttackTypeDoublePrepare,
			Sequence: 100,
			Round:    0,
		},
		{
			Type:     api.AttackTypeDoubleCommit,
			Sequence: 100,
			Round:    0,
		},
		{
			Type:     api.AttackTypeSilentProposer,
			Sequence: 101,
			Round:    0,
		},
	}

	var attackIDs []string
	for _, params := range attacks {
		id, err := byzantineAPI.ConfigureAttack(params)
		assert.NoError(t, err)
		attackIDs = append(attackIDs, id)
	}

	// When - Check which attacks should execute for different conditions

	// Sequence 100, Round 0, Prepare message
	prepareAttacks := attackManager.GetAttacksToExecute(100, 0, uint64(attack.MessageCodePrepare))
	assert.Len(t, prepareAttacks, 1)
	assert.Equal(t, attackIDs[0], prepareAttacks[0].ID())

	// Sequence 100, Round 0, Commit message
	commitAttacks := attackManager.GetAttacksToExecute(100, 0, uint64(attack.MessageCodeCommit))
	assert.Len(t, commitAttacks, 1)
	assert.Equal(t, attackIDs[1], commitAttacks[0].ID())

	// Sequence 101, Round 0, PrePrepare message
	preprepareAttacks := attackManager.GetAttacksToExecute(101, 0, uint64(attack.MessageCodePrePrepare))
	assert.Len(t, preprepareAttacks, 1)
	assert.Equal(t, attackIDs[2], preprepareAttacks[0].ID())
}

// Test: Event collection and data availability
func TestIntegration_EventCollectionFlow(t *testing.T) {
	// Given - Setup components
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(nil)

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// Configure an attack that requires historical data
	params := api.AttackParams{
		Type:     api.AttackTypeTamperedHeader,
		Sequence: 100,
		Round:    0,
		Options: map[string]interface{}{
			"tamperFields": []attack.TamperField{
				{
					Target: "Proposal.Header.Coinbase",
					Value:  common.HexToAddress("0xdeadbeef"),
				},
			},
		},
	}

	_, err := byzantineAPI.ConfigureAttack(params)
	assert.NoError(t, err)

	// Simulate QBFT events
	events := []event.WBFTEvent{
		{
			Type:     event.EventTypeMessage,
			Sequence: 95,
			Round:    0,
			Message: &event.WBFTMessage{
				Code:     uint64(attack.MessageCodePrepare),
				Sequence: 95,
				Address:  common.HexToAddress("0x1111"),
			},
		},
		{
			Type:     event.EventTypeMessage,
			Sequence: 98,
			Round:    0,
			Message: &event.WBFTMessage{
				Code:     uint64(attack.MessageCodeCommit),
				Sequence: 98,
				Address:  common.HexToAddress("0x2222"),
			},
		},
		{
			Type:     event.EventTypeStateChange,
			Sequence: 99,
			Round:    0,
			OldState: event.StateIdle,
			NewState: event.StatePreprepared,
		},
	}

	// Broadcast events
	for _, event := range events {
		mockEventSource.BroadcastEvent(event)
	}

	// Give time for event processing
	time.Sleep(100 * time.Millisecond)

	// When - Query collected data
	filter := attack.DataFilter{
		FromSequence: 90,
		ToSequence:   99,
	}

	historicalData := eventCollector.GetHistoricalData(filter)

	// Then - Data should be available
	assert.Len(t, historicalData, 2) // 2 messages

	// Check state history
	stateHistory := eventCollector.GetStateHistory(99)
	assert.Len(t, stateHistory, 1)
	assert.Equal(t, event.StatePreprepared, stateHistory[0].NewState)
}

// Test: Attack lifecycle management
func TestIntegration_AttackLifecycle(t *testing.T) {
	// Given - Setup components
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(nil)
	mockEventSource.On("Unsubscribe", mock.Anything).Return(nil)

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// When - Configure attack
	params := api.AttackParams{
		Type:     api.AttackTypeDoublePrepare,
		Sequence: 100,
		Round:    0,
	}

	attackID, err := byzantineAPI.ConfigureAttack(params)
	assert.NoError(t, err)

	// Verify attack is active
	attacks := byzantineAPI.ListAttacks()
	assert.Len(t, attacks, 1)
	assert.Equal(t, attack.AttackStatusActive, attacks[0].Status)

	// When - Execute attack (mark as executed)
	attackManager.MarkExecuted(attackID)

	// Then - Status should be updated
	status := attackManager.GetAttackStatus(attackID)
	assert.Equal(t, attack.AttackStatusExecuted, status)

	// When - Stop attack
	err = byzantineAPI.StopAttack(attackID)
	assert.NoError(t, err)

	// Then - Attack should be removed
	attacks = byzantineAPI.ListAttacks()
	assert.Len(t, attacks, 0)
}

// Test: Multiple concurrent attacks
func TestIntegration_ConcurrentAttacks(t *testing.T) {
	// Given - Setup components
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(nil)

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// When - Configure multiple attacks concurrently
	done := make(chan bool)
	attackCount := 10

	for i := 0; i < attackCount; i++ {
		go func(seq uint64) {
			params := api.AttackParams{
				Type:     api.AttackTypeDoublePrepare,
				Sequence: seq,
				Round:    0,
			}

			_, err := byzantineAPI.ConfigureAttack(params)
			assert.NoError(t, err)
			done <- true
		}(uint64(100 + i))
	}

	// Wait for all configurations
	for i := 0; i < attackCount; i++ {
		<-done
	}

	// Then - All attacks should be registered
	attacks := byzantineAPI.ListAttacks()
	assert.Len(t, attacks, attackCount)

	// Verify each can be found at its sequence
	for i := 0; i < attackCount; i++ {
		seq := uint64(100 + i)
		seqAttacks := attackManager.GetAttacksToExecute(seq, 0, uint64(attack.MessageCodePrepare))
		assert.Len(t, seqAttacks, 1)
	}
}

// Test: Error handling in the flow
func TestIntegration_ErrorHandling(t *testing.T) {
	// Given - Setup with failing event source
	mockEventSource := new(MockQBFTEventSource)
	mockEventSource.On("Subscribe", mock.Anything).Return(errors.New("subscription failed"))

	eventCollector := event.NewEventCollector(mockEventSource)
	attackManager := attack.NewAttackManager(eventCollector)
	attackBuilder := attack.NewAttackBuilder()
	byzantineAPI := api.NewByzantineAPIWithManager(attackBuilder, attackManager)

	// When - Try to configure attack with data requirements
	params := api.AttackParams{
		Type:     api.AttackTypeTamperedHeader,
		Sequence: 100,
		Round:    0,
	}

	attackID, err := byzantineAPI.ConfigureAttack(params)

	// Then - Should handle error gracefully
	assert.Error(t, err)
	assert.Empty(t, attackID)

	// Attack should not be registered
	attacks := byzantineAPI.ListAttacks()
	assert.Len(t, attacks, 0)
}
