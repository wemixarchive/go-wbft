package event

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock QBFT event source
type MockQBFTEventSource struct {
	mock.Mock
}

func (m *MockQBFTEventSource) Subscribe(ch chan<- WBFTEvent) error {
	args := m.Called(ch)
	return args.Error(0)
}

func (m *MockQBFTEventSource) Unsubscribe(ch chan<- WBFTEvent) error {
	args := m.Called(ch)
	return args.Error(0)
}

// Test: Collector should start data collection
func TestEventCollector_StartCollection(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	dataReqs := []DataRequirement{
		{
			Type: "messages",
			Filter: DataFilter{
				FromSequence: 90,
				ToSequence:   100,
				MessageTypes: []uint64{uint64(types.MessageCodePrepare), uint64(types.MessageCodeCommit)},
			},
			MaxRecords: 100,
		},
	}

	// Mock expectations
	mockSource.On("Subscribe", mock.Anything).Return(nil)

	// When
	err := collector.Start(dataReqs)

	// Then
	assert.NoError(t, err)
	mockSource.AssertExpectations(t)
}

// Test: Collector should stop data collection
func TestEventCollector_StopCollection(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	// Start collection first
	dataReqs := []DataRequirement{
		{
			Type:       "messages",
			Filter:     DataFilter{},
			MaxRecords: 100,
		},
	}

	mockSource.On("Subscribe", mock.Anything).Return(nil)
	mockSource.On("Unsubscribe", mock.Anything).Return(nil)

	collector.Start(dataReqs)

	// When
	err := collector.Stop("attacks-123")

	// Then
	assert.NoError(t, err)
	mockSource.AssertExpectations(t)
}

// Test: Collector should collect QBFT messages
func TestEventCollector_CollectMessages(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	// Setup channel for events
	var eventChannel chan<- WBFTEvent
	mockSource.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		eventChannel = args.Get(0).(chan<- WBFTEvent)
	}).Return(nil)

	// Start collection
	dataReqs := []DataRequirement{
		{
			Type: "messages",
			Filter: DataFilter{
				FromSequence: 95,
				ToSequence:   105,
				MessageTypes: []uint64{uint64(types.MessageCodePrepare)},
			},
			MaxRecords: 10,
		},
	}

	err := collector.Start(dataReqs)
	assert.NoError(t, err)

	// When - Send events
	events := []WBFTEvent{
		{
			Type:     EventTypeMessage,
			Sequence: 95,
			Round:    0,
			Message: &WBFTMessage{
				Code:     uint64(types.MessageCodePrepare),
				Sequence: 95,
				Round:    0,
				Address:  common.HexToAddress("0x1234"),
			},
		},
		{
			Type:     EventTypeMessage,
			Sequence: 100,
			Round:    0,
			Message: &WBFTMessage{
				Code:     uint64(types.MessageCodePrepare),
				Sequence: 100,
				Round:    0,
				Address:  common.HexToAddress("0x5678"),
			},
		},
		{
			Type:     EventTypeMessage,
			Sequence: 110, // Outside filter range
			Round:    0,
			Message: &WBFTMessage{
				Code:     uint64(types.MessageCodePrepare),
				Sequence: 110,
				Round:    0,
			},
		},
	}

	for _, event := range events {
		eventChannel <- event
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Then - Query collected data
	filter := DataFilter{
		FromSequence: 95,
		ToSequence:   100,
		MessageTypes: []uint64{uint64(types.MessageCodePrepare)},
	}

	collected := collector.GetHistoricalData(filter)
	assert.Len(t, collected, 2)

	// Verify collected messages
	assert.Equal(t, uint64(95), collected[0].Sequence)
	assert.Equal(t, uint64(100), collected[1].Sequence)
}

// Test: Collector should track state changes
func TestEventCollector_TrackStateChanges(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	// Setup for state tracking
	var eventChannel chan<- WBFTEvent
	mockSource.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		eventChannel = args.Get(0).(chan<- WBFTEvent)
	}).Return(nil)

	dataReqs := []DataRequirement{
		{
			Type:       "state",
			Filter:     DataFilter{},
			MaxRecords: 50,
		},
	}

	collector.Start(dataReqs)

	// When - Send state change events
	stateEvents := []WBFTEvent{
		{
			Type:     EventTypeStateChange,
			Sequence: 100,
			Round:    0,
			OldState: types.StateIdle,
			NewState: types.StatePreprepared,
		},
		{
			Type:     EventTypeStateChange,
			Sequence: 100,
			Round:    0,
			OldState: types.StatePreprepared,
			NewState: types.StatePrepared,
		},
		{
			Type:     EventTypeStateChange,
			Sequence: 100,
			Round:    0,
			OldState: types.StatePrepared,
			NewState: types.StateCommitted,
		},
	}

	for _, event := range stateEvents {
		eventChannel <- event
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Then - Query state history
	states := collector.GetStateHistory(100)
	assert.Len(t, states, 3)

	// Verify state transitions
	assert.Equal(t, types.StateIdle, states[0].OldState)
	assert.Equal(t, types.StatePreprepared, states[0].NewState)
	assert.Equal(t, types.StateCommitted, states[2].NewState)
}

// Test: Collector should respect max records limit
func TestEventCollector_MaxRecordsLimit(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	var eventChannel chan<- WBFTEvent
	mockSource.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		eventChannel = args.Get(0).(chan<- WBFTEvent)
	}).Return(nil)

	// Set low max records limit
	dataReqs := []DataRequirement{
		{
			Type: "messages",
			Filter: DataFilter{
				MessageTypes: []uint64{uint64(types.MessageCodePrepare)},
			},
			MaxRecords: 3,
		},
	}

	collector.Start(dataReqs)

	// When - Send more events than limit
	for i := 0; i < 10; i++ {
		eventChannel <- WBFTEvent{
			Type:     EventTypeMessage,
			Sequence: uint64(100 + i),
			Round:    0,
			Message: &WBFTMessage{
				Code:     uint64(types.MessageCodePrepare),
				Sequence: uint64(100 + i),
			},
		}
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	// Then - Should only keep latest records up to limit
	filter := DataFilter{
		MessageTypes: []uint64{uint64(types.MessageCodePrepare)},
	}

	collected := collector.GetHistoricalData(filter)
	assert.Len(t, collected, 3)

	// Should have the latest 3 messages
	assert.Equal(t, uint64(107), collected[0].Sequence)
	assert.Equal(t, uint64(108), collected[1].Sequence)
	assert.Equal(t, uint64(109), collected[2].Sequence)
}

// Test: Collector should handle concurrent access
func TestEventCollector_ConcurrentAccess(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	var eventChannel chan<- WBFTEvent
	mockSource.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		eventChannel = args.Get(0).(chan<- WBFTEvent)
	}).Return(nil)

	dataReqs := []DataRequirement{
		{
			Type:       "messages",
			Filter:     DataFilter{},
			MaxRecords: 100,
		},
	}

	collector.Start(dataReqs)

	// When - Concurrent operations
	done := make(chan bool)

	// Goroutine 1: Send events
	go func() {
		for i := 0; i < 50; i++ {
			eventChannel <- WBFTEvent{
				Type:     EventTypeMessage,
				Sequence: uint64(100 + i),
				Message: &WBFTMessage{
					Code:     uint64(types.MessageCodePrepare),
					Sequence: uint64(100 + i),
				},
			}
		}
		done <- true
	}()

	// Goroutine 2: Query data
	go func() {
		for i := 0; i < 20; i++ {
			collector.GetHistoricalData(DataFilter{})
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: More queries
	go func() {
		for i := 0; i < 20; i++ {
			collector.GetStateHistory(uint64(100 + i))
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Then - No panic, data is collected
	data := collector.GetHistoricalData(DataFilter{})
	assert.NotNil(t, data)
	assert.True(t, len(data) > 0)
}

// Test: Collector should filter data correctly
func TestEventCollector_DataFiltering(t *testing.T) {
	// Given
	mockSource := new(MockQBFTEventSource)
	collector := NewEventCollectorWithEvent(mockSource)

	var eventChannel chan<- WBFTEvent
	mockSource.On("Subscribe", mock.Anything).Run(func(args mock.Arguments) {
		eventChannel = args.Get(0).(chan<- WBFTEvent)
	}).Return(nil)

	collector.Start([]DataRequirement{{
		Type:       "messages",
		Filter:     DataFilter{},
		MaxRecords: 100,
	}})

	// Send various messages
	validators := []common.Address{
		common.HexToAddress("0x1111"),
		common.HexToAddress("0x2222"),
		common.HexToAddress("0x3333"),
	}

	for seq := uint64(95); seq <= 105; seq++ {
		for _, val := range validators {
			for _, msgType := range []MessageCode{types.MessageCodePrepare, types.MessageCodeCommit} {
				eventChannel <- WBFTEvent{
					Type:     EventTypeMessage,
					Sequence: seq,
					Message: &WBFTMessage{
						Code:     uint64(msgType),
						Sequence: seq,
						Address:  val,
					},
				}
			}
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Test various filters
	testCases := []struct {
		name     string
		filter   DataFilter
		expected int
	}{
		{
			name: "Filter by sequence range",
			filter: DataFilter{
				FromSequence: 98,
				ToSequence:   102,
			},
			expected: 30, // 5 sequences * 3 validators * 2 message types
		},
		{
			name: "Filter by message type",
			filter: DataFilter{
				MessageTypes: []uint64{uint64(types.MessageCodePrepare)},
			},
			expected: 33, // 11 sequences * 3 validators
		},
		{
			name: "Filter by validator",
			filter: DataFilter{
				Validators: []common.Address{validators[0]},
			},
			expected: 22, // 11 sequences * 2 message types
		},
		{
			name: "Combined filter",
			filter: DataFilter{
				FromSequence: 100,
				ToSequence:   100,
				MessageTypes: []uint64{uint64(types.MessageCodeCommit)},
				Validators:   []common.Address{validators[1]},
			},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := collector.GetHistoricalData(tc.filter)
			assert.Len(t, data, tc.expected)
		})
	}
}
